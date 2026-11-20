package payments

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	repo "github.com/rahul810050/go-store/internal/adapters/postgresql/sqlc"
	"github.com/razorpay/razorpay-go"
)


type PaymentResponse struct {
    PaymentID      int64  `json:"payment_id"`
    OrderID        int64  `json:"order_id"`
    Amount         int64  `json:"amount"`
    Currency       string `json:"currency"`
    RazorpayKeyID  string `json:"razorpay_key_id"`
    RazorpayOrderID string `json:"razorpay_order_id"`
}

type Service interface {
	CreatePayment(ctx context.Context, userID int64, orderID int64) (PaymentResponse, error)
}

type service struct {
	repo repo.Querier
	razorpayClient *razorpay.Client
	razorpayKeyID string
}


func NewService(repo repo.Querier, razorpayClient *razorpay.Client, razorpayKeyID string) Service {
	return &service{
		repo: repo,
		razorpayClient: razorpayClient,
		razorpayKeyID: razorpayKeyID,
	}
}


func (s *service) CreatePayment(ctx context.Context, userID int64, orderID int64) (PaymentResponse, error) {
	if userID <= 0 || orderID <= 0 {
		return PaymentResponse{}, ErrInvalidPayment
	}

	// check that the order belongs to this user
	order, err := s.repo.GetOrderByID(ctx, repo.GetOrderByIDParams{
		ID: orderID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return PaymentResponse{}, ErrOrderNotFound
		}
		return PaymentResponse{}, err
	}

	// only payable order can be paid
	if order.Status != "pending" {
		return PaymentResponse{}, ErrOrderNotPayable
	}

	if order.TotalAmount <= 0 {
		return PaymentResponse{}, ErrInvalidPayment
	}

	// create payment record in db
	payment, err := s.repo.CreatePayment(ctx, repo.CreatePaymentParams{
		OrderID: order.ID,
		Amount: order.TotalAmount,
	})
	if err != nil {
		return PaymentResponse{}, err
	}

	data := map[string]interface{}{
		"amount": order.TotalAmount,
		"currency": "INR",
		"receipt": fmt.Sprintf("payment_%d", payment.ID),
	}

	razorpayOrder, err := s.razorpayClient.Order.Create(data, nil)
	if err != nil {
		return PaymentResponse{}, fmt.Errorf("failed to create Razorpay order %w", err)
	}

	// extract razorpay order id
	providerOrderID, ok := razorpayOrder["id"].(string)
	if !ok || providerOrderID == "" {
		return PaymentResponse{}, errors.New("invalid Razorpay order response")
	}

	// save Razorpay's order id against our payment record
	payment, err = s.repo.UpdatePaymentProviderOrderID(ctx, repo.UpdatePaymentProviderOrderIDParams{
		ID: payment.ID,
		ProviderOrderID: pgtype.Text{
			String: providerOrderID,
			Valid: true,
		},
	})

	if err != nil {
		return PaymentResponse{}, fmt.Errorf("failed to update payment: %w", err)
	}

	return PaymentResponse{
		PaymentID:       payment.ID,
		OrderID:         payment.OrderID,
		Amount:          payment.Amount,
		Currency:        payment.Currency,
		RazorpayKeyID:   s.razorpayKeyID,
		RazorpayOrderID: providerOrderID,
	}, nil
}
package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
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
	VerifyPayment(ctx context.Context, userID int64, input VerifyPaymentRequest) error
}

func verifyRazorpaySignature(orderID, paymentID, signature, secret string) bool {
	// construct the message
	// razorpay's signature formula requires the orderID followed by a pipe (|) and the paymentID
	message := orderID + "|" + paymentID
	
	// generate a cryptographic signature using private Razorpay Key Secret
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	
	// convert the signature into hexadecimal
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// compare signature
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

type service struct {
	repo repo.Querier
	db *pgxpool.Pool
	razorpayClient *razorpay.Client
	razorpayKeyID string
	razorpayKeySecret string
}


func NewService(db *pgxpool.Pool, razorpayClient *razorpay.Client, razorpayKeyID string, razorpayKeySecret string) Service {
	return &service{
		repo: repo.New(db),
		db: db,
		razorpayClient: razorpayClient,
		razorpayKeyID: razorpayKeyID,
		razorpayKeySecret: razorpayKeySecret,
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


func (s *service) VerifyPayment(ctx context.Context, userID int64, input VerifyPaymentRequest) error {
	if userID <= 0 ||
		input.RazorpayOrderID == "" ||
		input.RazorpayPaymentID == "" ||
		input.RazorpaySignature == "" {
		return ErrInvalidPayment
	}

	// get payment
	payment, err := s.repo.GetPaymentByProviderOrderID(ctx, pgtype.Text{
		String: input.RazorpayOrderID,
		Valid: true,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPaymentNotFound
		}
		return err
	}

	// get order
	_, err = s.repo.GetOrderByID(ctx, repo.GetOrderByIDParams{
		ID: payment.OrderID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPaymentNotFound
		}
		return err
	}

	if payment.Status == "success" {
		return nil
	}

	if payment.Status != "pending" {
		return ErrInvalidPayment
	}

	// verify signature
	isSignatureValid := verifyRazorpaySignature(
		payment.ProviderOrderID.String,
		input.RazorpayPaymentID,
		input.RazorpaySignature,
		s.razorpayKeySecret,
	)
	if !isSignatureValid {
		return ErrInvalidSignature
	}

	// retrieve payment details from Razorpay
	razorpayPayment, err := s.razorpayClient.Payment.Fetch(
		input.RazorpayPaymentID,
		nil,
		nil,
	)
	if err != nil {
		return err
	}

	razorpayOrderID, orderOK := razorpayPayment["order_id"].(string)
	razorpayPaymentID, paymentOK := razorpayPayment["id"].(string)
	razorpayAmount, amountOK := razorpayPayment["amount"].(float64)
	razorpayCurrency, currencyOK := razorpayPayment["currency"].(string)
	razorpayStatus, statusOK := razorpayPayment["status"].(string)
	captured, capturedOK := razorpayPayment["captured"].(bool)

	if !orderOK || !paymentOK || !amountOK ||
		!currencyOK || !statusOK || !capturedOK {
		return ErrPaymentMismatch
	}

	if razorpayOrderID != payment.ProviderOrderID.String ||
		razorpayPaymentID != input.RazorpayPaymentID ||
		int64(razorpayAmount) != payment.Amount ||
		razorpayCurrency != payment.Currency {
		return ErrPaymentMismatch
	}

	if razorpayStatus != "captured" || !captured {
		return ErrPaymentNotCaptured
	}

	// create transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// to make sure our updates belong to same transaction
	qtx := repo.New(tx)

	// lock payment row for concurrency
	lockedPayment, err := qtx.GetPaymentByIDForUpdate(ctx, payment.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrPaymentNotFound
		}
	}

	// recheck payment status after locking the row
	if lockedPayment.Status == "success" {
		return nil
	}

	if lockedPayment.Status != "pending" {
		return ErrInvalidPayment
	}

	// update payment status
	_, err = qtx.UpdatePaymentStatus(ctx, repo.UpdatePaymentStatusParams{
		ID: payment.ID,
		Status: "success",
		ProviderPaymentID: pgtype.Text{
			String: input.RazorpayPaymentID,
			Valid: true,
		},
	})
	if err != nil {
		return err
	}

	// update order status
	_, err = qtx.UpdateOrderStatus(ctx, repo.UpdateOrderStatusParams{
		ID: payment.OrderID,
		Status: "confirmed",
	})
	if err != nil {
		return err
	}

	// commit transaction
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
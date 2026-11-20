package orders

import (
	"context"
	"errors"
	"math"
	"sort"

	"github.com/jackc/pgx/v5"
	repo "github.com/rahul810050/go-store/internal/adapters/postgresql/sqlc"
)

type CreateOrderInput struct {
	Items []OrderItemInput `json:"items"`
}

type OrderItemInput struct {
	ProductID int64 `json:"product_id"`
	Quantity  int32 `json:"quantity"`
}

type OrderResponse struct {
	Order repo.Order
	Items []repo.OrderItem
}

type Service interface {
	CreateOrder(ctx context.Context, userID int64, input CreateOrderInput) (repo.Order, error)
	ListOrders(ctx context.Context, userID int64) ([]OrderResponse, error) 
	GetOrderByID(ctx context.Context, userID int64, orderID int64) (OrderResponse, error)
}

type service struct {
	repo repo.Querier
	db   *pgx.Conn
}

func NewService(db *pgx.Conn) Service {
	return &service{
		repo: repo.New(db),
		db:   db,
	}
}

func validateOrderInput(input CreateOrderInput) error {
	if len(input.Items) == 0 {
		return ErrEmptyOrder
	}

	seen := make(map[int64]bool)

	for _, item := range input.Items {
		if item.ProductID <= 0 || item.Quantity <= 0 {
			return ErrInvalidOrderItem
		}
		if seen[item.ProductID] {
			return ErrDuplicateProduct
		}
		seen[item.ProductID] = true
	}

	return nil
}

func (s *service) CreateOrder(ctx context.Context, userID int64, input CreateOrderInput) (repo.Order, error) {
	// input check
	if userID <= 0 {
		return repo.Order{}, errors.New("invalid user ID")
	}

	if err := validateOrderInput(input); err != nil {
		return repo.Order{}, err
	}

	// transaction begin
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return repo.Order{}, err
	}
	defer tx.Rollback(ctx)
	qtx := repo.New(tx)

	type preparedItem struct {
		ProductID int64
		Quantity  int32
		UnitPrice int64
		Subtotal  int64
	}

	var preparedItems []preparedItem
	var totalAmount int64

	// sort productID to reduce deadlock risk
	items := append([]OrderItemInput(nil), input.Items...)
	sort.Slice(items, func(i, j int) bool {
		return items[i].ProductID < items[j].ProductID
	})

	// lock products, validate stock and calculate price
	for _, item := range items {
		product, err := qtx.GetProductForUpdate(ctx, item.ProductID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return repo.Order{}, ErrProductNotFound
			}
			return repo.Order{}, err
		}
		if product.Quantity < item.Quantity {
			return repo.Order{}, ErrInsufficientStock
		}

		unitPrice := product.Price
		quantity := int64(item.Quantity)
		// prevent multiplication overflow
		if unitPrice > math.MaxInt64/quantity {
			return repo.Order{}, errors.New("order amount overflow")
		}

		subtotal := unitPrice * quantity

		// prevent totalAmount overflow
		if totalAmount > math.MaxInt64-subtotal {
			return repo.Order{}, errors.New("order total overflow")
		}

		preparedItems = append(preparedItems, preparedItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: unitPrice,
			Subtotal:  subtotal,
		})
		totalAmount += subtotal
	}

	// create oders
	order, err := qtx.CreateOrder(ctx, repo.CreateOrderParams{
		UserID:      userID,
		TotalAmount: totalAmount,
	})
	if err != nil {
		return repo.Order{}, err
	}

	// decrease stocks and create orderItems
	for _, item := range preparedItems {
		_, err := qtx.DecreaseProductStock(ctx, repo.DecreaseProductStockParams{
			ID:       item.ProductID,
			Quantity: item.Quantity,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return repo.Order{}, ErrInsufficientStock
			}
			return repo.Order{}, err
		}

		_, err = qtx.CreateOrderItem(ctx, repo.CreateOrderItemParams{
			OrderID:   order.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
			Subtotal:  item.Subtotal,
		})
		if err != nil {
			return repo.Order{}, err
		}
	}

	// commit transaction
	if err := tx.Commit(ctx); err != nil {
		return repo.Order{}, err
	}

	return order, nil
}


func (s *service) ListOrders(ctx context.Context, userID int64) ([]OrderResponse, error) {
	if userID <= 0 {
		return nil, errors.New("Invalid user")
	}

	orders, err := s.repo.ListOrdersByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]OrderResponse, 0, len(orders))

	for _, order := range orders {
		items, err := s.repo.ListOrderItemsByOrderID(ctx, order.ID)
		if err != nil {
			return nil, err
		}

		result = append(result, OrderResponse{
			Order: order,
			Items: items,
		})
	}
	return result, nil
}

func (s *service) GetOrderByID(ctx context.Context, userID int64, orderID int64) (OrderResponse, error) {
	if orderID <= 0 || userID <= 0 {
		return OrderResponse{}, ErrProductNotFound
	}

	order, err := s.repo.GetOrderByID(ctx, repo.GetOrderByIDParams{
		ID: orderID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OrderResponse{}, ErrOrderNotFound
		}
		return OrderResponse{}, err
	}

	items, err := s.repo.ListOrderItemsByOrderID(ctx, order.ID)
	if err != nil {
		return OrderResponse{}, err
	}

	return OrderResponse{
		Order: order,
		Items: items,
	}, nil
}
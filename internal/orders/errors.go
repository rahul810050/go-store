package orders

import "errors"

var (
	ErrEmptyOrder       = errors.New("order must contain at least one item")
    ErrInvalidOrderItem = errors.New("invalid order item")
    ErrDuplicateProduct = errors.New("duplicate product in order")
    ErrProductNotFound  = errors.New("product not found")
    ErrInsufficientStock = errors.New("insufficient product stock")
	ErrOrderNotFound = errors.New("order not found")
)
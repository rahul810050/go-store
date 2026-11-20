package payments

import "errors"

var (
	ErrInvalidPayment = errors.New("invalid payment request")
    ErrOrderNotFound  = errors.New("order not found")
    ErrOrderNotPayable = errors.New("order is not eligible for payment")
)
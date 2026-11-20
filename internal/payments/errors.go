package payments

import "errors"

var (
	ErrInvalidPayment = errors.New("invalid payment request")
    ErrOrderNotFound  = errors.New("order not found")
    ErrOrderNotPayable = errors.New("order is not eligible for payment")
	ErrPaymentNotFound = errors.New("payment not found")
	ErrInvalidSignature = errors.New("invalid payment signature")
	ErrPaymentMismatch = errors.New("payment details do not match")
	ErrPaymentNotCaptured = errors.New("payment has not been captured")
)
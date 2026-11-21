package webhooks

import "errors"

var (
	ErrMissingSignature = errors.New("missing webhook signature")
	ErrInvalidSignature = errors.New("invalid webhook signature")
	ErrPayloadTooLarge = errors.New("webhook payload too large")
	ErrInvalidBody = errors.New("invalid webhook body")
)
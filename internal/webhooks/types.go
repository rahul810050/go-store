package webhooks


const (
	EventPaymentCaptured = "payment.captured"
	EventPaymentFailed  = "payment.failed"
)

type WebhookEvent struct {
	ID string `json:"id"`
	Event string `json:"event"`
	CreatedAt int64 `json:"created_at"`
	AccountID string `json:"account_id"`
	Payload WebhookPayload `json:"payload"`
}

type PaymentEntity struct {
	ID string `json:"id"`
	OrderID string `json:"order_id"`
	Amount int64 `json:"amount"`
	Currency string `json:"currency"`
	Status string `json:"status"`
	Captured bool `json:"captured"`
}

type PaymentPayload struct {
	Entity PaymentEntity `json:"entity"`
}

type WebhookPayload struct {
	Payment PaymentPayload `json:"payment"`
}
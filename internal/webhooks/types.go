package webhooks

type WebhookEvent struct {
	ID string `json:"id"`
	Event string `json:"event"`
	CreatedAt int64 `json:"created_at"`
	AccountID string `json:"account_id"`
}
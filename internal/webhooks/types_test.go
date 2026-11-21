package webhooks

import (
	"testing"
)

func TestWebhookEventDecoding(t *testing.T) {
	payload := []byte(`{
		"event": "payment.captured",
		"created_at": 1795264800,
		"account_id": "acc_test",
		"payload": {
			"payment": {
				"entity": {
					"id": "pay_test123",
					"order_id": "order_test123",
					"amount": 190000,
					"currency": "INR",
					"status": "captured",
					"captured": true
				}
			}
		}
	}`)

	event, err := decodeWebhookEvent(payload)
	if err != nil {
		t.Fatalf("failed to decode webhook: %v", err)
	}

	if event.Event != EventPaymentCaptured {
		t.Fatalf("unexpected event type: %s", event.Event)
	}

	payment := event.Payload.Payment.Entity

	if payment.ID != "pay_test123" {
		t.Fatalf("unexpected payment ID: %s", payment.ID)
	}

	if payment.OrderID != "order_test123" {
		t.Fatalf("unexpected order ID: %s", payment.OrderID)
	}

	if payment.Amount != 190000 {
		t.Fatalf("unexpected amount: %d", payment.Amount)
	}

	if !payment.Captured {
		t.Fatal("expected payment to be captured")
	}
}
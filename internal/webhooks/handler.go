package webhooks

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)


const maxWebhookBodySize = 1 << 20 // 1 MB

func readAndVerifyWebhook(w http.ResponseWriter, r *http.Request, secret string) ([]byte, error) {
	signature := r.Header.Get("X-Razorpay-Signature")
	if signature == "" {
		return nil, ErrMissingSignature
	}
	// limit incoming request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBodySize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return nil, ErrPayloadTooLarge
		}
		return nil, ErrInvalidBody
	}

	// verify the signature against the exact raw bytes
	if !verifyWebhookSignature(body, signature, secret) {
		return nil, ErrInvalidSignature
	}
	return body, nil
}

type handler struct {
	service Service
	secret string
}

func NewHandler(service Service, secret string) *handler {
	return &handler{
		service: service,
		secret: secret,
	}
}


func (h *handler) HandleRazorpayWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := readAndVerifyWebhook(w, r, h.secret)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidSignature),
			errors.Is(err, ErrMissingSignature):
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case errors.Is(err, ErrPayloadTooLarge):
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}

	eventID := r.Header.Get("X-Razorpay-Event-Id")
	if eventID == "" {
		http.Error(w, "Missing event ID", http.StatusBadRequest)
		return
	}

	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		http.Error(w, "Invalid webhook payload", http.StatusBadRequest)
		return 
	}

	// save the event in our db
	err = h.service.SaveWebhookEvent(r.Context(), eventID, event.Event, body)
	if err != nil {
		http.Error(w, "Failed to save webhook event", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
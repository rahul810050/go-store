package payments

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rahul810050/go-store/internal/jsonwriter"
	"github.com/rahul810050/go-store/internal/middleware"
)

type CreatePaymentRequest struct {
	OrderID int64 `json:"order_id"`
}

type handler struct {
	service Service
}


func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if input.OrderID <= 0 {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	// create payment
	payment, err := h.service.CreatePayment(r.Context(), userID, input.OrderID)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidPayment):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, ErrOrderNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, ErrOrderNotPayable):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "Failed to create payment", http.StatusInternalServerError)
		}
		return
	}

	jsonwriter.Write(w, http.StatusCreated, payment)
}
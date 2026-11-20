package orders

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rahul810050/go-store/internal/jsonwriter"
	"github.com/rahul810050/go-store/internal/middleware"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var input CreateOrderInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// call service
	order, err := h.service.CreateOrder(r.Context(), userID, input)
	if err != nil {
		switch err {
		case ErrEmptyOrder, ErrInvalidOrderItem, ErrDuplicateProduct:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case ErrProductNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case ErrInsufficientStock:
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "Failed to create order", http.StatusInternalServerError)
		}
		return
	}
	jsonwriter.Write(w, http.StatusOK, order)
}


func (h *handler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orders, err := h.service.ListOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}
	jsonwriter.Write(w, http.StatusOK, orders)
}


func (h *handler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	id := chi.URLParam(r, "id")
	orderID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid order id", http.StatusBadRequest)
		return
	}

	order, err := h.service.GetOrderByID(r.Context(), userID, orderID)
	if err != nil {
		switch err {
		case ErrOrderNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "Failed to fetch order", http.StatusInternalServerError)
		}
		return 
	}
	
	jsonwriter.Write(w, http.StatusOK, order)
}
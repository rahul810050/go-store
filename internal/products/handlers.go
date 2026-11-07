package products

import (
	"net/http"

	"github.com/rahul810050/go-store/internal/json"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	// 1. first call the service to get all the products
	// 2. return the data in JSON format to user
	products := []string{"hello", "world"}
	json.Write(w, http.StatusOK, products)
}

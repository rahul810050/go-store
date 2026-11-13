package products

import (
	"log"
	"net/http"

	"github.com/rahul810050/go-store/internal/json"
)

type handler struct {
	service Service
}

// Constructor
func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	// 1. first call the service to get all the products
	products, err := h.service.ListProducts(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 2. return the data in JSON format to user
	json.Write(w, http.StatusOK, products)
}

package routes

import (
	"github.com/go-chi/chi/v5"
	repo "github.com/rahul810050/go-store/internal/adapters/postgresql/sqlc"
	"github.com/rahul810050/go-store/internal/middleware"
	"github.com/rahul810050/go-store/internal/products"
)

func RegisterProductRoutes(r chi.Router, queries *repo.Queries) {
    productService := products.NewService(queries)
    productHandler := products.NewHandler(productService)

    r.Route("/products", func(r chi.Router) {
        r.Get("/", productHandler.ListProducts)
        r.Get("/{id}", productHandler.FindProductByID)
        r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Use(middleware.RequireAdmin)
			r.Post("/", productHandler.CreateProduct)
		})
    })
}
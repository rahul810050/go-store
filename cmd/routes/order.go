package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/rahul810050/go-store/internal/middleware"
	"github.com/rahul810050/go-store/internal/orders"
)


func RegisterOrderRoutes(r chi.Router, db *pgx.Conn){
	orderService := orders.NewService(db)
	orderHandler := orders.NewHandler(orderService)

	r.Route("/orders", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Post("/", orderHandler.CreateOrder)
			r.Get("/", orderHandler.ListOrders)
			r.Get("/{id}", orderHandler.GetOrderByID)
		})
	})
}
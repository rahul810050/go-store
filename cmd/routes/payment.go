package routes

import (
	"github.com/go-chi/chi/v5"
	repo "github.com/rahul810050/go-store/internal/adapters/postgresql/sqlc"
	"github.com/rahul810050/go-store/internal/middleware"
	"github.com/rahul810050/go-store/internal/payments"
	"github.com/razorpay/razorpay-go"
)

func RegisterPaymentRoutes(r chi.Router, queries repo.Querier, razorpayClient *razorpay.Client, razorpayKeyID string) {
	paymentService := payments.NewService(queries, razorpayClient, razorpayKeyID)
	paymentHandler := payments.NewHandler(paymentService)

	r.Route("/payments", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)

			r.Post("/", paymentHandler.CreatePayment)
		})
	})
}
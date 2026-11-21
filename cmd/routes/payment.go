package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rahul810050/go-store/internal/middleware"
	"github.com/rahul810050/go-store/internal/payments"
	"github.com/razorpay/razorpay-go"
)

func RegisterPaymentRoutes(r chi.Router, db *pgxpool.Pool, razorpayClient *razorpay.Client, razorpayKeyID string, razorpayKeySecret string) {
	paymentService := payments.NewService(db, razorpayClient, razorpayKeyID, razorpayKeySecret)
	paymentHandler := payments.NewHandler(paymentService)

	r.Route("/payments", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)

			r.Post("/", paymentHandler.CreatePayment)
			r.Post("/verify", paymentHandler.VerifyPayment)
		})
	})
}
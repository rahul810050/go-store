package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/rahul810050/go-store/internal/webhooks"
)


func RegisterWebhookRoutes(r chi.Router, db *pgx.Conn, razorpayWebhookSecret string) {
	webhookService := webhooks.NewService(db)
	webhookHandler := webhooks.NewHandler(webhookService, razorpayWebhookSecret)

	r.Route("/webhooks", func(r chi.Router) {
		r.Post("/razorpay", webhookHandler.HandleRazorpayWebhook)
	})
}
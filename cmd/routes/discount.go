package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rahul810050/go-store/internal/discounts"
	"github.com/rahul810050/go-store/internal/middleware"
)

func RegisterDiscountRoutes(r chi.Router, db *pgxpool.Pool) {
	discountService := discounts.NewService(db)
	discountHandler := discounts.NewHandler(discountService)

	r.Route("/discounts", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		// Read endpoints - authenticated users
		r.Get("/", discountHandler.ListCampaigns)
		r.Get("/{id}", discountHandler.GetCampaignByID)
		r.Get("/{id}/products", discountHandler.ListProductsByCampaign)

		// Admin-only endpoints
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAdmin)

			r.Post("/", discountHandler.CreateCampaign)
			r.Put("/{id}", discountHandler.UpdateCampaign)

			r.Patch("/{id}/activate", discountHandler.ActivateCampaign)
			r.Patch("/{id}/cancel", discountHandler.CancelCampaign)

			r.Post("/{id}/products/{productID}", discountHandler.AddProductToCampaign)
			r.Delete("/{id}/products/{productID}", discountHandler.RemoveProductFromCampaign)
		})
	})
}
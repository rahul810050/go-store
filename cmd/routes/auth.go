package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rahul810050/go-store/internal/auth"
)


func RegisterAuthRoutes(r chi.Router, db *pgxpool.Pool) {
	authService := auth.NewService(db)
	authHandler := auth.NewHandler(authService)

	r.Route("/auth", func(r chi.Router){
		r.Post("/signup", authHandler.Signup)
        r.Post("/signin", authHandler.Signin)
        r.Post("/refresh", authHandler.Refresh)
        r.Post("/logout", authHandler.Logout)
	})
}
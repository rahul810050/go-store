package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rahul810050/go-store/cmd/routes"
	repo "github.com/rahul810050/go-store/internal/adapters/postgresql/sqlc"
	"github.com/razorpay/razorpay-go"
)

// user -> GET /products -> service getProducts -> call db SELECT * FROM PRODUCTS

// mount -> we are just creating the router here
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi there! its all good"))
	})
	queries := repo.New(app.db)

    routes.RegisterAuthRoutes(r, app.db)
    routes.RegisterProductRoutes(r, queries)
	routes.RegisterOrderRoutes(r, app.db)
	routes.RegisterPaymentRoutes(r, app.db, app.razorpayClient, app.cfg.RazorpayKeyID, app.cfg.RazorpayKeySecret)
	routes.RegisterWebhookRoutes(r, app.db, app.cfg.RazorpayWebhookSecret)

	return r
}

// run
func (app *application) run(h http.Handler) error {
	srv := &http.Server{
		Addr: app.cfg.addr,
		Handler: h,
		ReadTimeout: time.Second * 30,
		WriteTimeout: time.Second * 10,
		IdleTimeout: time.Minute,
	}
	log.Printf("server is running on port: %s", app.cfg.addr)
	return srv.ListenAndServe()
}

type application struct {
	cfg Config
	db *pgxpool.Pool
	razorpayClient *razorpay.Client
}

type Config struct {
	addr string
	db dbConfig
	RazorpayKeyID string
	RazorpayKeySecret string
	RazorpayWebhookSecret string
}

type dbConfig struct {
	dbstr string
}
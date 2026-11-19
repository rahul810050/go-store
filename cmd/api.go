package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	repo "github.com/rahul810050/go-store/internal/adapters/postgresql/sqlc"
	"github.com/rahul810050/go-store/internal/auth"
	"github.com/rahul810050/go-store/internal/products"
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
	productService := products.NewService(repo.New(app.db))
	productHandler := products.NewHandler(productService)

	// routes
	r.Get("/products", productHandler.ListProducts)
	r.Get("/product/{id}", productHandler.FindProductByID)
	r.Post("/product", productHandler.CreateProduct)


	// auth
	authService := auth.NewService(app.db)
	authHandler := auth.NewHandler(authService)
	r.Post("/auth/signup", authHandler.Signup)
	r.Post("/auth/signin", authHandler.Signin)
	r.Post("/auth/refresh", authHandler.Refresh)
	r.Post("/auth/logout", authHandler.Logout)

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
	db *pgx.Conn
}

type Config struct {
	addr string
	db dbConfig
}

type dbConfig struct {
	dbstr string
}
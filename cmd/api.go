package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// user -> GET /products -> service getProducts -> call db SELECT * FROM PRODUCTS

// mount -> we are just creating the router here
func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID) // it assigns a unique ID to each request
	// r.Use(middleware.RealIP) // important for rate limiting, analytics and tracing

	// logs all the incoming HTTP reqs such as status code, method and URL
	r.Use(middleware.Logger)
	// recovers from panics in HTTP handlers and prevents the server from crashing because of an unhandled request panic
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi there! its all good")) // sending response that this is the root route
	})

	return r
}

// run
func (app *application) run(h http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("server is running on port: %s", app.config.addr)

	return srv.ListenAndServe()
}

type application struct {
	config config
}

type config struct {
	addr string // port
	db   dbConfig
}

type dbConfig struct {
	dsn string // the db string
}

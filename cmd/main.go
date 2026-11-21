package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rahul810050/go-store/internal/env"
	"github.com/rahul810050/go-store/internal/webhooks"
	"github.com/razorpay/razorpay-go"
)

func main(){
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()
	if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }
	cfg := Config{
		addr: ":8080",
		db: dbConfig{
			dbstr: env.GetString("DATABASE_URL", "host=localhost user=postgres password=random123 dbname=ecom sslmode=disable"),
		},
		RazorpayKeyID: env.GetString("RAZORPAY_KEY_ID", ""),
		RazorpayKeySecret: env.GetString("RAZORPAY_KEY_SECRET", ""),
		RazorpayWebhookSecret: env.GetString("RAZORPAY_WEBHOOK_SECRET", ""),
	}

	if cfg.RazorpayKeyID == "" || cfg.RazorpayKeySecret == "" {
		log.Fatal("Razorpay API credentials are missing")
	}
	if cfg.RazorpayWebhookSecret == "" {
		log.Fatal("Razorpay webhook secret is missing")
	}

	// logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	pool, err := pgxpool.New(ctx, cfg.db.dbstr)
	if err != nil {
		log.Fatal("failed to create database pool:", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	log.Println("database connected successfully")

	// creating Razorpay client
	razorpayClient := razorpay.NewClient(
		cfg.RazorpayKeyID, 
		cfg.RazorpayKeySecret,
	)

	api := application{
		cfg: cfg,
		db: pool,
		razorpayClient: razorpayClient,
	}
	// start workers
	webhookWorker := webhooks.NewWorker(pool)
	workerDone := make(chan struct{})

	go func() {
		defer close(workerDone)
		webhookWorker.Run(ctx, 5*time.Second)
	}()
	if err := api.run(ctx, api.mount()); err != nil {
		slog.Error("error while running the server", err)
	}

	stop()
	<-workerDone

	log.Println("application shutdown complete")
	
}
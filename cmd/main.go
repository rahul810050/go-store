package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/rahul810050/go-store/internal/env"
	"github.com/razorpay/razorpay-go"
)

func main(){
	ctx := context.Background()
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

	conn, err := pgx.Connect(ctx, cfg.db.dbstr)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	log.Println("database connected successfully")

	// creating Razorpay client
	razorpayClient := razorpay.NewClient(
		cfg.RazorpayKeyID, 
		cfg.RazorpayKeySecret,
	)

	api := application{
		cfg: cfg,
		db: conn,
		razorpayClient: razorpayClient,
	}
	if err := api.run(api.mount()); err != nil {
		slog.Error("error while running the server", err)
		os.Exit(1)
	}
	
}
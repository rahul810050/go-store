package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/rahul810050/go-store/internal/env"
)

func main() {
	ctx := context.Background()

	cfg := config{
		addr: ":8080",
		db:   dbConfig{
			dsn: env.GetString("GOOSE_DBSTRING", "host=localhost user=postgres password=random123 dbname=goecom sslmode=disable"),
		},
	}

	

	// logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// database connection
	conn, err := pgx.Connect(ctx, cfg.db.dsn)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	logger.Info("database connection successful", "dns", cfg.db.dsn)

	api := application{
		config: cfg,
		db: conn,
	}

	// start server
	if err := api.run(api.mount()); err != nil {
		slog.Error("error while starting the server error", err)
		os.Exit(1)
	}
}

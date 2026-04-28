package main

import (
	"context"
	"log"

	"github.com/neto-app/neto/api/internal/config"
	"github.com/neto-app/neto/api/internal/infrastructure/postgres"
	"github.com/neto-app/neto/api/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	srv := server.New(cfg, pool)
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

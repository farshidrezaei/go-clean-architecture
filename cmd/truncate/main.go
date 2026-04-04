package main

import (
	"context"
	"log"

	"clean_architecture/internal/infrastructure/config"
	databasepostgres "clean_architecture/internal/infrastructure/database/postgres"
)

func main() {
	_ = config.LoadDotEnv()
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	pool, err := databasepostgres.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	if err := databasepostgres.Truncate(context.Background(), pool); err != nil {
		log.Fatalf("truncate db: %v", err)
	}
}

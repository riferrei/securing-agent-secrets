// Command seed loads the demo dataset into Redis. It is idempotent.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/riferrei/securing-agent-secrets-1password/internal/config"
	"github.com/riferrei/securing-agent-secrets-1password/internal/redisstore"
	"github.com/riferrei/securing-agent-secrets-1password/internal/seed"
	"github.com/riferrei/securing-agent-secrets-1password/internal/vault"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	if token == "" {
		log.Fatal("OP_SERVICE_ACCOUNT_TOKEN is required")
	}
	redisURL, err := vault.ResolveRedisURL(ctx, token, cfg.RedisHost, cfg.RedisPort, cfg.RedisUser, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("1password: %v", err)
	}

	store, err := redisstore.New(ctx, redisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer store.Close()

	if err := seed.Apply(ctx, store); err != nil {
		log.Fatalf("seeding: %v", err)
	}

	fmt.Printf("Seeded %d customers into Redis.\n", len(seed.Customers))
}

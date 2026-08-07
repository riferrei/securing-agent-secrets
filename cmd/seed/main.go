// Command seed loads the demo dataset into Redis. It is idempotent.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/riferrei/securing-agent-secrets/internal/config"
	"github.com/riferrei/securing-agent-secrets/internal/redisstore"
	"github.com/riferrei/securing-agent-secrets/internal/seed"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	store, err := redisstore.New(ctx, cfg.RedisURL())
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer store.Close()

	if err := seed.Apply(ctx, store); err != nil {
		log.Fatalf("seeding: %v", err)
	}

	fmt.Printf("Seeded %d customers into Redis.\n", len(seed.Customers))
}

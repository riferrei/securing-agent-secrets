// Command seed loads the demo dataset into Valkey. It is idempotent.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/riferrei/securing-agent-secrets-1password/internal/config"
	"github.com/riferrei/securing-agent-secrets-1password/internal/seed"
	"github.com/riferrei/securing-agent-secrets-1password/internal/valkeystore"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	store, err := valkeystore.New(ctx, cfg.ValkeyHost, cfg.ValkeyPort, cfg.ValkeyUser, cfg.ValkeyPassword)
	if err != nil {
		log.Fatalf("valkey: %v", err)
	}
	defer store.Close()

	if err := seed.Apply(ctx, store); err != nil {
		log.Fatalf("seeding: %v", err)
	}

	fmt.Printf("Seeded %d customers into Valkey.\n", len(seed.Customers))
}

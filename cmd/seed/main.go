// Command seed loads the demo dataset into Valkey. It is idempotent.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/riferrei/securing-agent-secrets-1password/internal/config"
	"github.com/riferrei/securing-agent-secrets-1password/internal/seed"
	"github.com/riferrei/securing-agent-secrets-1password/internal/valkeystore"
	"github.com/riferrei/securing-agent-secrets-1password/internal/vault"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	if token == "" {
		log.Fatal("OP_SERVICE_ACCOUNT_TOKEN is required")
	}
	host, port, user, password, err := vault.ResolveValkey(ctx, token, cfg.ValkeyHost, cfg.ValkeyPort, cfg.ValkeyUser, cfg.ValkeyPassword)
	if err != nil {
		log.Fatalf("1password: %v", err)
	}

	store, err := valkeystore.New(ctx, host, port, user, password)
	if err != nil {
		log.Fatalf("valkey: %v", err)
	}
	defer store.Close()

	if err := seed.Apply(ctx, store); err != nil {
		log.Fatalf("seeding: %v", err)
	}

	fmt.Printf("Seeded %d customers into Valkey.\n", len(seed.Customers))
}

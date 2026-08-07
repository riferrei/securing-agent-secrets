// Command server wires the demo together and serves the REST API.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/riferrei/securing-agent-secrets/internal/agent"
	"github.com/riferrei/securing-agent-secrets/internal/config"
	"github.com/riferrei/securing-agent-secrets/internal/httpapi"
	"github.com/riferrei/securing-agent-secrets/internal/llm"
	"github.com/riferrei/securing-agent-secrets/internal/redisstore"
	"github.com/riferrei/securing-agent-secrets/internal/seed"
	"github.com/riferrei/securing-agent-secrets/internal/vault"
)

func main() {
	// -health probes the local endpoint for the Docker healthcheck (distroless
	// has no shell or wget).
	healthFlag := flag.Bool("health", false, "probe the local server's health endpoint and exit 0 if healthy")
	flag.Parse()

	cfg := config.Load()

	if *healthFlag {
		os.Exit(runHealthCheck(cfg.HTTPAddr))
	}

	ctx := context.Background()

	token := os.Getenv("OP_SERVICE_ACCOUNT_TOKEN")
	if token == "" {
		log.Fatal("OP_SERVICE_ACCOUNT_TOKEN is required")
	}
	redisURL, err := vault.ResolveRedisURL(ctx, token, cfg.RedisHost, cfg.RedisPort, cfg.RedisUser, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("1password: %v", err)
	}
	log.Printf("server: resolved Redis connection from 1Password")

	store, err := redisstore.New(ctx, redisURL)
	if err != nil {
		log.Fatalf("redis: %v", err)
	}
	defer store.Close()

	if ok, _ := store.Exists(ctx, "0001"); !ok {
		if err := seed.Apply(ctx, store); err != nil {
			log.Printf("server: auto-seed failed: %v", err)
		} else {
			log.Printf("server: auto-seeded demo data")
		}
	}

	llmClient := llm.New(cfg.OllamaBaseURL)
	ag := agent.New(llmClient, store, cfg.OllamaModel, cfg.MaxToolIterations)

	srv := httpapi.NewServer(cfg.HTTPAddr, ag, store)

	go func() {
		log.Printf("server: listening on %s (model=%s)", cfg.HTTPAddr, cfg.OllamaModel)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("server: shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server: shutdown error: %v", err)
	}
}

func runHealthCheck(addr string) int {
	url := "http://127.0.0.1" + addr + "/api/health"
	if !strings.HasPrefix(addr, ":") {
		url = "http://" + addr + "/api/health"
	}
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "healthcheck:", err)
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "healthcheck: status", resp.StatusCode)
		return 1
	}
	return 0
}

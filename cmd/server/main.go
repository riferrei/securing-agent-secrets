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

	"github.com/riferrei/securing-agent-secrets-1password/internal/agent"
	"github.com/riferrei/securing-agent-secrets-1password/internal/config"
	"github.com/riferrei/securing-agent-secrets-1password/internal/httpapi"
	"github.com/riferrei/securing-agent-secrets-1password/internal/llm"
	"github.com/riferrei/securing-agent-secrets-1password/internal/valkeystore"
	"github.com/riferrei/securing-agent-secrets-1password/internal/vault"
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
	host, port, user, password, err := vault.ResolveValkey(ctx, token, cfg.ValkeyHost, cfg.ValkeyPort, cfg.ValkeyUser, cfg.ValkeyPassword)
	if err != nil {
		log.Fatalf("1password: %v", err)
	}
	log.Printf("server: resolved Valkey connection from 1Password")

	store, err := valkeystore.New(ctx, host, port, user, password)
	if err != nil {
		log.Fatalf("valkey: %v", err)
	}
	defer store.Close()

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

// Package httpapi exposes the agent over GET /api/health, GET /api/ready, and
// POST /api/chat, with per-session conversation history.
package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/riferrei/securing-agent-secrets/internal/agent"
	"github.com/riferrei/securing-agent-secrets/internal/llm"
)

const maxHistory = 40

type Responder interface {
	Respond(ctx context.Context, history []llm.Message, question string) (agent.TurnResult, []llm.Message, error)
}

type Seeded interface {
	Exists(ctx context.Context, id string) (bool, error)
}

type chatRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
}

type chatResponse struct {
	UserMessage      string `json:"user_message"`
	AssistantMessage string `json:"assistant_message"`
	SessionID        string `json:"session_id"`
}

type readyResponse struct {
	Status string `json:"status"`
	Seeded bool   `json:"seeded"`
}

type errorResponse struct {
	Detail string `json:"detail"`
}

type sessions struct {
	mu sync.Mutex
	m  map[string][]llm.Message
}

func newSessions() *sessions { return &sessions{m: map[string][]llm.Message{}} }

func (s *sessions) get(id string) []llm.Message {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.m[id]
}

func (s *sessions) set(id string, history []llm.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[id] = trimHistory(history)
}

func trimHistory(h []llm.Message) []llm.Message {
	if len(h) <= maxHistory {
		return h
	}
	trimmed := make([]llm.Message, 0, maxHistory)
	trimmed = append(trimmed, h[0])
	trimmed = append(trimmed, h[len(h)-(maxHistory-1):]...)
	return trimmed
}

func newSessionID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func NewServer(addr string, responder Responder, store Seeded) *http.Server {
	sess := newSessions()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("GET /api/ready", handleReady(store))
	mux.HandleFunc("POST /api/chat", handleChat(responder, sess))

	return &http.Server{
		Addr:              addr,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleReady(store Seeded) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		seeded, err := store.Exists(r.Context(), "0001")
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, readyResponse{Status: "error", Seeded: false})
			return
		}
		status := "ok"
		if !seeded {
			status = "loading"
		}
		writeJSON(w, http.StatusOK, readyResponse{Status: status, Seeded: seeded})
	}
}

func handleChat(responder Responder, sess *sessions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Detail: "invalid JSON body"})
			return
		}
		msg := strings.TrimSpace(req.Message)
		if msg == "" {
			writeJSON(w, http.StatusBadRequest, errorResponse{Detail: "message is required"})
			return
		}

		sessionID := req.SessionID
		if sessionID == "" {
			sessionID = newSessionID()
		}

		history := sess.get(sessionID)
		result, updated, err := responder.Respond(r.Context(), history, msg)
		if err != nil {
			log.Printf("httpapi: chat failed: %v", err)
			writeJSON(w, http.StatusBadGateway, errorResponse{Detail: "agent turn failed"})
			return
		}
		sess.set(sessionID, updated)

		writeJSON(w, http.StatusOK, chatResponse{
			UserMessage:      result.UserMessage,
			AssistantMessage: result.AssistantMessage,
			SessionID:        sessionID,
		})
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

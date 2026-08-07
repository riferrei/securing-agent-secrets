# CLAUDE.md

Guidance for working in this repository. These rules are not incidental. They
are the security invariants this project is built around, so treat them as hard
constraints, not preferences.

## What this is

A small Go assistant for a customer support team. It answers business questions
about customers (tier, balance, account status, region), using a local model
(via Ollama) to interpret intent. Here every field, business
and PII (email, phone, SSN, address) alike, sits in one `customer:*` hash: the
naive baseline. Read and write both flow through the one Valkey connection, so the
agent's reach is broad, which is the blast radius a later branch scopes down. The
`controlling-blast-radius` branch refactors the entity into two hashes so a scoped
identity can be denied the PII. The secret being protected across the project is
the Valkey connection credential.

This is the `vaults-as-source-truth` branch. The connection no longer lives in
`.env`; its address, user, and password are resolved just in time from a
1Password vault via the Go SDK, authed by a service account. Nothing sensitive
touches disk or version control. The agent's access is still broad, which the
next branch scopes down.

## Architecture

```
UI (nginx)  ->  REST API (Go, :8000)  ->  Agent (Go)  ->  valkey-prod
                                             |
                                             +--> Ollama (local LLM)
```

The agent is naively wired to `valkey-prod` (the production database) with full
access. A later branch scopes it to a read-only, PII-blind Valkey identity.

- `frontend/` static UI served by nginx, which proxies `/api/` to the backend.
- `internal/httpapi` REST surface: `/api/health`, `/api/ready`, `/api/chat`.
- `internal/agent` the model-and-tools reasoning loop.
- `internal/valkeystore` typed access to customers, stored as a single `customer:*` hash (business fields and PII together).
- `internal/llm` minimal Ollama chat client with tool calling.
- `internal/config` loads configuration and the `op://` connection references.
- `internal/vault` resolves secrets from 1Password via the service account.
- `cmd/server` and `cmd/seed` entrypoints.

## Security invariants (do not break)

1. **Never log or print a resolved secret.** Not in server logs, not in an error
   message, not in an API response, not in the UI.
2. **The model interprets intent. The Go code owns the Valkey call.** The model
   picks a customer id through the `get_customer` tool. It never generates Valkey
   commands. This is a safety property (no model-authored destructive commands)
   and a correctness one.
3. **Keep the credential out of the model's context.** Do not put the connection
   string in the system prompt, tool descriptions, or any message sent to the
   model. It is only used by the Go code to open the Valkey connection.
4. **No plaintext secret touches disk or version control.** The Valkey credential
   lives in 1Password; `.env` holds only an `op://` reference. The service
   account token is provided at runtime via `OP_SERVICE_ACCOUNT_TOKEN` and is
   never written down.
5. **Resolve, use, discard.** The credential is resolved from 1Password at
   startup, used to build the Valkey client, and never written anywhere else.

## The series

Each step of the hardening journey is its own branch. `main` is the landing page
that lists them. Each branch is a complete, runnable version of the same agent.
Keep the diff between consecutive branches small: the diff is the story.

## Run commands

Export a 1Password service account token, then:

```bash
export OP_SERVICE_ACCOUNT_TOKEN=ops_...
op run --env-file=op.env -- docker compose up --build -d   # resolves the credential, starts the stack
# UI at http://localhost:8080

docker compose exec valkey-prod valkey-cli --no-auth-warning -a "$(op read 'op://Agent Prod/valkey-prod/password')" HGETALL customer:0001
docker compose down
```

Prerequisites: Docker, and a 1Password Business account with a service account
scoped to read `op://Agent Prod/valkey-prod/password`. Ollama and the model run
as compose services. On Apple Silicon the containerized Ollama is CPU-only; see
the README for host Ollama.

Conversations have short-term memory: the server keeps per-session history so
follow-up questions resolve against earlier turns.

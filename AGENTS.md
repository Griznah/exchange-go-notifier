# AGENTS.md

This guide is for agentic coding agents working in the exchange-go-notifier repository.

## Commands

```bash
# Run app locally (uses ./api_state.json for persisted state; created on first
# request if missing. Export keys first:
#   set -a; . ./.env; set +a
# )
go run .

# Run all tests
go test -v

# Run a single test by name
go test -run TestExchangeRateHandler_InputValidation
```

### Podman

```bash
# Build image
podman build -t localhost/exchange-go-notifier:dev .

# Run manually (--rm and --user keep the state file host-owned)
mkdir -p data
podman run --rm --userns=keep-id --user "$(id -u):$(id -g)" \
  -e API_STATE_FILE=/data/api_state.json --env-file .env \
  -p 8080:8080 -v ./data:/data:Z localhost/exchange-go-notifier:dev

# Or with Podman Compose (export UID/GID so the state file is host-owned)
mkdir -p data
env UID=$(id -u) GID=$(id -g) podman-compose up
```

### Environment Setup

1. `cp api_state.example.json api_state.json`
2. Create `.env` with:
   - `EXCHANGERATE_API_KEY` (ExchangeRate-API)
   - `OPENEXCHANGERATES_APP_ID` (Open Exchange Rates)

## Workflow

All changes go through feature branches and PRs — no direct commits to `main`.

Before committing: `go fmt ./... && go vet ./... && go test ./...`, then review the diff with a review skill (e.g. ponytail-review, caveman-review, or `/review` in your agent). Open PRs with the `gh` CLI.

- Branch from `main`, rebase if it has moved
- One PR per change; keep diffs small
- Never merge own PR without review unless trivial
- Never force-push shared branches

## Architecture

Single-file Go application (`main.go`): unified HTTP API over multiple exchange rate providers.

- **API config** (`APIs` slice): registry of providers, limits, endpoints, runtime state
- **State** (`apiStateMutex`): guards request counters persisted to `api_state.json`; monthly limits with automatic 30-day reset per provider — see `api_state.example.json` for the file format
- **Handlers**: `exchangeRateHandler` (`/exchange-rates`, with input validation), `healthHandler` (`/health`)
- **Fetch** (`fetchExchangeRates`): normalizes provider-specific formats into one `{"rates": {...}}` response

Errors return JSON bodies with appropriate HTTP status codes.

## Code Style

- Standard library when possible; idiomatic error handling (`if err != nil` with clear messages)
- Table-driven tests with `t.Run()` subtests; external API calls mocked via test servers
- `.yaml` never `.yml`
- Follow the Go standards in `.github/go.instructions.md`

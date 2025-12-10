# AGENTS.md

This guide is for agentic coding agents working in the exchange-go-notifier repository.

## Build, Lint, and Test Commands
- **Run app:** `go run .`
- **Run with Docker:** `docker compose up`
- **Build Docker image:** `docker build -t exchange-go-notifier .`
- **Run all tests:** `go test -v`
- **Run a single test:** `go test -v -run TestName`
- **Format code:** `go fmt ./...`
- **Vet code:** `go vet ./...`

## Code Style Guidelines
- **Imports:** Group standard, third-party, and local imports separately.
- **File naming:** Use `.yaml` for YAML files (never `.yml`).
- **Types:** Use explicit types; structs and exported functions/types use PascalCase.
- **Variables/Functions:** Use camelCase for unexported, PascalCase for exported.
- **Error Handling:** Always check errors (`if err != nil`), return/handle with clear messages.
- **Formatting:** Use `go fmt` for consistent style.
- **Naming:** Be descriptive and concise; avoid abbreviations except for well-known ones (e.g., API, URL).
- **Testing:** Use table-driven tests and clear assertions.

## Other Notes
- Environment variables are required for API keys (see README).
- No Cursor rules present.
- See README.md for more details on usage and environment setup.

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Common Development Commands

### Running the Application
```bash
# Local development (requires api_state.json and .env files)
go run .

# With Docker Compose
docker compose up

# Manual Docker build and run
docker build -t exchange-go-notifier .
docker run -p 8080:8080 --env-file .env -v $(pwd)/api_state.json:/app/api_state.json exchange-go-notifier
```

### Testing
```bash
# Run all tests with verbose output
go test -v

# Run specific test categories
go test -run TestAPIsInitialization
go test -run TestExchangeRateHandler_InputValidation
go test -run TestStatePersistence
```

### Environment Setup
1. Copy the state file: `cp api_state.example.json api_state.json`
2. Create `.env` file with required API keys:
   - `EXCHANGERATE_API_KEY` (for ExchangeRate-API)
   - `OPENEXCHANGERATES_APP_ID` (for Open Exchange Rates)

## Architecture Overview

This is a single-file Go application (`main.go`) that provides a unified HTTP API for multiple exchange rate providers. The architecture follows these principles:

### Core Components

1. **API Configuration** (`main.go:29-44`): Global `APIs` slice defines supported providers with their limits and endpoints
2. **State Management** (`main.go:50-95`): Thread-safe persistence of request counters to `api_state.json`
3. **HTTP Handler** (`main.go:212-263`): Single `/exchange-rates` endpoint with input validation
4. **Request Processing** (`main.go:130-186`): Fetches and transforms data from different APIs

### Key Patterns

- **Global State**: APIs slice holds configuration and runtime state
- **Mutex Protection**: `apiStateMutex` ensures thread-safe state updates
- **Monthly Rate Limiting**: Each API has independent request tracking with automatic reset
- **Simple Error Handling**: Returns JSON error responses with appropriate HTTP status codes

### API Response Formats

The application normalizes different API response formats into a unified structure:
```json
{
  "rates": {
    "EUR": 0.92,
    "GBP": 0.79,
    "JPY": 156.42,
    "USD": 1.0
  }
}
```

### Testing Strategy

The test suite uses table-driven tests with the standard `testing` package:
- Tests are organized by functionality (initialization, validation, state persistence)
- Uses `t.Run()` for subtests
- Mocks external API calls using test servers

## Development Guidelines

Follow the Go coding standards defined in `.github/go.instructions.md`:
- Use standard library when possible
- Write idiomatic error handling
- Create table-driven tests
- Comment functions with clear purpose
- Avoid overengineering

## State File Format

`api_state.json` stores request counts and reset times:
```json
[
  {
    "Name": "er-a",
    "RequestCount": 45,
    "LastReset": "2025-12-01T00:00:00Z"
  }
]
```

## API Providers

| Provider | ID | Limit | Reset |
|----------|----|-------|-------|
| ExchangeRate-API | `er-a` | 1,500/month | 30 days |
| Open Exchange Rates | `oer` | 1,000/month | 30 days |
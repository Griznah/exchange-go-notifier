# exchange-go-notifier

An app with selectable API's for checking exchange rates and optional automatic notifications using different backends, built with Go.

## Features

- Multiple exchange rate API support (ExchangeRate-API, Open Exchange Rates)
- Rate limiting and request tracking
- Simple HTTP API
- Containerized with Podman (rootless, no daemon)
- State persistence between restarts

## Prerequisites

- Go 1.16+
- Podman, optionally with `podman-compose` (see [PODMAN_USAGE.md](PODMAN_USAGE.md))
- API keys for the desired exchange rate providers

## Environment Variables

Copy `.env.example` to `.env` and fill in your keys:

```sh
cp .env.example .env
```

```env
# ExchangeRate-API (er-a) - Get your key from https://www.exchangerate-api.com/
EXCHANGERATE_API_KEY=your_api_key_here

# Open Exchange Rates (oer) - Get your key from https://openexchangerates.org/
OPENEXCHANGERATES_APP_ID=your_app_id_here

# State file path inside the container (with the ./data:/data mount).
# Optional for local `go run .`; defaults to api_state.json in the working dir.
API_STATE_FILE=/data/api_state.json
```

## Getting Started

1. Copy `.env.example` to `.env` and fill in your API keys (see below)

2. Run the application:

   ```sh
   # Using Go (writes state to ./api_state.json in the current directory)
   go run .

   # Using Podman (state persists to ./data/api_state.json via the mount)
   podman run --userns=keep-id --user "$(id -u):$(id -g)" --env-file .env \
     -p 8080:8080 -v ./data:/data localhost/exchange-go-notifier:dev

   # Or using Podman Compose (reads keys from .env)
   podman-compose up
   ```

   In a container the app writes its state to `/data/api_state.json`. Create the
   `./data` dir first (`mkdir -p data` — Podman doesn't auto-create bind mounts);
   the app creates the state file on first request. To seed zero counters instead:
   `cp api_state.example.json data/api_state.json`

The server will start on `http://localhost:8080`

## API Usage

### Get Exchange Rates

```http
GET /exchange-rates?api={provider}&base={currency}
```

#### Parameters

- `provider` (required): The API provider to use (`er-a` or `oer`)

#### Example Requests

```bash
# Using curl
curl "http://localhost:8080/exchange-rates?api=er-a&base=USD"

# Using httpie
http ":8080/exchange-rates" api==er-a base==USD
```

#### Example Response

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

## Available API Providers

| Provider | ID  | Request Limit | Documentation |
|----------|-----|--------------|---------------|
| ExchangeRate-API | `er-a` | 1,500/month | [docs](https://www.exchangerate-api.com/docs) |
| Open Exchange Rates | `oer` | 1,000/month | [docs](https://docs.openexchangerates.org/) |

## Development

### Running Tests

```sh
go test -v
```

### Building with Podman

```sh
podman build -t localhost/exchange-go-notifier:dev .
mkdir -p data
podman run --userns=keep-id --user "$(id -u):$(id -g)" --env-file .env \
  -p 8080:8080 -v ./data:/data localhost/exchange-go-notifier:dev
```

`--user` keeps the state file owned by you; see [PODMAN_USAGE.md](PODMAN_USAGE.md)
for the full setup, including running without `--user`.

## License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.

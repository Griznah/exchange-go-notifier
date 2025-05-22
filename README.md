# exchange-go-notifier

An app with selectable API's for checking exchange rates and optional automatic notifications using different backends, built with Go.

## Features

- Multiple exchange rate API support (ExchangeRate-API, Open Exchange Rates)
- Rate limiting and request tracking
- Simple HTTP API
- Containerized with Docker
- State persistence between restarts

## Prerequisites

- Go 1.16+
- Docker and Docker Compose (optional)
- API keys for the desired exchange rate providers

## Environment Variables

Create a `.env` file in the project root with the following variables:

```env
# ExchangeRate-API (er-a) - Get your key from https://www.exchangerate-api.com/
EXCHANGERATE_API_KEY=your_api_key_here

# Open Exchange Rates (oer) - Get your key from https://openexchangerates.org/
OPENEXCHANGERATES_APP_ID=your_app_id_here
```

## Getting Started

1. Copy the example state file:

   ```sh
   cp api_state.example.json api_state.json
   ```

2. Set up environment variables (see above)

3. Run the application:

   ```sh
   # Using Go
   go run .
   
   # Or using Docker Compose
   docker compose up
   ```

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

### Building with Docker

```sh
docker build -t exchange-go-notifier .
docker run -p 8080:8080 --env-file .env exchange-go-notifier
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

# exchange-go-notifier

An app with selectable API's for checking exchange rates and optional automatic notifications using different backends, built with Go.

## Getting Started

1. Copy the example state file:

   ```sh
   cp api_state.example.json api_state.json
   ```

2. Build and run with Docker Compose:

   ```sh
   docker compose up
   ```

- Make sure to set your API keys in the environment or in an `.env` file.
- The `api_state.json` file must exist before starting the container (see above).

# exchange-go-notifier Docker Usage

## Build the Docker image

```sh
docker build -t exchange-go-notifier .
```

## Run the container

```sh
docker run --rm -p 8080:8080 \
  -e EXCHANGERATE_API_KEY=your_key \
  -e OPENEXCHANGERATES_APP_ID=your_app_id \
  -v $(pwd)/api_state.json:/app/api_state.json \
  exchange-go-notifier
```

- The app will listen on port 8080.
- API keys are provided via environment variables.
- The state file is persisted using a volume mount.
- You can also mount `export.env` if you want to use `--env-file`.

## Example with env file

```sh
docker run --rm -p 8080:8080 \
  --env-file export.env \
  -v $(pwd)/api_state.json:/app/api_state.json \
  exchange-go-notifier
```

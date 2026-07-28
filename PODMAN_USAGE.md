# exchange-go-notifier Podman Usage

Runs rootless with [Podman](https://podman.io/). No daemon required.

Launched the same way as `absa-ac`: `--userns=keep-id` + a host-owned, read-write
`./data:/data` mount + `--env-file .env`.

## Prerequisites

- Podman (and, for the Compose workflow, `podman-compose` or a `podman compose` provider)
- For rootless mode, the `uidmap` package must be installed (`newuidmap`/`newgidmap`):
  ```sh
  sudo apt install -y uidmap
  ```

## Set up `.env`

```sh
cp .env.example .env
# edit .env and fill in EXCHANGERATE_API_KEY / OPENEXCHANGERATES_APP_ID
```

`.env` is plain `KEY=VALUE` (no `export`) so `--env-file` can read it. It holds
only your API keys — `API_STATE_FILE` is **not** stored here, because this file is
also loaded for local `go run .` (where it would override the working-dir default).
Pass it container-side via `-e API_STATE_FILE=/data/api_state.json` (or set it in
`compose.yaml`), pointing the app at the mounted `./data` dir.

## Build the image

Podman reads the `Dockerfile` natively:

```sh
podman build -t localhost/exchange-go-notifier:dev .
```

## Run the container

Create the `./data` dir first — rootless Podman does not auto-create missing bind
mounts:

```sh
mkdir -p data
podman run --userns=keep-id --user "$(id -u):$(id -g)" \
  -e API_STATE_FILE=/data/api_state.json --env-file .env \
  -p 8080:8080 -v ./data:/data:Z localhost/exchange-go-notifier:dev
```

- `-e API_STATE_FILE=/data/api_state.json` points the app at the mounted `./data`
  dir (the container WORKDIR is `/app`, so without this the state file would land
  inside the ephemeral container filesystem).

- `--userns=keep-id` maps your host user into the container so it can write to the
  bind-mounted `./data` dir (rootless Podman otherwise can't write host-owned dirs).
- `--user "$(id -u):$(id -g)"` runs the process as your own UID/GID, so the state
  file is **owned by you** (not a mapped subuid) and stays directly editable. It's
  still non-root inside the container.
- `-v ./data:/data:Z` exposes the host `./data` dir as `/data`; the app writes its
  state to `/data/api_state.json` (`./data/api_state.json` on the host) and creates
  that file on the first successful request. The `:Z` suffix relabels the mount for
  SELinux-enabled hosts (Fedora/RHEL); it's a no-op elsewhere.
- API keys are injected from `.env` via `--env-file`.

> Without `--user`, the app runs as the image's non-root `appuser`, which `keep-id`
> maps to a host subuid (e.g. UID 100100) — the plain `absa-ac`-aligned form, and
> the same behavior `absa-ac` gets. Functionally fine (you still own `./data`), but
> the state file won't be directly owned by you.

To seed zero counters instead of letting the app create the file:

```sh
cp api_state.example.json data/api_state.json
```

This targets the default path (`API_STATE_FILE=/data/api_state.json` →
`./data/api_state.json` on the host). If you customized `API_STATE_FILE`, copy to
the host path that maps to it instead.

## Using an env file you already have

If your keys live in a shell file that uses `export` (e.g.
`~/vars/exchange-go-notifier.env`), `--env-file` can't parse it — `source` it and
pass the vars through instead:

```sh
source ~/vars/exchange-go-notifier.env
mkdir -p data
podman run --userns=keep-id --user "$(id -u):$(id -g)" -p 8080:8080 -v ./data:/data:Z \
  -e EXCHANGERATE_API_KEY -e OPENEXCHANGERATES_APP_ID \
  -e API_STATE_FILE=/data/api_state.json \
  localhost/exchange-go-notifier:dev
```

## Using Podman Compose

```sh
mkdir -p data
UID=$(id -u) GID=$(id -g) podman-compose up      # or: podman compose up
```

`compose.yaml` mounts `./data:/data:Z`, sets `API_STATE_FILE=/data/api_state.json`
in the service environment, and uses `userns_mode: "keep-id"` with
`x-podman: in_pod: false` (keep-id needs a standalone container, not a pod). For
the state file to be owned by you, export `UID`/`GID` as shown — bash doesn't
export them by default, so interpolation otherwise falls back to 1000:1000. Keys
are read from `.env` automatically.

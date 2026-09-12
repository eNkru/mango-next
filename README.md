# Mango

![banner-paddings](https://user-images.githubusercontent.com/38988286/199423262-68f03906-5444-499c-8616-aa675039544e.png)

A self-hosted manga server and web-based reader. The server is a single
**Go** binary with HTML templates, the React production bundle, fonts, and
images embedded via `//go:embed` — no runtime dependency on Node, npm, or any
CDN.

This repository (`eNkru/mango-next`) is a from-scratch Go rewrite of the
original [getmango/Mango](https://github.com/getmango/Mango) (which was
written in Crystal). The Crystal codebase and its `spec/` suite have been
removed; Go is the only implementation and the only test suite.

## ✨ Features

- **Multi-user support** — manage multiple users with per-user permissions and preferences.
- **OPDS catalog** — connect external manga readers and apps via an OPDS feed (`/opds`).
- **Theme customization** — built-in light/dark mode toggle.
- **Multiple archive formats** — `.cbz`, `.zip`, `.cbr`, `.rar`, and `.7z`.
- **Nested library** — supports nested folders for structural organization.
- **Smart resume** — automatically preserves reading progress per user.
- **Thumbnail generation** — grid display with generated cover thumbnails.
- **Responsive web reader** — desktop, tablet, and mobile, with no dedicated app required.
- **Pain-free deployment** — all static assets are embedded in a single binary.
- **Chinese (zh-CN) localization** — UI translated into Chinese, with frontend JS
  dependencies bundled locally (no external CDN reliance).

## How to develop

**Prerequisites:** Go 1.26+ and Node.js 24+ with npm. Node is required only to
build the React frontend into `go/web/public/react/` before `go:embed`; the
built Mango binary has no Node/npm or CDN runtime dependency.

```bash
make run          # build React assets, then go run

# Or, build the frontend explicitly first:
npm ci && npm run build
cd go && go run ./cmd/mango/
```

Clean frontend install and production build:

```bash
npm ci
npm run typecheck
npm run build
npm run check
```

The server starts on port **9000** by default. On first launch against an empty
database it creates a default config and an admin user, printing the random
admin password to stdout — **save this password**.

### Tests & checks

```bash
make test    # go test ./...
make check   # React typecheck + output check, then go vet
make all     # check + test + build
```

### Frontend / UI

The UI is migrating to **React 19 + Vite 6 + TypeScript**. Migrated routes use
a Go HTML shell (`go/web/views/react-shell.tmpl`) that loads the React bundle
from `go/web/public/react/`. Unmigrated routes still use Go templates under
`go/web/views/` with legacy CSS/JS. Both paths are served by the same Go server
with the same auth and `BaseURL` mount. See
[FRONTEND_DEV_GUIDE.md](FRONTEND_DEV_GUIDE.md) for the route contract and dev
workflow (including Vite HMR with `npm run dev` + `npm run server`).

## How to build

```bash
make build     # builds React assets, then produces ./mango
make static    # fully static binary (CGO_ENABLED=0), ideal for Docker
```

Or directly:

```bash
cd go && go build -o ../mango ./cmd/mango/
```

The binary embeds HTML templates, the React production bundle, legacy CSS/JS,
fonts, and images. Docker regenerates the React assets from `package-lock.json`
in a Node build stage before the Go stage embeds them.

### Run locally

```bash
make build
./mango

# Custom config:
./mango -c /path/to/config.yml

# Env overrides:
PORT=9001 DB_PATH=/tmp/mango.db ./mango
```

**Environment variables** (precedence: config file > env > default; see
`go/internal/config/config.go`):

`HOST`, `PORT`, `BASE_URL`, `SESSION_SECRET` *(parsed for compatibility, unused
in Go — tokens are DB-backed)*, `LIBRARY_PATH`, `LIBRARY_CACHE_PATH`, `DB_PATH`,
`UPLOAD_PATH`, `SCAN_INTERVAL_MINUTES`, `THUMBNAIL_GENERATION_INTERVAL_HOURS`,
`LOG_LEVEL`, `CACHE_ENABLED`, `CACHE_SIZE_MBS`, `CACHE_LOG_ENABLED`,
`DISABLE_LOGIN`, `DEFAULT_USERNAME`, `AUTH_PROXY_HEADER_NAME`.

#### Config example (`~/.config/mango/config.yml`)

```yaml
host: 0.0.0.0
port: 9000
base_url: /                          # non-root e.g. /mango/ mounts all routes under that prefix
library_path: /path/to/your/manga/collection
library_cache_path: /path/to/library.yml.gz
db_path: /path/to/mango.db
upload_path: /path/to/uploads
scan_interval_minutes: 5             # background library rescan interval
thumbnail_generation_interval_hours: 24
log_level: info                      # debug | info | warn | error
cache_enabled: true                  # false skips library cache load/save
disable_login: false
# auth_proxy_header_name: X-Remote-User  # only behind a reverse proxy that strips/overwrites this header
# session_secret: ignored in Go (DB-backed tokens)
# cache_size_mbs / cache_log_enabled: parsed for compatibility, unused in Go
```

#### Auth and reverse proxies

- Auth cookies are `HttpOnly` + `SameSite=Lax`. `Secure` is set automatically when
  the request is HTTPS or `X-Forwarded-Proto: https` is present (plain local HTTP
  keeps working).
- If you terminate TLS at a reverse proxy, set `X-Forwarded-Proto` correctly and
  **strip/overwrite client-supplied** `X-Forwarded-Proto` values.
- `auth_proxy_header_name` trusts that header for any existing username. Only
  enable it when Mango is not directly reachable and the proxy overwrites the
  header on every request. The process logs a startup warning when this option
  is set.
- Login failures are rate-limited per client IP (`RemoteAddr`, ~5/minute). Edge
  proxies should apply their own limits as well.
- Browser CORS no longer sends `Access-Control-Allow-Origin: *`; same-origin UI
  and non-browser Bearer/OPDS clients are unchanged.

### Admin CLI

```bash
./mango admin user list
./mango admin user add -u alice -p secret -a   # -a marks the user as admin
./mango admin user update -u newname -p secret alice
./mango admin user delete alice
```

`mango admin` operates directly on the configured database (no server needs to
be running).

## Docker

```bash
docker build -t mango .
docker run -d \
  -p 9000:9000 \
  -v /path/to/data:/root/mango \
  -v /path/to/config:/root/.config/mango \
  --name mango \
  mango
```

Or with Compose (the container always listens on **9000** inside the image):

```bash
cp env.example .env   # edit MAIN_DIRECTORY_PATH / CONFIG_DIRECTORY_PATH / PORT
docker compose config # should succeed
docker compose up -d
```

`env.example` defaults to `./data` and `./config`. The host `PORT` only controls
the published host port (`${PORT}:9000`).

**Backup / rollback:** stop the container, copy the data and config directories,
pull or rebuild a previous image tag, and restart. The first-admin password is
printed once to the container log on an empty DB.

The Dockerfile is a multi-stage build: a `node:24-alpine` stage runs the Vite
build, a `golang:1.26-alpine` stage embeds the built assets and compiles a
static `CGO_ENABLED=0` binary, and the final image is `FROM scratch`. It
supports multi-arch builds (`linux/amd64`, `linux/arm64`, `linux/arm/v7`).

### Pre-built Docker Hub image

- **Repository**: [`enkru/mango`](https://hub.docker.com/r/enkru/mango)
- **Pull**: `docker pull enkru/mango:latest`

> Prefer building from this repo's `Dockerfile` for the latest tree.

For publishing your own image, see [DOCKER_HUB.md](DOCKER_HUB.md). For QNAP NAS
deployment, see [DEPLOY_QNAP.md](DEPLOY_QNAP.md).

## Notes

- SQLite DB schema and the on-disk library layout are stable; existing data
  directories continue to work.
- On first launch the server creates an admin user with a random password
  (printed to stdout). **Save this password.**
- The static binary (`make static`) has no runtime C dependencies.
- JSON API routes live under `{base_url}api/...` (authenticated). There is no
  embedded OpenAPI/ReDoc UI in this build.
- Tests: `cd go && go test ./...` (the Crystal `spec/` suite was removed; Go is
  the only suite).

## Special thanks

[LINUX DO](https://linux.do/)

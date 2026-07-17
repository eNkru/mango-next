# Mango (Chinese Regional Fork)

![banner-paddings](https://user-images.githubusercontent.com/38988286/199423262-68f03906-5444-499c-8616-aa675039544e.png)

This repository is a customized fork of [getmango/Mango](https://github.com/getmango/Mango). It has been tailored for local deployment in regions with restricted access to public CDNs, optimized for Synology NAS environments, and translated into Chinese.

Mango is a self-hosted manga server and web-based reader. The server is implemented in **Go** — a single binary with templates and static assets embedded.

## ✨ Features

- **Multi-user support**: Manage multiple users with customized permissions and preferences.
- **OPDS support**: Connect with external manga readers and applications easily.
- **Theme customization**: Built-in dark/light mode toggle.
- **Support for multiple format files**: `.cbz`, `.zip`, `.cbr`, `.rar`, and `.7z`.
- **Organized library**: Supports nested folders for structural organization within your library.
- **Smart resume**: Automatically preserves reading progress for each user.
- **Thumbnail generation**: Beautiful grid display with generated cover thumbnails.
- **Plugin system**: Support downloading from third-party websites with plugins.
- **Responsive Web Reader**: Enjoy responsive viewing across all devices (desktop, tablet, mobile), without the need for a dedicated app.
- **Pain-free deployment**: All static assets are embedded within a single binary.

### 🌟 Fork-Specific Features
- Fully translated to Chinese.
- JS frontend dependencies are localized (removed external CDN reliance for mainland users).
- Fixed Docker file permission limitations in Synology NAS environments.
- Packed easily installable Synology SPK (install directly via Package Center).

---

## 🛠️ How to Develop

**Prerequisites:** Go 1.26+

```bash
# Build and run:
make run

# Or:
cd go && go run ./cmd/mango/
```

The server starts on port 9000 by default. On first launch it creates a default config and admin user (password printed to stdout).

### Tests & checks

```bash
make test    # go test ./...
make check   # go vet
make all     # check + test + build
```

### Frontend / UI

Templates and static assets live under `go/web/` and are embedded into the binary. See [FRONTEND_DEV_GUIDE.md](FRONTEND_DEV_GUIDE.md).

---

## 📦 How to Build

```bash
make build     # produces ./mango
make static    # fully static binary (CGO_ENABLED=0), ideal for Docker
```

Or:

```bash
cd go && go build -o ../mango ./cmd/mango/
```

The binary embeds HTML templates, JavaScript, CSS, and images.

### Run locally

```bash
make build
./mango

# Custom config:
./mango -c /path/to/config.yml

# Env overrides:
PORT=9001 DB_PATH=/tmp/mango.db ./mango
```

**Env vars:** `HOST`, `PORT`, `BASE_URL`, `SESSION_SECRET`, `DB_PATH`, `LIBRARY_PATH`, `QUEUE_DB_PATH`, `LOG_LEVEL`, `DISABLE_LOGIN`, etc. (see `go/internal/config/config.go`).

#### Config example (`~/.config/mango/config.yml`)

```yaml
host: 0.0.0.0
port: 9000
base_url: /                          # non-root e.g. /mango/ mounts all routes under that prefix
library_path: /path/to/your/manga/collection
db_path: /path/to/mango.db
queue_db_path: /path/to/queue.db
log_level: info                      # debug | info | warn | error
download_timeout_seconds: 30         # plugin HTTP / page download timeout
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
./mango admin user add --username alice --password secret
./mango admin user update ...
./mango admin user delete ...
```

---

## 🐳 Docker

```bash
docker build -t mango .
docker run -d \
  -p 9000:9000 \
  -v /path/to/data:/root/mango \
  -v /path/to/config:/root/.config/mango \
  --name mango \
  mango
```

Or with Compose (container always listens on **9000** inside the image):

```bash
cp env.example .env   # edit MAIN_DIRECTORY_PATH / CONFIG_DIRECTORY_PATH / PORT
docker compose config # should succeed
docker compose up -d
```

`env.example` defaults to `./data` and `./config`. Host `PORT` only controls the
published host port (`${PORT}:9000`).

**Backup / rollback (disposable check):** stop the container, copy the data and
config directories, pull or rebuild a previous image tag, restart. First-admin
password is printed once to the container log on empty DB.

See [DOCKER_HUB.md](DOCKER_HUB.md) for publishing. For QNAP NAS, see [DEPLOY_QNAP.md](DEPLOY_QNAP.md).

### Pre-built Docker Hub image

- **Repository**: [`enkru/mango`](https://hub.docker.com/r/enkru/mango)
- **Pull**: `docker pull enkru/mango:latest`

> Prefer building from this repo’s Go `Dockerfile` for the latest tree.

---

## Notes

- SQLite DB schema and plugin directory layout are stable; existing data directories continue to work.
- On first launch the server creates an admin user with a random password (printed to stdout). **Save this password**.
- Static binary (`make static`) has no runtime C dependencies.
- JSON API routes live under `{base_url}api/...` (authenticated). There is no embedded OpenAPI/ReDoc UI in this build.
- Tests: `cd go && go test ./...` (Crystal `spec/` was removed; Go is the only suite).

## Special thanks
[LINUX DO](https://linux.do/)

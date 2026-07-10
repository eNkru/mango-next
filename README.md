# Mango (Chinese Regional Fork)

![banner-paddings](https://user-images.githubusercontent.com/38988286/199423262-68f03906-5444-499c-8616-aa675039544e.png)

This repository is a customized fork of [getmango/Mango](https://github.com/getmango/Mango). It has been tailored for local deployment in regions with restricted access to public CDNs, optimized for Synology NAS environments, and translated into Chinese. 

Mango is a self-hosted manga server and web-based reader. 

## ✨ Features

- **Multi-user support**: Manage multiple users with customized permissions and preferences.
- **OPDS support**: Connect with external manga readers and applications easily.
- **Theme customization**: Built-in dark/light mode toggle.
- **Support for multiple format files**: Seamlessly handle `.cbz`, `.zip`, `.cbr`, and `.rar` formats.
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

There are two versions:

| Version | Language | Requirements | Status |
|---------|----------|-------------|--------|
| **Go** (recommended) | Go | Only Go 1.26+ | Active migration |
| Crystal | Crystal + Node.js | Crystal, Shards, Node.js, Yarn, system libs | Original |

### Go Development (Recommended)

Go development is much simpler — no Node.js, no system libraries, just Go:

**Prerequisites:** Go 1.26+

**Steps:**
```bash
# Build and run with one command:
make go-run

# Or manually:
cd go && go run ./cmd/mango/
```

The Go server starts on port 9000 by default. On first launch it creates a default config and admin user.

### Crystal Development

Crystal requires more tooling:

**Prerequisites**
- **Crystal `1.0.0` or higher & Shards** (macOS: `brew install crystal`, Ubuntu/Debian: `sudo snap install crystal --classic`)
- **Node.js & Yarn** (macOS: `brew install node yarn`)
- **Required system libraries** (e.g., `libarchive`, `sqlite3`, and `wget`. macOS: `brew install libarchive sqlite3 wget`)

**Steps**
1. Clone the repository and navigate into the directory.
2. Initialize and compile the frontend assets:
   ```bash
   make setup
   ```
   (This runs `yarn` install and the gulp development task).
3. To start the application in development mode:
   ```bash
   make run
   ```
4. Testing & linting:
   - Run tests: `make test` (Executes `crystal spec`)
   - Check formatting and linting: `make check` (Executes `crystal tool format` and `ameba`)

---

## 📦 How to Build

Building Mango is simplified by the provided `Makefile`.

1. **Install Shards (Crystal dependencies):**
   ```bash
   make libs
   ```
2. **Build frontend UI assets:**
   ```bash
   make uglify
   ```
3. **Build the standard binary:**
   ```bash
   make build
   ```
   *This compiles `src/mango.cr` and produces the `mango` binary executable.*

**Note:** For a fully standalone statically linked binary, you can run `make static`.
For ARM architectures:
- 32-bit ARM: `make arm32v7`
- 64-bit ARM: `make arm64v8`

---

## Go Migration

Mango is being migrated from Crystal to Go. The Go version lives under `go/` and is **fully self-contained** — no Crystal, no Node.js, no system libraries needed. Everything (templates, JS, CSS, images) is embedded into a single binary.

### Prerequisites

- **Go 1.26+** — install via `brew install go` (macOS) or download from [go.dev](https://go.dev/dl/)
- No other dependencies needed (SQLite, archive libraries, etc. are built-in)

### Build the Binary

```bash
# From the project root:
make go-build       # builds go/mango-go — a dynamically linked binary
make go-static      # builds go/mango-go — fully static (no cgo), ideal for Docker
```

Or manually:

```bash
cd go
go build -o mango-go ./cmd/mango/
```

The binary embeds all HTML templates, JavaScript, CSS, and images — it's a single file you can move anywhere.

### Run Locally (first time)

When you start the binary for the first time, it will:
1. Create a default config file at `~/.config/mango/config.yml`
2. Create a SQLite database at `~/.config/mango/...` (see the generated config)
3. Print a one-time admin password to the console

**Quick start:**

```bash
# Build
make go-build

# Run (creates default config + admin user on first launch)
./mango-go

# Or specify a custom config path:
./mango-go -c /path/to/config.yml
```

**Using environment variables** (override config file values):

```bash
PORT=9001 DB_PATH=/tmp/mango.db ./mango-go
```

**Available env vars:** `HOST`, `PORT`, `BASE_URL`, `SESSION_SECRET`, `DB_PATH`, `LIBRARY_PATH`, `QUEUE_DB_PATH`, `LOG_LEVEL`, `DISABLE_LOGIN`, etc. (see `go/internal/config/config.go`).

#### Config File Reference

The config file is YAML. Example (`~/.config/mango/config.yml`):

```yaml
host: 0.0.0.0
port: 9000
base_url: /
session_secret: your-secret-here
library_path: /path/to/your/manga/collection
db_path: /path/to/mango.db
queue_db_path: /path/to/queue.db
log_level: info
disable_login: false
```

Set `library_path` to the directory containing your `.cbz`/`.zip`/`.cbr`/`.rar` files.

### Run with Docker (Go version)

**Build the Docker image:**

```bash
# From project root:
docker build -t mango:go -f go/Dockerfile .
```

**Run:**

```bash
docker run -d \
  -p 9000:9000 \
  -v /path/to/your/manga:/library \
  -v /path/to/config.yml:/config.yml \
  --name mango \
  mango:go -c /config.yml
```

**Or use docker-compose:**

```bash
docker compose -f docker-compose.go.yml up -d
```

### Test

```bash
make go-test        # run all 170+ tests
make go-check       # go vet (static analysis)
make go-all         # vet + test + build (all in one)
```

Or manually:

```bash
cd go && go test ./...
```

### Important Notes

- The Go binary uses the **same SQLite database** and **same plugin directory** as the Crystal version — you can switch between them freely.
- On first launch, the server creates an admin user with a random password (printed to stdout). **Save this password** or add a user via the admin CLI.
- The binary is fully static (`make go-static` or `CGO_ENABLED=0 go build`) — no external dependencies at runtime. Works on any Linux distro without installing anything.

---

## 🚀 How to Deploy

The simplest and recommended way to deploy Mango is using Docker or via Synology NAS Manager. Both the Crystal and Go versions bundle all static assets (HTML/CSS/JS) inside the compiled binary. 

### Using Docker

1. **Build the image**:
   ```bash
   docker build -t mango .
   ```
2. **Run the container**:
   Ensure you mount your manga library containing your archive files to `/root/mango/library` within the container.
   ```bash
   docker run -d \\
     -p 9000:9000 \\
     -v /path/to/your/manga:/root/mango/library \\
     -v /path/to/mango/config:/root/mango/config \\
     --name manga-server \\
     mango
   ```

*Or use the included `docker-compose.yml` for convenience!*

> **💡 Go version users:** See the [Go Migration → Run with Docker](#run-with-docker-go-version) section above for a simpler, dependency-free Docker image.

### Use the Pre-built Docker Hub Image

A pre-built image is available on Docker Hub:

- **Repository**: [`enkru/mango`](https://hub.docker.com/r/enkru/mango)
- **Pull**: `docker pull enkru/mango:latest`
- **Run with Docker**:
  ```bash
  docker run -d \\
    -p 9000:9000 \\
    -v /path/to/your/manga:/root/mango/library \\
    -v /path/to/mango/config:/root/mango/config \\
    --name manga-server \\
    enkru/mango:latest
  ```
- **Run with Docker Compose**:
  Save the following to a `docker-compose.yml` (edit the volume paths as needed), then run:
  ```yaml
  services:
    mango:
      image: enkru/mango:latest
      container_name: manga-server
      ports:
        - "9000:9000"
      volumes:
        - /path/to/your/manga:/root/mango/library
        - /path/to/mango/config:/root/mango/config
  ```
  ```bash
  docker compose up -d
  ```

### Publish to Docker Hub

See [DOCKER_HUB.md](DOCKER_HUB.md) for instructions on pushing the image to Docker Hub.

### Synology NAS Direct Installation
Since this fork is heavily optimized for Synology devices, you can install the packaged `.spk` application directly via the Synology Package Center instead of using Docker manually.

---

## 🛠️ Troubleshooting

### 'ameba' failed to compile during `make setup` / `make libs`
If you see an error like `Error: type must be Ameba::Severity, not (Ameba::Severity | Nil)` during dependency installation:
- **Solution**: The `ameba` linter may be incompatible with newer versions of Crystal. This has now been moved to `development_dependencies` in `shard.yml` so that it is ignored during production builds. 
- Ensure you have the updated `shard.yml` from this repository.
- If you're encountering the error from a stale cache, run the following commands to clean the lock file and reinstall:
  ```bash
  shards install
  make setup
  ```
  
## Special thanks
[LINUX DO](https://linux.do/)

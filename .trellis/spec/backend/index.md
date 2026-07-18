# Backend Development Guidelines

> Project-specific backend contracts for the Go server under `go/`.

---

## Overview

| Location | Language | Status |
|----------|----------|--------|
| `go/` | Go | Sole implementation (Crystal removed) |

Historical migration notes: archived Trellis tasks under `.trellis/tasks/archive/2026-07/07-08-migrate-crystal-to-go/` etc.

---

## Go Migration Architecture

### Module
`github.com/eNkru/mango-next` under `go/` directory.

### Key Dependencies
- Web: `github.com/go-chi/chi/v5` — router with path params
- DB: `modernc.org/sqlite` (pure Go, no cgo)
- JS engine: `dop251/goja` (pure Go, replaces duktape)
- HTML: `PuerkitoBio/goquery` (replaces myhtml)
- CLI: `spf13/cobra` + `olekukonko/tablewriter`
- Archive: `nwaples/rardecode` + `bodgit/sevenzip` (pure Go)
- Images: `golang.org/x/image` (draw, webp)
- Crypto: `golang.org/x/crypto/bcrypt` (compatible with Crystal hashes)
- Config: `gopkg.in/yaml.v3`

### Package Structure
```
go/
├── cmd/mango/            # cobra entry + admin CLI
├── web/
│   ├── embed.go          # //go:embed views/ + public/
│   ├── views/            # Go html/template files (.tmpl)
│   └── public/           # Static assets (JS/CSS/fonts/images)
├── internal/
│   ├── config/           # YAML+env config loading
│   ├── storage/          # SQLite + migrations + CRUD (users/tags/thumbnails/progress)
│   │   └── migration/    # Versioned migrations (latest: 15)
│   ├── archive/          # zip/cbz/rar/cbr/7z reader (Reader interface)
│   ├── thumbnail/        # Image size + thumbnail generation (200w/300h)
│   ├── plugin/           # goja sandbox + mango.* helpers + v1/v2 lifecycle
│   │   ├── subscriptions.go  # Subscription CRUD + JSON file storage
│   │   ├── updater.go        # Background subscription checker
│   │   └── downloader.go     # Background download queue consumer
│   ├── queue/            # Download queue (separate SQLite DB)
│   ├── library/          # Library scanning, Title/Entry types, natural sort
│   ├── tasks/            # Background task runner (scan/thumbnail/updater/downloader)
│   └── server/           # HTTP server: routes, middleware, handlers, templates
│       ├── auth.go           # AuthMiddleware/AdminMiddleware (cookie/bearer/proxy)
│       ├── security.go       # Safe redirect, Secure cookie, login rate limit helpers
│       ├── middleware.go     # Security headers, CORS, logging, upload containment
│       ├── response.go       # sendJSON/sendError/sendImage/sendAttachment
│       ├── server.go         # Server struct, timeouts/Shutdown, RegisterRoutes
│       ├── web.go            # TemplateManager, TemplateData structs
│       ├── handlers_api.go   # 46 API handlers (library/book/page/cover/progress/etc.)
│       └── handlers_pages.go # 20+ page handlers + OPDS XML
```

### Cross-Layer Contracts
- **Auth**: Middleware extracts token from `mango-token-{port}` cookie or `Authorization: Bearer` header; stores username in request context. Falls back to legacy `mango-sessid-{port}` cookie. Also supports `disable_login` + `auth_proxy_header` (startup warning; only safe behind a reverse proxy that strips/overwrites the header). Cookie is `HttpOnly` + `SameSite=Lax`; `Secure` is set for HTTPS or `X-Forwarded-Proto=https`. Logout calls `Storage.Logout` then clears the cookie. Form and React login callbacks are restricted to same-app relative paths (`safeRedirectPath`). Page `requireAuth` redirects to BaseURL-aware `/login?callback=...`. `GET /login` is React-mounted; valid sessions redirect away. Browser login uses `POST /api/login` (JSON); form `POST /login` remains for rollback. Failed logins are rate-limited per `RemoteAddr` IP (~5/min).
- **DB**: Single writer (`SetMaxOpenConns(1)`), `PRAGMA foreign_keys=1`, version via `PRAGMA user_version`.
- **Progress**: Table `progress` (migration 14) stores per-user per-entry reading progress with page + updated_at.
- **Queue**: Separate SQLite DB at `queue_db_path` config key.
- **Thumbnails**: All formats resized to JPEG (including webp source → JPEG output).
- **Plugin IDs**: Base64-encoded in queue job IDs, decoded on read.
- **Templates**: Go `html/template` with `layout.tmpl` wrapper; pages define `{{ define "content" }}...{{ end }}` blocks.
- **Static files**: Embedded via `//go:embed` at `go/web/embed.go`, served via `http.FileServer`.
- **OPDS**: XML rendered inline (not html/template), served at `/opds` and `/opds/book/{title_id}`.

### Route Groups
- **Unauthenticated**: `/login` (GET/POST), `/logout`, static files
- **Authenticated** (cookie/bearer): `/`, `/library`, `/book/*`, `/reader/*`, `/tags`, `/api`, `/opds`, `/download/plugins`
- **Admin** (requires admin flag): `/admin/*`, `/api/admin/*`
- **API** (JSON): 46 routes under `/api` + `/api/admin`

### Route Count
Total: ~68 routes (46 API + 8 admin pages + 10 main pages + 2 reader + 2 OPDS)

---

## Docker / Deployment

### Scratch image convention
`Dockerfile` uses `FROM scratch` and **must** set `ENV HOME=/root` after `FROM scratch`.
Without it `os.UserHomeDir()` returns `/` causing `~/mango/...` defaults to expand to `/mango/...`
which collides with the binary at `/mango`.

### Volume mount paths
docker-compose mounts data at `/root/mango` and config at `/root/.config/mango`.
Default config paths (`~/mango/library`, `~/mango.db`, etc.) expand to `/root/mango/...` when `HOME=/root`.

---

## Guidelines Index

| Guide | Description | Status |
|-------|-------------|--------|
| [User Management](./user-management.md) | Admin/user storage and API invariants | Active |
| [Library Background Jobs](./library-background-jobs.md) | Cache identity, thumbnail locking, progress, and startup ordering | Active |

---

## Pre-Development Checklist

### For Go changes
- Go module: `cd go/` and run `go build ./... && go vet ./... && go test ./...`.
- Template changes: `.tmpl` Go html/template under `go/web/views/`; run `go build ./...` to verify embed patterns.
- New routes: Register in `internal/server/server.go` `RegisterRoutes()` (under `BaseURL` mount) and add handler to `handlers_api.go` or `handlers_pages.go`.
- Storage changes: Add migration in `internal/storage/migration/migrations.go` and bump `LatestVersion()`. Run full test suite.
- Config: document implemented vs deprecated keys; wire new knobs with tests.

## Quality Check

- For Go: `go build ./... && go vet ./... && go test ./...` (and `go test -race ./...` for concurrent paths).
- Add or update Trellis specs when changing storage invariants, validation errors, or API behavior.
- Crystal is no longer an implementation or test suite in this repository.

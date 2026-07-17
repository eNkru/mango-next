# Configuration deployment and documentation cleanup design

## Boundary

Truthful config, deploy docs, Compose/Makefile, dead Crystal specs, API docs
surface honesty, and full `base_url` routing. Do not own CI workflows, frontend
asset modernization, or scratch-image UID changes.

## Config field classification

| Key | Action |
|---|---|
| `host`, `port`, `library_*`, `db_path`, `queue_db_path`, intervals, paths, `disable_login`, `default_username`, `auth_proxy_header_name` | Already used — keep; document accurately |
| `base_url` | **Implement** full route mount under prefix |
| `download_timeout_seconds` | **Wire** into plugin HTTP client / downloader |
| `log_level` | **Wire** into logging (std log or thin wrapper) |
| `cache_enabled` | **Wire** library cache load/save skip when false, or mark deprecated if unsafe |
| `session_secret` | **Deprecated** — parse only; Go auth is DB token, not signed sessions |
| `cache_size_mbs`, `cache_log_enabled` | **Deprecated** — parse only; no size/log consumer in Go library cache today |

Document deprecated keys in README config section with "ignored in Go" wording
so old YAML still loads.

## BaseURL routing

### Problem

Templates and JSON use `cfg.BaseURL` as a path prefix; chi registers at `/`.
Auth `requireAuth` redirects to `/login`. Cookie Path already uses `BaseURL`.

### Approach

1. Normalize base as today (`/` or `/mango/` with trailing slash).
2. Compute mount path without trailing slash for chi (`""` for root, `/mango`
   for `/mango/`) — chi `Mount` / `Route` conventions.
3. Register the full application router under that prefix (including static,
   login/logout, API, OPDS).
4. Replace hard-coded `/login` (and similar absolute paths in server package)
   with base-aware helpers, e.g. `pathJoinBase(base, "login")` → `/login` or
   `/mango/login`.
5. Prefer not double-prefixing: if handlers already emit `BaseURL + "api/..."`,
   keep that; only fix redirects and any absolute paths that ignore BaseURL.
6. Integration test: config `base_url: /mango/`, hit `/mango/login` and a static
   or health-like route; assert `/login` without prefix is 404 (or non-app).

Optional: root request `/` when base is non-root may 404 — document that reverse
proxies should strip prefix or set base to `/`.

## Plugin download timeout

- Pass `cfg.DownloadTimeoutSeconds` into plugin sandbox HTTP client and page
  downloader instead of hard-coded 30s (default remains 30).
- Unit test: constructed client timeout matches config.

## Log level

- On startup after config load, apply `log_level` (`debug`/`info`/`warn`/`error`
  case-insensitive). Minimal acceptable: filter `log` package via custom
  `io.Writer` or document mapping if stdlib cannot fully filter — prefer a small
  helper used from `cmd/mango`.
- Test: invalid level rejected or defaulted consistently with docs.

## API docs page

- Remove `handleAPIDocs` route and nav links that present ReDoc as live docs, or
  replace body with a short "API is JSON routes under /api; OpenAPI TBD" page
  without external assets.
- Delete dependency on missing `openapi.json` / `redoc.standalone.js`.
- Prefer removal of false UI over a half-empty page.

## Compose / env

```env
# env.example
PORT=9000
MAIN_DIRECTORY_PATH=./data
CONFIG_DIRECTORY_PATH=./config
```

- `docker-compose.yml`: drop `version`; keep `${PORT}:9000`; volumes from env.
- QNAP files: drop `version`; fix `src/config.cr` comments → `go/internal/config`.
- README: `cp env.example .env` then `docker compose up -d`; note container
  always listens on 9000 inside the image.

## Makefile / docs alignment

- Make `all` match documented quality gate: `check test build` (same as
  `go-all`), **or** change README to state `all` is build-only and point
  quality to `make go-all` / `make check test`. Prefer `all: check test build`
  so one command matches CI intent; keep aliases.

## Crystal spec/

- Delete `spec/` tree (user decision).
- Update `.trellis/spec/backend/index.md` Pre-Development / Quality Check to
  remove Crystal as current gate; Go commands only; fix migration version note
  if still wrong (14 vs 15).

## Docs content

- README: config table with implemented vs deprecated; BaseURL reverse-proxy
  note; Docker compose prerequisites; make targets.
- DEPLOY_QNAP: stale Crystal paths, DB volume notes already partial — complete
  against current Go defaults and compose files.
- First-admin: point to existing log-password behavior; backup = volume + db
  paths; rollback = previous image tag / binary.

## Trade-offs

| Choice | Benefit | Cost |
|---|---|---|
| Full BaseURL mount | Real reverse-proxy subpath support | Careful path join; many redirect sites |
| Deprecate unused cache knobs | No false cache-size control | Operators must learn keys are no-ops |
| Delete Crystal specs | No fake suite | Lose historical Crystal assertions (user accepted) |
| Hide API ReDoc | Honest UX | No browsable OpenAPI until later |

## Rollback

- Config deprecations are non-breaking (still parse).
- BaseURL mount is the highest risk; isolate and test before Compose-only
  commits if needed.
- Spec deletion is irreversible in git history only — recoverable from git.

## Validation

```bash
# Go
cd go && GOCACHE=/tmp/mango-next-go-cache go test ./... && go vet ./... && go build ./...

# Compose
cp -n env.example .env   # or use example values
docker compose config
docker compose -f docker-compose.qnap.yml config
docker compose -f docker-compose.qnap-prebuilt.yml config
```

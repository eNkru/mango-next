# Mango Next project-wide technical review

## Executive assessment

The Go migration has a solid core: the application builds cleanly, passes its
tests and race detector, has strong library regression coverage, uses
transactional SQLite migrations, embeds local assets, and preserves explicit
background-job contracts. A rewrite is not justified.

The highest-value work is at boundaries the migration did not fully reconnect:
authentication lifecycle, HTTP resource limits, browser/API contracts,
configuration truthfulness, and automated integration checks. The recommended
order is:

1. P0 authentication and HTTP hardening.
2. P1 repair broken frontend/API contracts and add CI/critical route tests.
3. P1 make configuration and deployment documentation truthful.
4. P2 continue frontend dependency/build modernization within the same frontend
   task after its functional repair milestone.

Effort: S = hours, M = several focused days, L = staged multi-area work.
Confidence: Confirmed = directly traced or reproduced; Strong = complete static
call-path evidence but blocked runtime reproduction; Candidate = needs dedicated
design/validation.

## Findings

### 1. High: logout does not revoke the active token

**Evidence:** Confirmed, effort S, owner `07-17-auth-http-hardening`.

- `go/internal/server/handlers_pages.go:44-46` only clears the cookie.
- `go/internal/storage/storage.go:432-435` already provides `Storage.Logout`, but
  no server handler calls it.
- `go/internal/storage/storage.go:259-261` reuses an existing token, and
  `go/internal/server/auth.go:133-135` gives the cookie a one-year lifetime.

Logging out therefore leaves copied bearer tokens and the database token valid.
The handler should extract the presented token, revoke it, clear the cookie even
if revocation fails, and test that the old token is rejected afterward.

### 2. High, configuration-dependent: proxy auth trusts any matching header

**Evidence:** Confirmed call path, effort M, owner
`07-17-auth-http-hardening`.

`go/internal/server/auth.go:77-89` accepts the configured header from every
request, checks only that the username exists, and derives admin status from that
username. If Mango remains directly reachable when this option is enabled, a
client can send the header and impersonate any existing user, including an
admin.

This mode needs an explicit trusted-proxy boundary: bind/reachability guidance,
mandatory upstream header stripping, and ideally a trusted-proxy allowlist or an
authenticated proxy integration. This is not a vulnerability when the service
is correctly isolated behind a proxy that overwrites the header, so the report
does not claim unconditional exploitability.

### 3. High: HTTP server and request bodies are insufficiently bounded

**Evidence:** Confirmed configuration, effort M, owner
`07-17-auth-http-hardening`.

- `go/internal/server/server.go:53-60` configures no read-header, read, write, or
  idle timeout and uses immediate `Close` instead of bounded graceful shutdown.
- Unauthenticated login parsing at `handlers_pages.go:20-24` and
  `handlers_api.go:24-34` has no body limit.
- `handlers_api.go:555-606` calls `ParseMultipartForm(32 << 20)`, which controls
  in-memory buffering but does not impose a total upload limit before saving.

A slow or oversized request can retain resources; an authenticated oversized
upload can consume disk. Add server timeouts, `http.MaxBytesReader` at relevant
entry points, finite shutdown, and regression tests. Rate-limit failed logins as
part of the same boundary work.

### 4. High: download manager is disconnected from the Go queue API

**Evidence:** Confirmed cross-layer mismatch, effort M, owner
`07-17-frontend-deps-build` first milestone.

- `go/web/public/js/download-manager.js:9-31` opens a nonexistent WebSocket at
  `/api/admin/mangadex/queue`.
- Lines 33-95 use the same removed prefix for list/delete/retry/pause/resume.
- The server registers only `/api/admin/queue` and `/queue/{action}` at
  `go/internal/server/server.go:149-150`.
- `go/internal/server/handlers_api.go:850-884` returns `{success,data}` and
  accepts JSON `{id}` with only delete/retry; the page expects `{jobs,paused}`,
  query-string IDs, pause/resume, and lower-case job fields.
- No handler test exercises the page-to-queue contract.

This page cannot be fixed by changing one URL. Define one queue DTO and action
contract, decide polling versus push, remove unsupported controls or implement
them end to end, and cover list/delete/retry in a browser/API smoke test.

### 5. Medium-high: configured non-root `base_url` is not mounted by the router

**Evidence:** Strong, effort M, owner `07-17-config-deploy-docs-cleanup`.

`go/internal/config/load.go:110-117` accepts and normalizes values such as
`/mango/`. Templates then prefix links and assets with that value, but
`go/internal/server/server.go:69-160` registers every route at root and never
mounts a subrouter under `BaseURL`. Authentication redirects also hard-code
`/login` at `go/internal/server/auth.go:185-190`.

A deployment configured at `/mango/` will generate prefixed URLs the router does
not serve. Add an integration test for a non-root base path, then either mount all
routes consistently or deprecate/redefine the option.

### 6. Medium-high: plugin HTTP can hang or consume unbounded memory/disk

**Evidence:** Confirmed, effort M, owner `07-17-auth-http-hardening` for outbound
HTTP boundaries; configuration wiring belongs to
`07-17-config-deploy-docs-cleanup`.

- `go/internal/plugin/sandbox.go:35-41` creates an HTTP client without a timeout.
- `sandbox.go:193-198` reads the entire response without a size bound and ignores
  the read error.
- The page downloader hard-codes 30 seconds at
  `go/internal/plugin/downloader.go:29-42`; configured
  `download_timeout_seconds` is not used.
- `downloader.go:287-310` streams remote page bodies into archives without a
  configurable size/disk budget.

Plugins are admin-installed, so unrestricted destinations are partly an explicit
capability, not automatically an SSRF defect. Remote content is still untrusted.
Document the capability and add timeout, response-size, redirect, and disk-budget
controls that fit comic downloads.

### 7. Medium-high: no CI protects a thinly tested HTTP/migration boundary

**Evidence:** Confirmed, effort M, owner `07-17-test-ci-baseline`.

There is no `.github` workflow. Local checks pass, but coverage is about 11.2%
for `internal/server`, 0% for `cmd/mango`, and 0% direct statement coverage for
the migration package. The repository has 65 page/API handlers but only a small
set of middleware/progress handler tests. Existing Go tests total 161 while the
backend spec still states a 170-test baseline.

Add a CI gate for format, vet, test, race, and build, then prioritize route
authorization/error matrices, template/static startup, queue contracts, and
historical migration tests. Avoid chasing a single global coverage percentage.

### 8. Medium: several public configuration knobs have no runtime effect

**Evidence:** Confirmed, effort M, owner
`07-17-config-deploy-docs-cleanup`.

`go/internal/config/config.go:15,22,25,27-28` exposes `session_secret`,
`log_level`, `download_timeout_seconds`, `cache_size_mbs`, and
`cache_log_enabled`. Repository-wide usage shows these are loaded/defaulted but
not consumed by runtime behavior (apart from a log-level parsing test).

This creates false operational confidence. Classify each key: implement it,
deprecate it with compatibility handling, or remove it. Generate/document config
from the same schema and test representative effects rather than only parsing.

### 9. Medium: default Compose and build instructions are not self-consistent

**Evidence:** Confirmed, effort S-M, owner
`07-17-config-deploy-docs-cleanup`.

- `README.md:125-129` says `docker compose up -d` without required setup.
- `env.example:6,10` leaves both bind sources empty; `docker compose config`
  fails with `empty section between colons` under those defaults.
- `docker-compose.yml:9-15` uses interpolated `PORT` for metadata/published port
  but does not pass it to the container, while the target remains fixed at 9000.
- All Compose files retain obsolete `version` keys.
- `README.md:49` says `make all` runs check, test, and build, but `Makefile:8`
  makes it build-only; only `go-all` has the documented behavior.
- QNAP guidance still points to deleted `src/config.cr` and describes an old DB
  default at `docker-compose.qnap-prebuilt.yml:22-26`.

Provide safe example paths or required-variable checks, align port semantics,
remove obsolete keys, and make docs/Makefile/CI share the same commands.

### 10. Medium: API documentation is presented but cannot load

**Evidence:** Confirmed, effort S-M, owner
`07-17-config-deploy-docs-cleanup`.

`go/web/views/api.tmpl:11-12` requests `openapi.json` and
`js/redoc.standalone.js`. Neither file exists in the embedded public tree and no
OpenAPI route is registered in `server.go`. Restore a current generated/static
spec plus embedded viewer, or remove the page until it is supported.

### 11. Medium: browser dependencies and generated assets are unmanaged

**Evidence:** Confirmed inventory, risk severity not fully assessed, effort L,
owner `07-17-frontend-deps-build` after functional repairs.

The embedded bundle includes jQuery 3.2.1, jQuery UI 1.12.1, Alpine.js 2.8.0,
Moment, UIkit, Select2, and other copied assets, but no application-level package
manifest, lockfile, asset build, drift check, or browser test exists. The only
package manifest found belongs to `.opencode`, not the application.

Do not replace the server-rendered architecture. First inventory versions,
licenses, and actual use; then upgrade/remove incrementally behind browser smoke
tests. Add a reproducible LESS/asset command that checks committed generated CSS.
A vulnerability scanner was unavailable, so no unverified CVE claim is made.

### 12. Medium-low: migration-era artifacts obscure the current source of truth

**Evidence:** Confirmed, effort S-M, owner
`07-17-config-deploy-docs-cleanup`.

- Top-level `spec/` retains six Crystal specs and fixtures, but no `shard.yml` or
  Crystal implementation remains.
- `.trellis/spec/backend/index.md:11` says Go is the sole implementation while
  lines 110-123 still prescribe Crystal checks.
- The same index says latest migration 14 at line 44, while
  `go/internal/storage/migration/migrations.go:168-181` defines version 15.
- Most frontend spec files remain generic placeholders rather than documenting
  the actual template/jQuery/Alpine architecture.

Archive/remove dead executable-looking artifacts after preserving compatibility
knowledge, and refresh specs from current code. Keep useful Crystal parity
comments in Go where they explain data compatibility; do not mechanically delete
historical rationale.

### 13. Candidate defenses: upload containment, active content, and archive limits

**Evidence:** Strong/candidate, effort M, owners noted below.

- `go/internal/server/middleware.go:48-64` uses string-prefix containment for
  uploads; replace it with `filepath.Rel`-style boundary validation and test
  sibling-prefix traversal paths (`uploads` versus `uploads-other`).
- Upload type is inferred from filename extension and SVG is accepted
  (`go/internal/upload/upload.go:20-35`, `handlers_api.go:581-606`). Since upload
  is admin-only, this is not treated as an unauthenticated stored-XSS finding,
  but same-origin active content deserves an explicit policy.
- Archive page readers use `io.ReadAll` for decompressed entries. Locally managed
  libraries reduce exposure, but corrupt/hostile archives can still exhaust
  memory; add limits only after measuring legitimate page sizes.

Upload defenses belong to `07-17-auth-http-hardening`. Archive decompression
limits are deferred until a dedicated compatibility/size study defines safe
bounds.

### 14. Candidate usability work: keyboard semantics and visual regression

**Evidence:** Confirmed patterns, impact requires browser validation, effort M,
owner `07-17-frontend-deps-build`.

Several templates use clickable `<div>` cards and `<a>` elements without an
`href` (for example `library.tmpl:42`, `title.tmpl:17-18`, and
`download-manager.tmpl:51-53`). Some controls have labels, but keyboard behavior
is inconsistent. Convert commands to buttons and navigation to links as each
screen is touched, then validate both Comic/Flat themes at mobile and desktop
viewports. Do not launch a broad visual rewrite.

## What should remain

- Keep the Go monolith and internal package boundaries. Large files such as
  `storage.go` and handler files can be split opportunistically by responsibility,
  but their size alone does not justify an architecture rewrite.
- Keep `html/template`, embedded local assets, pure-Go SQLite/archive choices,
  and the scratch-image model unless measured constraints require change.
- Preserve the cache identity, nested-title, thumbnail decoder, admin-removal,
  and dual-theme contracts already captured in Trellis specs.
- Keep single-writer SQLite until profiling shows it is a bottleneck; current race
  tests pass and the design avoids lock errors.
- Do not add Redis, a SPA framework, microservices, or a new database without a
  measured problem that demands them.

## Verification performed

- Passed: `go test ./...`.
- Passed: `go test -race ./...`.
- Passed: `go vet ./...`.
- Passed: `go build ./...`.
- Measured: per-package statement coverage; library ~80.4%, server ~11.2%.
- Parsed: default and QNAP Compose variants; default fails with empty example
  bind paths, QNAP variants parse with obsolete-version warnings.
- Disposable startup reached config generation, admin creation, migration 15,
  library scan, thumbnails, plugin updater, and queue initialization.

## Verification limitations

The sandbox denied local socket binding and Docker daemon access. Approval
escalation could not run because the automatic reviewer returned HTTP 503. No
browser engine/Playwright, `govulncheck`, `gosec`, `staticcheck`,
`golangci-lint`, or application frontend package toolchain was installed.
Therefore Docker image behavior, real browser layout/interactions, and current
dependency vulnerability status remain unverified rather than assumed healthy.

## Roadmap acceptance

Every accepted recommendation maps to one of the four child PRDs. The only
explicitly deferred item is an archive decompression size policy, which needs a
real-library size study before choosing limits. Each child remains in planning
and requires its own design/implementation approval before product changes.

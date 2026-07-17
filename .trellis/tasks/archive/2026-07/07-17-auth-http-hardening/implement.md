# Authentication and HTTP hardening implementation plan

## Preconditions

- Planning artifacts reviewed: `prd.md`, `design.md`, this file.
- User approval before `task.py start` (status must be `planning` until then).
- Product code changes stay in `go/internal/server` (+ short docs updates).

## Ordered checklist

1. [x] Helpers
   - [x] `safeRedirectPath(callback string) string`
   - [x] `requestIsHTTPS(r *http.Request) bool` for Secure cookie
   - [x] `clientIP(r *http.Request) string` from `RemoteAddr` only
   - [x] In-memory login failure limiter (sliding window, 5 / min / IP)
2. [x] Cookie + logout + login handlers
   - [x] Set/clear cookie with Secure from request
   - [x] Form login: body limit, rate limit, safe callback, generic failure
   - [x] API login: body limit, rate limit, generic error JSON (no `err.Error()`)
   - [x] Logout: `extractToken` + `Storage.Logout` then clear cookie
3. [x] HTTP server lifecycle
   - [x] Timeouts on `http.Server`
   - [x] `Shutdown` with finite deadline instead of immediate `Close`
4. [x] Middleware and uploads
   - [x] Drop CORS `Access-Control-Allow-Origin: *`
   - [x] Security headers middleware
   - [x] Admin upload `MaxBytesReader` + keep JPEG/PNG-only covers
   - [x] Harden `/uploads` absolute-prefix containment if needed
5. [x] Proxy auth ops boundary
   - [x] Startup warning when `auth_proxy_header_name` is set
   - [x] Document proxy header stripping + Secure / X-Forwarded-Proto trust
6. [x] Tests for every behavior above (prefer `auth_test.go` / new focused files)
7. [x] Quality gate:
   ```bash
   cd go
   GOCACHE=/tmp/mango-next-go-cache go test ./...
   GOCACHE=/tmp/mango-next-go-cache go test -race ./...
   GOCACHE=/tmp/mango-next-go-cache go vet ./...
   GOCACHE=/tmp/mango-next-go-cache go build ./...
   ```

## Risky files

| File | Risk |
|---|---|
| `go/internal/server/auth.go` | Cookie Secure + extractToken shared by all auth paths |
| `go/internal/server/handlers_pages.go` | Login redirect + logout behavior |
| `go/internal/server/handlers_api.go` | API login errors; admin upload limits |
| `go/internal/server/server.go` | Timeouts may break slow large responses if too aggressive |
| `go/internal/server/middleware.go` | CORS change may break unexpected browser clients |
| `README.md` / deploy docs | Incomplete warning leaves proxy misconfig easy |

## Rollback points

- After helpers only: no user-visible change.
- After handlers: logout revocation is the main behavioral fix; revert handlers if needed.
- After CORS/timeouts: highest external compatibility risk; isolate in its own commit if practical.

## Review gates

- Do not implement until user reviews planning artifacts and starts the task.
- After code: run full quality gate and trellis-check before claiming done.

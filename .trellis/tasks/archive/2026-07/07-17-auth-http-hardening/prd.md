# Authentication and HTTP hardening

## Goal

Harden authentication and HTTP boundaries without breaking local HTTP, reverse
proxy, bearer-token, OPDS, or existing database compatibility.

## Confirmed Facts

- Post-login form callback is taken from `r.FormValue("callback")` and redirected
  with no allowlist (`go/internal/server/handlers_pages.go`).
- Logout only clears the cookie; `Storage.Logout` exists and is unit-tested but
  never called from the HTTP handler.
- Auth cookie is `HttpOnly` + `SameSite=Lax`, one-year MaxAge, no `Secure` flag
  (`go/internal/server/auth.go`).
- Proxy auth accepts any client-supplied value of `auth_proxy_header_name` if the
  username exists; there is no trusted-proxy allowlist or reachability check.
- `http.Server` has no Read/Write/Idle timeouts and uses immediate `Close` on
  context cancel (`go/internal/server/server.go`).
- Login body parsing (form + JSON `/api/login`) has no `MaxBytesReader` bound;
  admin multipart upload uses `ParseMultipartForm(32<<20)` without a total body
  cap before save.
- CORS middleware sets `Access-Control-Allow-Origin: *` for all responses.
- Public `/uploads` path is prefix-checked; upload MIME whitelist still includes
  `image/svg+xml` while cover upload currently rejects non-JPEG/PNG.
- `session_secret` is loaded but unused by the Go auth path (token is DB-backed).
- Existing focused auth tests cover middleware token/proxy/disable_login and
  cookie set/clear, but not login callback sanitization, logout revocation,
  rate limits, server timeouts, or CORS trust model.
- Parent review archived at
  `.trellis/tasks/archive/2026-07/07-17-project-wide-review/`.

## Requirements

- Restrict post-login redirects to safe, application-local destinations.
- Invalidate the persisted token on logout before clearing the browser cookie.
- Cookie `Secure` policy: set Secure when the request is HTTPS or
  `X-Forwarded-Proto` indicates https; leave Secure unset for plain HTTP so
  local deployments keep working. Document that reverse proxies must strip or
  overwrite client-supplied forwarded proto headers.
- Proxy-header auth trust boundary: document that the service must sit behind a
  reverse proxy that strips/overwrites the header; emit a startup warning when
  `auth_proxy_header_name` is set; add tests that the header still authenticates
  only when a matching user exists. No trusted-proxy allowlist or bind lock in
  this task.
- Login abuse protection: process-local sliding-window rate limit keyed by
  client IP on both POST `/login` and POST `/api/login` (default about 5
  failures per minute). Over-limit and failed auth return the same generic
  failure shape so username existence is not leaked. Counts reset on process
  restart; no new config knobs in this task.
- Add HTTP server read, header, write, and idle timeouts plus graceful shutdown.
- CORS: stop emitting `Access-Control-Allow-Origin: *`; same-origin browser use
  and non-browser Bearer/OPDS clients remain supported. No origin whitelist
  config in this task.
- CSRF: rely on `SameSite=Lax` auth cookies plus tightened CORS; document the
  trust model rather than adding CSRF tokens.
- Add baseline security response headers (at least `X-Content-Type-Options`,
  `X-Frame-Options`/`frame-ancestors` equivalent, and a conservative
  `Referrer-Policy`).
- Cap request bodies for login and admin upload paths; keep cover uploads
  JPEG/PNG-only; do not introduce new SVG upload surfaces; keep `/uploads`
  path containment and harden prefix checks if needed.
- Preserve API and database compatibility unless a migration is explicitly
  justified.
- Add focused regression tests for every behavior changed by this task.

## Acceptance Criteria

- [x] External and protocol-relative login callbacks cannot redirect off-site.
- [x] A token used before logout is rejected after logout.
- [x] Cookie `Secure` is true for HTTPS / `X-Forwarded-Proto=https` and false
      for plain HTTP; both paths are tested and documented for reverse proxies.
- [x] Enabling `auth_proxy_header_name` logs a startup warning; docs require a
      reverse proxy that strips/overwrites the header; tests cover header auth
      success/failure without adding a trusted-proxy allowlist.
- [x] After the IP failure budget is exceeded, further login attempts from that
      IP fail with the same generic error/redirect as a bad password, for both
      form and API login; tests cover under/over limit.
- [x] Slow or oversized requests are bounded and shutdown drains requests within
      a finite deadline.
- [x] Responses no longer advertise `Access-Control-Allow-Origin: *`; security
      headers are present on representative routes; login/upload body limits and
      `/uploads` containment have focused tests.
- [x] `go test ./...`, `go test -race ./...`, `go vet ./...`, and `go build ./...`
      pass.

## Out of Scope

- Full OIDC/SSO integration beyond the existing proxy-header mode.
- Rewriting the token model or introducing signed cookies that need
  `session_secret`.
- Frontend dependency modernization and download-manager API repair (owned by
  `07-17-frontend-deps-build`).
- Config field deprecation/docs cleanup for unused keys such as
  `session_secret` / `log_level` (owned by `07-17-config-deploy-docs-cleanup`),
  except docs required for new security knobs introduced here.
- Plugin outbound HTTP timeouts (tracked for this task only if reused as a shared
  helper; otherwise remain with config/plugin follow-up).

## Open Questions

- None currently blocking planning.

## Notes

- Dependency: none. This is the first recommended implementation task.
- Coordination: the test/CI task may proceed independently, but this task owns
  regression tests for its own security behavior.
- Evidence: `go/internal/server/auth.go`, `middleware.go`, `server.go`,
  `handlers_pages.go`, `handlers_api.go`, and current auth tests.

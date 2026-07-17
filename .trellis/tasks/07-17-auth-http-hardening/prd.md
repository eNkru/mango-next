# Authentication and HTTP hardening

## Goal

Harden authentication and HTTP boundaries without breaking local HTTP, reverse
proxy, bearer-token, OPDS, or existing database compatibility.

## Requirements

- Restrict post-login redirects to safe, application-local destinations.
- Invalidate the persisted token on logout before clearing the browser cookie.
- Define a deployment-safe `Secure` cookie policy while preserving explicit
  local-HTTP support.
- Document and enforce the trust boundary for proxy-header authentication.
- Add appropriate login abuse protection without leaking whether a username
  exists.
- Add HTTP server read, header, write, and idle timeouts plus graceful shutdown.
- Review CORS, CSRF exposure, security headers, upload size limits, uploaded SVG
  handling, and public upload path containment; apply least-privilege defaults.
- Preserve API and database compatibility unless a migration is explicitly
  justified.
- Add focused regression tests for every behavior changed by this task.

## Acceptance Criteria

- [ ] External and protocol-relative login callbacks cannot redirect off-site.
- [ ] A token used before logout is rejected after logout.
- [ ] Cookie security behavior is explicit, tested, and documented for direct
      HTTP and TLS-terminating reverse proxies.
- [ ] Proxy authentication cannot be enabled without a clearly documented trusted
      proxy/header-stripping deployment boundary.
- [ ] Login failures are rate-limited or otherwise protected against unbounded
      online guessing.
- [ ] Slow or oversized requests are bounded and shutdown drains requests within
      a finite deadline.
- [ ] CORS, CSRF, response-header, and upload decisions have tests that demonstrate
      the intended trust model.
- [ ] `go test ./...`, `go test -race ./...`, `go vet ./...`, and `go build ./...`
      pass.

## Notes

- Dependency: none. This is the first recommended implementation task.
- Coordination: the test/CI task may proceed independently, but this task owns
  regression tests for its own security behavior.
- Evidence: `go/internal/server/auth.go`, `middleware.go`, `server.go`,
  `handlers_pages.go`, `handlers_api.go`, and current auth tests.

# Authentication and HTTP hardening design

## Boundary

Change only the HTTP authentication lifecycle and request/server resource
boundaries in `go/internal/server` (plus minimal docs). Do not redesign the
DB-backed token model, OPDS Basic auth, library scanning, plugins, or frontend
asset pipeline.

## Current model

```text
Request
  -> CORSMiddleware (*)
  -> LoggingMiddleware
  -> UploadHandler (/uploads public files)
  -> public routes: /login, /logout, /api/login
  -> AuthMiddleware (token cookie|bearer, disable_login, auth_proxy_header)
  -> AdminMiddleware for /admin and /api/admin
```

Tokens live in SQLite (`Storage.VerifyUser` / `VerifyToken` / `Logout`). Cookies
are named `mango-token-{port}`, `HttpOnly`, `SameSite=Lax`, 365-day MaxAge, no
`Secure` flag today.

## Target contracts

### 1. Safe post-login redirect

- Accept only relative application paths: must start with `/`, must not start
  with `//`, must not contain a scheme (`:` before first `/` after optional
  leading slash is rejected via `url.Parse` + reject absolute URLs).
- Empty or invalid callback → `/` (or `BaseURL` home if already used elsewhere;
  default `/` matches current form behavior).
- Apply on form login only; API login returns JSON and does not redirect.

### 2. Logout revokes token

```text
handleLogout:
  token = extractToken(r, cfg)  // cookie then bearer, existing helper
  if token != "" { _ = Storage.Logout(token) }  // best-effort
  ClearAuthTokenCookie(...)
  redirect /login
```

Cookie clear always runs even if DB revoke fails. After logout, the same token
must fail `VerifyToken` for both cookie and bearer presentations.

### 3. Smart Secure cookie

`SetAuthTokenCookie` / `ClearAuthTokenCookie` take the request (or a boolean
derived from it):

- `Secure=true` when `r.TLS != nil` or
  `strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")`.
- Otherwise `Secure=false` so plain local HTTP still works.
- Docs: reverse proxies terminating TLS must set `X-Forwarded-Proto` and strip
  client-supplied values; same class of trust as proxy auth.

No new config key.

### 4. Proxy auth trust boundary

- Runtime behavior of header auth unchanged (username must exist).
- On server start (or first config use at `Start`), if
  `AuthProxyHeaderName != ""`, log a clear warning that the process must not be
  directly reachable and the proxy must overwrite/strip the header.
- README / deploy notes document the boundary.
- No `trusted_proxies` allowlist in this task.

### 5. Login rate limit

- Process-local sliding window keyed by client IP.
- Default budget: 5 failed attempts per 1 minute per IP (constants in code;
  no config knobs this task).
- Apply to POST `/login` and POST `/api/login` before password verify when
  possible; always record failures after failed verify; successful login may
  clear that IP's failure window (optional nicety, recommended).
- Client IP: `X-Forwarded-For` left-most only if we already trust proxy headers
  for Secure cookie / proxy auth docs; otherwise `RemoteAddr`. Prefer
  `net/http` RemoteAddr by default and document that accurate limits behind a
  proxy need the proxy to preserve/connect correctly — simplest safe default:
  parse `RemoteAddr` host; optionally honor `X-Real-IP` / first XFF only when
  documented as proxy deployment. **Decision for implement:** use `RemoteAddr`
  for the limiter key to avoid unauthenticated clients spoofing XFF and
  rotating identity; document that edge proxies should rate-limit too.
- Over-limit responses match normal failure: form → redirect `/login`; API →
  403 + generic `login failed` (stop returning raw `err.Error()` from verify).

### 6. HTTP server timeouts and shutdown

```go
ReadHeaderTimeout: 5s
ReadTimeout:       15s   // enough for normal API; uploads need care
WriteTimeout:      60s   // covers large page/image responses
IdleTimeout:       60s
Shutdown:          context with 10s deadline via srv.Shutdown
```

Upload may need a longer read path than generic ReadTimeout. Prefer:

- Keep moderate `ReadTimeout` on the server, and for admin upload rely on
  `MaxBytesReader` so oversized bodies fail fast; if ReadTimeout proves too
  tight for large admin uploads in tests, raise Write/Read only as needed and
  document constants in code.

Replace `srv.Close()` on cancel with `Shutdown(ctx10s)`.

### 7. Body limits

- Login form: `http.MaxBytesReader` ~ 1<<20 before `ParseForm` / `FormValue`.
- API login: `MaxBytesReader` ~ 1<<20 before JSON decode.
- Admin upload: wrap body with `MaxBytesReader` at 32<<20 (match memory
  multipart budget) before `ParseMultipartForm`.

### 8. CORS, CSRF, headers, uploads

| Topic | Decision |
|---|---|
| CORS | Remove `Access-Control-Allow-Origin: *`. Keep Allow-Methods/Headers only if still useful for same-origin preflight; OPTIONS can return 204 without `*`. |
| CSRF | No token framework. `SameSite=Lax` + no credentialed cross-origin CORS is the model; state-changing POSTs from other sites do not send the cookie on cross-site POST in modern browsers for Lax (top-level GET navigations still send cookie — login callback sanitization remains required). |
| Security headers | Middleware sets `X-Content-Type-Options: nosniff`, `X-Frame-Options: SAMEORIGIN`, `Referrer-Policy: strict-origin-when-cross-origin` on responses. |
| Cover upload | Keep JPEG/PNG-only; do not expand SVG admin covers. |
| `/uploads` | Keep abs-path containment; ensure prefix check uses path separator boundary (`absUpload+sep`) to avoid `/uploads_evil` style siblings. |
| SVG | No new SVG upload API; library comic SVG remains out of scope except not widening upload MIME for admin covers. |

### 9. Compatibility

- Cookie name, Path=`BaseURL`, MaxAge, HttpOnly, SameSite unchanged except
  Secure policy above.
- API login JSON shape `{success, session_id, is_admin}` preserved; error
  message normalized to generic text.
- No DB migration; `Storage.Logout` already exists.
- OPDS Basic and bearer tokens unchanged aside from logout revocation of the
  presented token.

## Trade-offs

| Choice | Benefit | Cost |
|---|---|---|
| Smart Secure via X-Forwarded-Proto | Works with TLS-terminating proxies without new config | Spoofable if process is directly exposed; mitigated by docs (same class as proxy auth) |
| RemoteAddr rate-limit key | Cannot forge limit identity via XFF | Shared NAT IPs share budget; reverse proxies should also rate-limit |
| Docs+warning for proxy auth | No deploy break for correct setups | Direct exposure remains dangerous if operators ignore warning |
| Drop CORS `*` | Least privilege for browser | Cross-origin browser JS clients break until a future whitelist task |

## Rollback

- Revert the server package commit(s); no migration to undo.
- Feature is behavioral; operators who relied on CORS `*` or open redirects must
  adjust clients — called out in release notes if shipped.

## Test plan (design-level)

- Unit: safe redirect helper (good paths, `//evil`, `https://`, empty).
- Handler: logout revokes token; API login error is generic; rate limit trips.
- Cookie Secure true/false matrix (TLS nil + X-Forwarded-Proto).
- Middleware: no `Access-Control-Allow-Origin: *`; security headers present.
- Upload: oversize body rejected; path containment sibling prefix.
- Package: `go test ./...`, race, vet, build with GOCACHE in allowed temp dir.

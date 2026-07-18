# React login page design

## Scope

Migrate browser login only:

- `GET /login` → React shell
- submit via existing `POST /api/login`
- safe post-login navigation + callback plumbing
- already-authenticated redirect off `/login`

Do not migrate logout UX redesign, home/library, or other auth modes.

## Routes and pageId

| Browser route | Go handler result | React pageId |
|---|---|---|
| `GET /login` (anonymous) | React shell | `login` |
| `GET /login` (valid session) | 302 to safe callback or BaseURL home | n/a |
| `POST /login` | keep existing form handler | n/a (rollback) |
| `POST /api/login` | existing JSON handler | used by React |
| `GET /logout` | unchanged | n/a |

## Auth and redirect flow

```
requireAuth (page) → /login?callback=<original path+query>
GET /login
  ├─ valid token → 302 safe(callback) or BaseURL home
  └─ React LoginPage
       POST /api/login { username, password }
         ├─ success + Set-Cookie → location = safe(callback) or BaseURL home
         └─ failure → in-page generic error
```

### Callback rules

- Server: reuse / mirror `safeRedirectPath` (relative `/...`, reject `//`, `\`,
  absolute URLs, schemes/hosts).
- Client: same intent before `window.location` assignment; never trust raw
  query alone without sanitization.
- BaseURL: if callback is app-relative under non-root mount, compose with
  BaseURL the same way form login already does.
- Default destination when empty/unsafe: BaseURL home (`baseUrl()` / `appPath("")`).

### Already authenticated

Prefer server-side check in `handleLoginPage` using existing token extract +
storage validate (same sources as middleware). Avoid mounting React only to
bounce.

## JSON contract (existing)

`POST /api/login`

Request:

```json
{ "username": "alice", "password": "secret" }
```

Success:

```json
{ "success": true, "session_id": "<token>", "is_admin": true }
```

Failure / rate limit:

- HTTP 403 + `{ "success": false, "error": "login failed" }`

React must not surface credential-specific messages; show a generic login
failed string (Chinese UI copy consistent with other React pages).

## Frontend

- `LoginPage` — standalone layout (no admin `AppShell` topbar/nav).
  - username / password fields
  - password visibility toggle
  - submit loading / disabled state
  - in-page error banner
  - read `callback` from boot and/or URL search params
- Register `login` in `App.tsx`.
- Styles: page-local CSS using `tokens.css` variables; centered card on full
  viewport. No UIkit / legacy login CSS dependency for the migrated route.
- Use `apiFetch` for POST; navigation via full page assign after cookie set
  (not SPA client router).

### Boot payload (optional)

```json
{
  "baseUrl": "/",
  "pageId": "login",
  "pageName": "login",
  "isAdmin": false,
  "version": "...",
  "callback": "/library"
}
```

If boot omits callback, page may still parse `?callback=` client-side.

## Go changes

- `handleLoginPage` → auth check redirect + `renderReactShell("login", ...)`.
- Pass sanitized callback into boot when present on query.
- `requireAuth` page branch: append `?callback=` with path (and query when
  safe) after BaseURL mount strip, still BaseURL-aware login URL.
- Leave `POST /login` form handler for rollback.
- Prefer focused tests:
  - requireAuth callback on page 302
  - login page redirects when already authenticated
  - safe callback rejection still lands on home after API login is client-tested
    conceptually; keep Go `safeRedirectPath` tests green
  - existing API login tests remain authoritative for cookie + rate limit

## Rollback

- Restore `handleLoginPage` to `renderPage("views/login", ...)`.
- Leave React foundation and other migrated routes intact.
- Form `POST /login` remains available during/after switch.

## Validation

```bash
npm run typecheck
npm run build
make check
make test
make build
```

# React migrate login page

## Goal

Migrate the browser login page from Go template + jQuery/UIkit to the React +
Vite shell, while keeping Go as the only backend, authenticator, cookie issuer,
and embed host.

## Confirmed Facts

- Parent migration strategy remains route-level coexistence: only migrated routes
  render React; other pages keep Go templates.
- Foundation already provides React shell (`react-shell.tmpl`), BaseURL boot,
  dual-theme FOUC markers, `apiFetch`, and pageId routing in `App.tsx`.
- Current routes:
  - `GET /login` — server-rendered `login.tmpl` (no React)
  - `POST /login` — form login: body limit, rate limit, sets cookie, safe
    callback redirect
  - `POST /api/login` — JSON login: body limit, rate limit, sets cookie, returns
    `{ success, session_id, is_admin }` or generic `login failed`
  - `GET /logout` — revokes token, clears cookie, redirects to login
- Login page is public (outside auth middleware). Form failure currently
  redirects back to `/login` with no user-visible error message.
- Callback sanitization already lives in Go `safeRedirectPath`; form POST uses
  it. API login does not redirect.
- `requireAuth` page redirects go to BaseURL-aware `/login` (no callback query
  today).
- Login visuals live in legacy CSS (`login-page` / `comic-login-page` card,
  decorations, dual-theme rules). React currently has no login styles or page.
- Current `GET /login` does not auto-redirect already-authenticated users.
- Work is a child of `07-17-frontend-react-vite`, following the same pattern as
  missing-items / admin-users / tags migrations.

## Decisions

- **Submit path**: React uses existing `POST /api/login` (JSON) with
  `credentials: 'same-origin'`; on success navigate client-side. Keep
  `POST /login` temporarily for rollback / non-React clients.
- **Post-login destination**: default to BaseURL home; if a safe same-app
  relative `callback` is present (query and/or boot), use it after
  client-side sanitization matching Go `safeRedirectPath` intent. Update
  `requireAuth` page redirects to include `?callback=` with the original
  path when practical.
- **Visual fidelity**: simplified React-token centered login card (title,
  fields, in-page error, password visibility toggle, loading submit). Match
  shell dual-theme tokens; do not 1:1 recreate UIkit/halftone decorations.
- **Already authenticated**: if `GET /login` has a valid session token, redirect
  immediately to safe callback or BaseURL home (server-side preferred).

## Requirements

- Keep Go as the only long-running backend and deployable binary.
- Use the existing React shell mount pattern (`renderReactShell` + pageId).
- Preserve public access, auth cookie issuance, rate limits, body limits, and
  generic failure messages for security parity.
- Preserve non-root BaseURL for assets, API calls, and post-login navigation.
- Preserve dual-theme (comic/flat × light/dark) intent without CDN runtime deps.
- Scope is the login page only (not a full auth redesign).
- Leave unmigrated routes on Go templates.
- Route-level rollback must remain possible by restoring the template handler.
- Login submit goes through `POST /api/login`; show in-page errors on failure.
- Successful login navigates to safe callback or BaseURL home.
- Unauthenticated page redirects to login may carry a safe callback query.
- Authenticated visits to `/login` auto-redirect away from the form.
- UI is a clean React login card using existing tokens, not pixel-perfect legacy
  comic decorations.

## Acceptance Criteria

- [ ] `GET /login` is served by the React shell with a dedicated login pageId.
- [ ] User can sign in with username/password from the React page via
      `POST /api/login`.
- [ ] Successful login sets the existing auth cookie and navigates into the app
      (safe callback when present, otherwise BaseURL home).
- [ ] Failed login shows a clear in-page error (no silent re-render loop).
- [ ] Rate-limit / generic failure behavior remains non-enumerating.
- [ ] External / protocol-relative callbacks cannot redirect off-site.
- [ ] `requireAuth` page redirects include a safe `callback` when practical.
- [ ] Already-authenticated `GET /login` redirects to safe callback or home.
- [ ] Root and non-root BaseURL work for the page, assets, and API.
- [ ] Dual-theme markers from the React shell continue to apply on login.
- [ ] Login UI is a self-contained React card (no dependency on legacy
      `login.tmpl` / UIkit login CSS for the migrated route).
- [ ] Unrelated pages remain on Go templates.
- [ ] Route-level rollback remains possible without removing the React foundation.

## Out of Scope

- Migrating home, library, title, reader, tags, users, subscriptions, plugin
  download, or OPDS.
- Full SPA client-side ownership of all routes.
- Auth redesign (OAuth, SSO, password reset, registration).
- Changing rate-limit numbers, cookie names, or token storage model.
- Product changes to `disable_login` / auth-proxy modes beyond compatibility.
- Pixel-perfect recreation of legacy login decorations/halftone backgrounds.
- Cleaning remaining legacy jQuery/LESS assets for unmigrated pages.

# React Vite frontend migration

## Goal

Replace Mango's current Go-template, jQuery, Alpine, and UIkit browser UI with a
React + Vite frontend while keeping the Go server as the only backend, API host,
authenticator, and single-binary deployment target.

## Confirmed Facts

- The current UI is server-rendered Go `html/template` under `go/web/views/`,
  with page scripts under `go/web/public/js/` and styles under
  `go/web/public/css/`.
- Runtime dependencies include jQuery, Alpine, UIkit, Select2, Moment, Font
  Awesome, and dual comic/flat theme CSS.
- Go embeds `go/web/views` and `go/web/public` at compile time; the production
  binary does not depend on a Node runtime.
- Authentication, authorization, `BaseURL` mounting, SQLite, library scanning,
  queue storage, and background tasks already live in Go.
- React shell + dual-theme tokens + BaseURL-aware assets are in place.
- Core authenticated browse → read loop is React: home, library, title, tags,
  login, admin users, missing-items, and reader.

## Task Map (children)

### Done (archived)

| Child | Outcome |
|-------|---------|
| `07-18-frontend-react-foundation` | React/Vite shell, embed, themes, Make/Docker |
| `07-18-frontend-react-missing-items` | `/admin/missing` + real APIs |
| `07-18-frontend-react-admin-users` | `/admin/user` list/edit |
| `07-18-frontend-react-tags` | tags index/detail |
| `07-18-frontend-react-login` | login + safe redirects |
| `07-18-frontend-react-browse` | home / library / title-detail |
| `07-19-frontend-react-reader` | immersive reader + bootstrap API |

### Remaining (follow-up children — not yet created)

| Priority | Proposed child | Routes / surface | Notes |
|----------|----------------|------------------|-------|
| P1 | `frontend-react-admin` | `/admin` | Scan, thumbnails, library health, settings-ish admin home |
| P2 | `frontend-react-subscriptions` | `/admin/subscriptions` + plugin subscription APIs | Queue-coupled |
| P2 | `frontend-react-plugin-download` | `/download/plugins` + plugin search/download APIs | Optional when plugin path set |
| P3 | `frontend-react-opds` | n/a (or keep XML) | OPDS is XML clients, not browser UI; usually leave as-is |
| P3 | `frontend-legacy-retirement` | delete unused tmpl/js/css after full cutover | Only after smoke on all migrated routes |

Disabled download manager stays product-out-of-scope cleanup.

## Requirements (parent, still valid)

- Use React + Vite + TypeScript as the browser UI stack.
- Keep Go as the only long-running backend and final deployable binary.
- Build React assets at dev/release time; embed/serve via Go; no public CDN runtime.
- Support non-root `BaseURL` for routes and static assets.
- Child tasks independently verifiable; parent owns task map + cross-child AC.
- Migrated routes: Go HTML shell + React; unmigrated keep templates.
- Style with local CSS/tokens; no full component library requirement.
- Stable JSON contracts per migrated page.
- Leave disabled download-manager product cleanup separate.

## Cross-Task Acceptance Criteria

### First-wave (met by completed children)

- [x] Foundation shell embedded under root/non-root BaseURL.
- [x] Missing-items pilot against real JSON.
- [x] Browse + reader close authenticated read loop.
- [x] Login, tags, admin users migrated.
- [x] Unmigrated routes still template-render.
- [x] No Node runtime in final binary.
- [x] Comic/flat light/dark in React shell.

### Still open (parent-level)

- [ ] Admin home (`/admin`) migrated or explicitly deferred with docs.
- [ ] Subscriptions + plugin download migrated or deferred.
- [ ] Legacy jQuery/Alpine/UIkit assets retired only after remaining pages done.
- [ ] Parent integration review / archive when remaining scope is deferred or done.

## Out of Scope (parent)

- Rewriting Go backend into Node/Next/etc.
- Full design-system rewrite or public CDN assets.
- Completing every page in one change set.
- Product removal of the disabled download manager (unless a dedicated task).
- Rewriting OPDS XML protocol into React (clients expect OPDS, not SPA).

## Recommended next child

**`frontend-react-admin`** — migrate `/admin` (scan / thumbnail generation /
status surface). Highest remaining browser frequency after the read loop, admin
auth already proven on users/missing-items, and it unblocks day-2 ops without
plugin/queue complexity of subscriptions.

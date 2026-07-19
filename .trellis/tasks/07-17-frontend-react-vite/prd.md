# React Vite frontend migration

## Goal

Replace Mango's current Go-template, jQuery, Alpine, and UIkit browser UI with a
React + Vite frontend while keeping the Go server as the only backend, API host,
authenticator, and single-binary deployment target.

## Confirmed Facts

- React shell + dual-theme tokens + BaseURL-aware assets are in place.
- Core authenticated browse → read loop is React: home, library, title, tags,
  login, admin users, missing-items, reader, and admin home.
- Go remains the only long-running backend; production binary needs no Node.
- Product decision (2026-07-19): **subscriptions** and **plugin download** UI
  will not be migrated; features are not needed. Routes may stay as legacy
  templates or unused until a later product change.

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
| `07-19-frontend-react-admin` | `/admin` ops + AppShell theme controls |

### Explicitly deferred / out of product scope

| Item | Decision |
|------|----------|
| Subscriptions UI (`/admin/subscriptions`) | **Won't migrate** — feature not needed |
| Plugin download UI (`/download/plugins`) | **Won't migrate** — feature not needed |
| Download manager | Already product-disabled; out of scope |
| OPDS | Keep Go XML clients; not a React page |
| Legacy asset retirement | Child `07-19-frontend-legacy-retirement` (in progress / done on branch): disable deferred UIs + delete dead assets |

## Requirements (parent, still valid)

- Use React + Vite + TypeScript as the browser UI stack.
- Keep Go as the only long-running backend and final deployable binary.
- Build React assets at dev/release time; embed/serve via Go; no public CDN runtime.
- Support non-root `BaseURL` for routes and static assets.
- Child tasks independently verifiable; parent owns task map + cross-child AC.
- Migrated routes: Go HTML shell + React; unmigrated keep templates if still mounted.
- Style with local CSS/tokens; no full component library requirement.

## Cross-Task Acceptance Criteria

### First-wave (met)

- [x] Foundation shell embedded under root/non-root BaseURL.
- [x] Missing-items pilot against real JSON.
- [x] Browse + reader close authenticated read loop.
- [x] Login, tags, admin users migrated.
- [x] Admin home migrated (scan / thumbnails / theme in shell).
- [x] No Node runtime in final binary.
- [x] Comic/flat light/dark in React shell.
- [x] Subscriptions + plugin download **deferred by product** (not required).

### Still open (optional)

- [ ] Parent integration review / archive (all required children done).
- [ ] Optional: legacy jQuery/Alpine/UIkit asset retirement when ready.
- [ ] Unmigrated template routes (subscriptions/plugin) either left as-is or
      unlinked from nav (admin home already omits them).

## Out of Scope (parent)

- Rewriting Go backend into Node/Next/etc.
- Full design-system rewrite or public CDN assets.
- Migrating subscriptions, plugin download, OPDS, or download manager.
- Completing every historical template page.

## Recommended next

**Archive this parent** after a short integration note (all required migration
children are done). Optional follow-up only: `frontend-legacy-retirement` if
you want to delete unused tmpl/js later — not required for product use.

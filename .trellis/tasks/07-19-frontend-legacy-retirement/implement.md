# Legacy frontend asset retirement execution plan

## 0. Preconditions

- Parent: 07-17-frontend-react-vite
- Branch: `feature/legacy-retirement`
- Product: no subscriptions / plugin download / download-manager UI

## 1. Unmount deferred product routes

- [x] Remove `/admin/subscriptions`, `/admin/downloads`, `/download/plugins`
- [x] Delete handlers for those pages + form POST user-edit rollback
- [x] Delete corresponding templates and page JS

## 2. Remove migrated-page rollback assets

- [x] Delete reader/admin/home/library/login/title/tags/user/missing templates
- [x] Delete page-only + vendor JS (jquery/alpine/uikit/…)
- [x] Delete legacy css/webfonts
- [x] Remove `handleHomeLegacy`, `handleTitleLegacy`, dead layout helpers
- [x] Trim `web.go` types to React shell + library/tag helpers + OPDS-free structs

## 3. Shared chrome

- [x] No remaining `renderLayout` consumers; deleted head/top/bottom/vendor stacks
- [x] Kept `react-shell.tmpl` + `public/react/**` + PWA icons + favicon/manifest
- [x] Updated static path allowlist (`/img`, `/react`)

## 4. Validation

- [x] `npm run typecheck` && `npm run build`
- [x] `go test ./...` (246 passed)
- [x] Spot inventory documented below

## Remaining surface

```
go/web/views/react-shell.tmpl
go/web/public/react/**
go/web/public/img/icons/**
go/web/public/favicon.ico|manifest.json|robots.txt
OPDS handlers (inline XML)
```

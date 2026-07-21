# Frontend Development Guide

Go embeds templates and static assets from `go/web/` (see `go/web/embed.go`).
The browser UI is migrating to **React + Vite + TypeScript**. Node is a
build-time prerequisite only; the Mango binary has no Node/npm or CDN runtime
dependency.

## Architecture

```text
Migrated route  → Go HTML shell (react-shell.tmpl) + React bundle under public/react/
Unmigrated route → existing Go templates + legacy public JS/CSS
```

Both paths are served by the same Go server with the same auth and BaseURL mount.

## Key folders

| Folder | Purpose |
|---|---|
| **`frontend/`** | React + TypeScript source (Vite app) |
| **`go/web/public/react/`** | Generated React production assets (from `npm run build`) |
| **`go/web/views/`** | Go HTML templates, including `react-shell.tmpl` |
| **`go/web/public/css/`**, **`js/`**, **`webfonts/`** | Legacy template-era assets still used by unmigrated pages |

## Stack

| Layer | Technology |
|---|---|
| **Migrated UI** | React 19 + Vite 6 + TypeScript |
| **Shell** | Go `html/template` injects BaseURL / pageId JSON boot |
| **Legacy pages** | Go templates + jQuery / Alpine / UIkit (until migrated) |
| **Embed** | `//go:embed public/*` |

## Quick start

```bash
npm ci
npm run build          # typecheck + Vite build + output check
make run               # rebuild React assets, then go run
```

Developer commands:

```bash
npm run dev            # Vite HMR; proxies /api and /img → http://127.0.0.1:9000
# other terminal: npm run server   # Go API (default :9000)
npm run typecheck
npm run build
npm run check          # fails if go/web/public/react/assets/main.{js,css} missing
make check             # frontend check + go vet
make test
make build
```

`vite.config.ts` `server.proxy` is **dev-only** and does not affect `npm run build` / embed.

### Vite multi-page dev (no Go HTML shell)

When `#mango-boot` is missing, React infers `pageId` from the URL path
(`frontend/src/lib/boot.ts` → `bootFromPathname`). Production always has Go
inject `mango-boot`, so this path is unused after build/embed.

Examples (Vite port, typically :5173):

| URL | pageId |
|-----|--------|
| `/` | home |
| `/library` | library |
| `/book/:id` | title-detail |
| `/tags`, `/tags/:tag` | tags-index / tag-detail |
| `/reader/:tid/:eid[/:page]` | reader |
| `/admin`, `/admin/user`, … | admin / user-list / … |
| `/login` | login |

Still run Go for APIs (`npm run server`). Auth cookies are issued for Go’s port;
if login fails under pure Vite, use `make run` or open API-backed flows on :9000.

## Migrated route contract

1. Admin (or auth) middleware still runs in Go.
2. Handler renders `views/react-shell` with `ReactShellData`:
   - `BaseURL`
   - `PageName`
   - `BootJSON` → `<script type="application/json" id="mango-boot">…</script>`
3. Shell loads `{{.BaseURL}}react/assets/main.css` and `main.js`.
4. React reads boot config and mounts the page for `pageId`.

## Theme

Comic/flat and light/dark markers are applied on `<html>` before paint in
`react-shell.tmpl`, using the same `localStorage` keys as legacy `head.tmpl`
(`ui-style`, `theme`). React CSS tokens live under `frontend/src/styles/`.

## Coexistence rules

- Only routes that explicitly render `react-shell` use React.
- Unmigrated routes keep their templates and legacy scripts.
- Navigation may link both directions.
- Do not delete legacy public assets until their last template consumer is gone.

## Docker / embed order

```text
npm ci → npm run build → go build → go:embed public/*
```

The production Dockerfile uses a Node stage for the Vite build, then copies
`go/web/public` into the Go builder before `go build`.

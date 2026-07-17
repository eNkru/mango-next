# React Vite foundation design

## Scope

Establish the React + Vite + TypeScript toolchain and a Go-served HTML shell that
can mount future migrated pages. No real business page is switched yet, but a
placeholder migrated route proves the mount path.

## Directory layout

```text
frontend/                     # React source (TypeScript)
  index.html
  src/
    main.tsx
    App.tsx
    pages/PlaceholderPage.tsx
    shell/                    # layout, alerts, confirm, states
    lib/baseUrl.ts
    lib/api.ts
    styles/tokens.css
    styles/shell.css
  vite.config.ts
  tsconfig.json

go/web/views/react-shell.tmpl # Go HTML shell for migrated routes
go/web/public/react/          # Vite build output (generated, committed or rebuilt)
```

Exact folder names may adjust during implementation, but React source stays
outside legacy jQuery page scripts, and generated outputs remain under
`go/web/public/` so `go:embed public/*` continues to work.

## Build contract

- Root `package.json` becomes the React/Vite/TypeScript app manifest.
- Scripts:
  - `npm run dev` — Vite dev server (optional local DX)
  - `npm run build` — production assets into `go/web/public/react/`
  - `npm run typecheck` — `tsc --noEmit`
  - `npm run assets:check` or `npm run check` — ensure declared generated outputs exist / match build
- Remove or stop relying on the old jQuery/LESS `assets:copy` inventory as the
  primary architecture. Legacy template assets may remain in place for
  unmigrated pages.
- Make: `build`, `static`, `run` depend on React production build.
- Docker: Node stage runs `npm ci` + `npm run build`, then Go stage embeds
  `go/web/public`.

## Go shell contract

For a placeholder migrated admin route (for example `/admin/react-preview`):

1. Admin middleware still applies.
2. Handler renders `react-shell.tmpl` with:
   - `BaseURL`
   - `PageName` / `PageID` (e.g. `react-preview`)
   - optional JSON boot config in a `<script type="application/json" id="mango-boot">`
3. Template loads only the React CSS/JS outputs with BaseURL prefixes.
4. Body contains `<div id="root"></div>`.
5. FOUC script sets comic/flat + light/dark markers on `html` before paint using
   the same localStorage keys as the legacy UI.

Unmigrated routes keep existing templates and assets.

## React shell primitives

Minimum shared primitives for the pilot:

- `AppShell` layout frame compatible with admin subpage chrome
- page heading/subtitle slot
- alert/toast helper
- confirm dialog
- loading / empty / error panels
- `baseUrl()` helper and `apiFetch()` that prefixes BaseURL and handles JSON
  envelopes / non-2xx errors

Placeholder page renders a simple admin-style panel proving theme classes,
BaseURL display, and shell primitives.

## Theme and assets

- Port comic/flat tokens into React CSS variables.
- Keep local Bangers/Fredoka fonts if comic theme requires them.
- No CDN runtime dependencies.
- No full component library.

## Rollback

- Delete or unroute the placeholder handler without affecting template pages.
- React build tooling can remain even if no production page uses it yet.

## Validation

```bash
npm ci
npm run typecheck
npm run build
make check
make test
make build
docker build -t mango-react-foundation .
```

Also verify:

- placeholder route returns React shell HTML containing the root mount and asset tags
- root and `/mango/` BaseURL asset prefixes
- unmigrated `/library` or `/login` still render templates
- no `fonts.googleapis.com` in React shell HTML/CSS/JS

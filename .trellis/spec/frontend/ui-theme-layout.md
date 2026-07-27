# UI Theme Layout Contracts (Comic / Flat)

## Scope / Trigger

Dual UI styles share shell DOM but skin independently. Apply when changing top bar, sidebar, home rails, library cards, or theme CSS.

## Dual shell

| Style | Chrome | Body/html markers | Sidebar | Top bar (`.flat-topbar`) |
|-------|--------|-------------------|---------|---------------------------|
| `flat` | Netflix browse | `flat-theme` / `flat-theme-dark` | hidden | shown (desktop) |
| `comic` | Same geometry + comic skin | `comic-theme` / `comic-theme-dark` | hidden | shown (desktop) + comic skin |

DOM lives in `go/web/views/top.tmpl` (sidebar + topbar both present). Visibility is CSS-only so `setUIStyle` stays client-side.

## Visibility rules

- **Hide topbar only when neither theme owns chrome**:
  - `body:not(.flat-theme):not(.comic-theme):not(.comic-theme-dark) .flat-topbar { display: none }`
- **Do not** use `body:not(.flat-theme) .flat-topbar { display: none }` — that blocks comic top bar.
- Comic layout-critical rules also use `html.comic-theme` / `html.comic-theme-dark` (FOUC: head script marks html before body).

## Geometry

Mirror Flat for both styles:

- Top bar height: `@flat-topbar-height` (68px)
- `app-content`: `margin-left: 0`, `padding-top: calc(68px + 16px)`, horizontal `4vw` (mobile: 72px / 16px)
- Poster media: `aspect-ratio: 2 / 3`; library titles reserve 2-line height for equal cards

## React home: continue-reading (stacked deck)

Home continue-reading is **not** a poster rail (unlike start-reading / recently-added).

| Piece | Classes | Behavior |
|-------|---------|----------|
| Shell | `.mango-continue-stack` | Stacked deck under section heading |
| Stage | `.mango-continue-stack__stage` | Positioning context; pads for offset peeks |
| Card | `.mango-continue-stack__card` (+ `--active` / `--back`) | Active front large; backs offset + scale behind |
| Face | `.mango-continue-stack__face` | Active: cover + meta grid |
| Back | `.mango-continue-stack__back` | Cover-only button; click brings card to front |
| Meta | `.mango-continue-stack__meta` | Active only: title + page + progress + **Continue** → reader |
| Arrows | `.mango-continue-stack__arrow` | Prev/next when `length > 1`; desktop ≥768px only |

Rules:

- Initial active index is `0` (API order)
- Inactive cards sit **behind** active: **previous** stack left, **next** stack right (`--stack-side` ±1, `--stack-depth`)
- **Circular**: index wraps; shortest path picks left/right so both sides stay populated when `length > 1`
- Arrows always shown when multi (wrap forever); click back card or arrows → active front; **does not** open reader
- Reader entry only via **Continue** on the active card
- No horizontal scroll track / show more
- Single item: no arrows (`.mango-continue-stack--single`)
- Mobile: hide arrows; smaller stack shift; same bring-to-front click
- Cap visible backs (~4 deep per side) for layout sanity
- Do **not** use `PosterCard` / `.mango-poster-rail` for continue
- Comic: thick border / hard shadow on stack cards
- `prefers-reduced-motion: reduce` disables transform transitions
- Source: `frontend/src/browse/ContinueCarousel.tsx`, styles in `frontend/src/styles/shell.css`

## Skin isolation

- Flat tokens / Netflix red: only under `flat-theme*`
- Comic palette (`@comic-red`, paper, thick borders): only under `comic-theme*`
- Comic: **sharp corners** — `border-radius: 0` for all surfaces under comic (cards, buttons, modals, FAB, inputs)
- Comic motion: medium only (soft scale/fade; no strong rotate card pop)

## Files

```
go/web/public/css/flat-theme.less|.css   # flat skin + shared topbar show for flat
go/web/public/css/comic-theme.less|.css  # comic shell + skin + sharp corners
go/web/public/css/_variables.less        # tokens (do not mix palettes)
go/web/views/top.tmpl                    # shared chrome DOM
go/web/public/js/common.js               # setUIStyle / setTheme class toggles
```

## React AppShell theme controls

- Runtime toggles live in `frontend/src/shell/AppShell.tsx` via
  `frontend/src/lib/theme.ts` (not only on admin home).
- Keys: `localStorage.theme` = `dark|light|system`,
  `localStorage['ui-style']` = `comic|flat`.
- `applyHtmlTheme` must clear all four markers then add the active pair,
  matching `react-shell.tmpl` FOUC script. Never leave comic + flat together.
- When `theme=system`, subscribe to `prefers-color-scheme` changes and re-apply.

Migrated React routes boot comic/flat + light/dark markers on `<html>` from
`go/web/views/react-shell.tmpl` (same `localStorage` keys as legacy
`head.tmpl`: `ui-style`, `theme`). Legacy LESS/CSS pages still use
`go/web/public/css/*`. React tokens live under `frontend/src/styles/`.

Build migrated assets with `npm run build` (Vite → `go/web/public/react/`).
Asset filenames are **content-hashed** and resolved at runtime from the Vite
manifest — see "Build / cache-busting contract" below.

## Build / cache-busting contract (cross-layer)

### 1. Scope / Trigger
Apply when changing `vite.config.ts` output names, the Go HTML shell asset
URLs, `scripts/check-react-outputs.mjs`, or anything that reads the Vite
manifest. This is a cross-layer contract: Vite emits → Go embeds → Go resolves
at boot → template injects. Breaking any link serves stale or 404 assets.

### 2. Signatures
- **Vite** (`vite.config.ts`): `build.manifest = true`; output names
  `entryFileNames/chunkFileNames: 'assets/[name]-[hash].js'`,
  `assetFileNames: 'assets/[name]-[hash][extname]'`. A `copyManifestToOutDir`
  plugin copies `.vite/manifest.json` → outDir `manifest.json` in `closeBundle`.
- **Go** (`go/internal/server/web.go`):
  `func loadReactAssets(publicFS fs.FS) reactAssets` — reads
  `react/manifest.json` from the embed FS, returns `{mainJS, mainCSS string}`
  (paths relative to `/react/`, e.g. `assets/index-<hash>.js`).
- **Go** (`ReactShellData`): fields `MainJS string`, `MainCSS string`.
- **Template** (`go/web/views/react-shell.tmpl`):
  `{{if .MainCSS}}<link rel="stylesheet" href="{{.BaseURL}}react/{{.MainCSS}}">{{end}}`
  and `{{if .MainJS}}<script type="module" src="{{.BaseURL}}react/{{.MainJS}}">{{end}}`.
- **Check** (`scripts/check-react-outputs.mjs`): verifies `manifest.json`
  exists and the entry `file` + `css[]` it references are on disk (no fixed
  filenames).

### 3. Contracts
- Manifest entry key is `"index.html"`; `file` = entry JS, `css` = array
  (take `[0]`). Other keys are dynamic chunks (`src/pages/*.tsx`) — not read
  by Go; the browser fetches them via the hashed URLs Vite writes into the
  entry JS.
- **Embed path**: `web/embed.go` does `//go:embed public/*`; `Public()` subs
  to `public/`. So `go/web/public/react/manifest.json` on disk is
  `react/manifest.json` in the embed FS. **`go:embed` skips dotfile dirs**
  (`.vite/`) — that is why the copy plugin moves the manifest to a clean path.
- **Empty-string fallback**: missing/malformed manifest → `loadReactAssets`
  returns zero-value `reactAssets` → template `{{if .MainJS}}`/`{{if .MainCSS}}`
  skip the tags → shell boots with no assets (blank page). Surfaces at boot as
  a `slog.Error("react manifest missing; run npm run build")`, not a panic.

### 4. Validation & Error Matrix
| Condition | Behavior |
|-----------|----------|
| `npm run build` not run before `go build` | manifest absent → empty MainJS/MainCSS → blank shell; check script fails the build |
| Manifest present but no `index.html` entry | `slog.Error("react manifest has no index.html entry")`; empty assets |
| Manifest JSON malformed | `slog.Error("react manifest parse")`; empty assets |
| Content changed, rebuild run | new `[hash]` → new URL → browser fetches fresh (no stale cache) |

### 5. Good/Base/Bad Cases
- **Good**: change a page → `npm run build` → new `index-<newhash>.js` → Go
  serves the new URL → no stale bundle possible.
- **Base**: rebuild with no content change → same hash → same URL → cache hit
  (immutable, safe).
- **Bad**: hardcode `assets/main.js` in the template (old behavior) → after
  upgrade the URL is identical → browser/CDN serves the **stale** old bundle.

### 6. Tests Required
- `scripts/check-react-outputs.mjs` after every build: manifest exists + entry
  JS/CSS files on disk. Run via `npm run build` / `npm run check`.
- `go vet ./internal/server/` — catches `ReactShellData` field / struct
  mismatches between `web.go`, `handlers_pages.go`, and the template.
- Manual: after a rebuild, the served `main.js`/`main.css` URLs carry a content
  hash and change when content changes.

### 7. Wrong vs Correct
#### Wrong
- Hardcode `assets/main.js` / `assets/main.css` in `react-shell.tmpl`.
- Force fixed Vite output names (`entryFileNames: 'assets/main.js'`).
- Read the manifest from `.vite/manifest.json` in Go (dotfile dir, not embedded).
- Assert fixed filenames in `check-react-outputs.mjs`.
#### Correct
- Hashed output names + `manifest: true` + copy plugin to clean path.
- Go reads `react/manifest.json` from the embed FS at boot.
- Template uses `{{.MainJS}}`/`{{.MainCSS}}` guarded by `{{if}}`.
- Check script globs via the manifest, not exact filenames.

## Route code-splitting (page-load)

`frontend/src/App.tsx` lazy-loads heavy, route-gated pages with `React.lazy`
behind a single `<Suspense fallback={<RouteFallback/>}>` (`RouteFallback` =
`LoadingState` + `t('loading')`). A login/home/library user no longer
downloads the reader or admin code.

- **Lazy** (named-export adapter required — every page uses `export function X`):
  `TitleDetailPage`, `TagDetailPage`, `TagsIndexPage`, `AdminPage`,
  `UserListPage`, `UserEditPage`, `MissingItemsPage`, `ReaderPage`. Adapter:
  `lazy(() => import('./pages/X').then((m) => ({ default: m.X })))`.
- **Eager** (keep static — first-paint-critical or the fallback):
  `HomePage`, `LoginPage`, `LibraryPage`, `AppShell`, `ErrorState`,
  `LoadingState` (the Suspense fallback — must NOT be lazy), `UnknownPage`.
- Vite emits each lazy page as its own `assets/<Name>-<hash>.js` chunk; the
  reader lands in `ReaderPage-<hash>.js` and is **absent from `main.js``.
- Do **not** re-eager-import a lazy page from another module (defeats the split).

## `<html lang>` before first paint (a11y)

`react-shell.tmpl` and `frontend/index.html` each carry an inline `<head>`
script that sets `documentElement.lang` from `localStorage['mango-language']`
**before paint**, mirroring `applyDocumentLanguage` in `frontend/src/lib/i18n.tsx`
(`en→en`, `zh-tw→zh-Hant`, else `zh-Hans`). The module-load call in `i18n.tsx`
remains the source of truth after hydration; the inline script only fixes the
pre-hydration paint. **Keep the two mappings in sync** — both files carry a
comment pointing at the other. Screen readers get the right language on first
paint for `en`/`zh-tw` users instead of always `zh-CN`.

## Mobile breakpoints (responsive shell)

- **Reader top bar** (`.mango-reader-topbar`): `flex-wrap: wrap`; at
  `max-width: 560px` the center title takes `flex: 1 1 100%; order: 3` and the
  visible labels on the controls/exit buttons hide via
  `.mango-reader-topbar__label--sm-hide { display: none }`. Both buttons keep an
  `aria-label` (exit is an `AppLink`, which forwards `aria-label` via `{...rest}`)
  so the accessible name persists when text drops.
- **App topbar**: at `max-width: 820px` `.mango-topbar__tools` is `flex-wrap: nowrap`
  (keep the 6-button cluster on one row → ≤2-row topbar); at `max-width: 560px`
  topbar padding/gap tighten.
- **Touch targets**: `@media (pointer: coarse)` raises `.mango-btn--icon` to
  `min-width/height: 2.75rem` (44px) and `.mango-tag-pill .mango-btn--icon` to
  `2.5rem`. Use `pointer: coarse` (not a px breakpoint) so mouse/trackpad users
  keep the compact 2.25rem density.
- **Sticky-topbar blur**: `@media (pointer: coarse), (max-width: 560px)` drops
  `backdrop-filter` on `.mango-topbar` with an opaque `--mango-bg-surface`
  fallback (avoids mobile scroll jank).

## React design tokens (`frontend/src/styles/tokens.css`)

| Token group | Purpose |
|-------------|---------|
| `--mango-accent*` / surfaces / text | Theme skins (flat/comic × light/dark) |
| `--mango-danger` / `--mango-danger-hover` | Destructive buttons (not accent red) |
| `--mango-success` | Success alert border |
| `--mango-on-accent` | Primary button label color |
| `--mango-ink` | Comic thick borders / offset shadows |
| `--mango-font-body` | Flat / default UI sans stack |
| `--mango-font-comic` | Comic UI stack (Fredoka + system CJK) |
| `--mango-reader-*` | Immersive reader chrome (fixed dark; **not** theme-switched) |

### Fonts (React shell)

| Style | Token | Loading |
|-------|-------|---------|
| Flat | `--mango-font-body` (`Segoe UI` / Helvetica / Arial) | System only; unchanged by comic work |
| Comic | `--mango-font-comic` | Body under `html.comic-theme` / `html.comic-theme-dark` |

Comic stack (Latin first, then system CJK — **no** full Noto CJK binaries in repo):

```css
"Fredoka",
"Noto Sans CJK SC", "Noto Sans CJK TC",
"Noto Sans SC", "Noto Sans TC",
"PingFang SC", "Hiragino Sans GB", "Microsoft YaHei",
"Segoe UI", sans-serif
```

- **Fredoka**: self-hosted WOFF2 400 + 700 via `@font-face` in
  `frontend/src/styles/fonts.css` → `frontend/src/assets/fonts/fredoka/`
  (Vite packs into `go/web/public/react/assets/`). SIL OFL; see `OFL.txt`.
- Only faces **400** and **700** are shipped. `font-weight: 800` (e.g. brand)
  synthesizes from 700 — acceptable; do not add extra faces unless needed.
- **Do not** rely on runtime Google Fonts CDN for comic UI.
- **Do not** use heading-only comic overrides (brand / page-header / login);
  comic body inherits the same token for all AppShell chrome.
- **Reader**: `.mango-reader` keeps `font-family: var(--mango-font-body)`;
  never force comic display on immersive chrome.

### Buttons (React shell)

- `.mango-btn--danger` uses danger tokens, not accent.
- Comic: `.mango-btn` gets thick border + offset shadow (`--mango-ink`).
- Ghost (reader): border uses `--mango-reader-ghost-border`.
- Icons: `.mango-btn` uses `inline-flex` + gap; `.mango-btn--icon` for compact
  icon-only controls. Icons use `currentColor` so comic/flat contrast inherits.

### Brand mark

- Topbar / reader brand: `baseUrl('img/icons/mango-mark.svg')` + “Mango” text
  (`.mango-topbar__mark` ~24–28px). Mark is decorative when text is present.

### Language control

- Shared `LanguageSelect` in AppShell, Login, Reader top bar.
- Key: `localStorage['mango-language']`.

### Scaffold removed

- Do **not** reintroduce `GET /admin/react-preview` / `PlaceholderPage` — foundation
  playground was deleted after migration.

## Common mistakes

| Wrong | Correct |
|-------|---------|
| Hide topbar with `:not(.flat-theme)` only | Exclude comic markers too |
| Comic side rail uses aspect-ratio only | Full-height media where row layout requires it |
| Library card height follows title wrap | Fixed 2-line title slot + stretch grid |
| Change Flat accent when restyling comic | Scope comic only |
| Continue-reading uses poster rail like start/recent | Stacked deck (active front + offset backs) |
| TagDetail invents progress/modified sort | `BrowseToolbar modes={['natural','title']}` + `showProgress={false}` |
| Primary/danger both use accent red | Danger uses `--mango-danger*` |
| Re-add react-preview for “component playground” | Use real pages or a local story; route removed |
| Comic only on brand / h1 (heading split) | `html.comic-theme body` → `--mango-font-comic` for all UI |
| Reader inherits comic body font | `.mango-reader { font-family: var(--mango-font-body) }` |
| Vendor full Noto CJK / runtime GF CDN for comic | Fredoka WOFF2 self-host + system CJK stack only |

## Smoke checklist

- [ ] comic dark/light: top bar, no sidebar, full-width
- [ ] flat dark/light: unchanged Netflix chrome
- [ ] toggle ui-style: class mutual exclusion
- [ ] home continue-reading: stacked deck (not poster rail); back click brings to front; Continue on active only
- [ ] library cards equal height, sharp corners
- [ ] comic buttons: thick border + shadow; danger distinct from accent
- [ ] Login: language select works before session
- [ ] TagDetail: PosterCard grid, no empty progress bars
- [ ] Reader chrome: dark immersive; ghost/primary readable
- [ ] Topbar: mark + nav icons + logout; comic sharp corners still apply
- [ ] Icon buttons: spacing/contrast OK under flat and comic light/dark
- [ ] comic body/nav/buttons share Fredoka + CJK stack; flat body unchanged
- [ ] Reader chrome still `--mango-font-body` (not comic display)
- [ ] build emits Fredoka `*.woff2` under `go/web/public/react/assets/`

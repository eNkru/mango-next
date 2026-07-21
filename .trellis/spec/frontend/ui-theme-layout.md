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

## React home: continue-reading (hero + list)

Home continue-reading is **not** a poster rail (unlike start-reading / recently-added).

| Piece | Classes | Behavior |
|-------|---------|----------|
| Stack | `.mango-continue` | Vertical stack under section heading |
| Hero | `.mango-continue-hero` | `items[0]`: cover + title + page text + thicker progress + single **Continue** → reader |
| List | `.mango-continue-list` / `.mango-continue-row` | `items[1..]`: compact row (thumb + title + thin progress); whole row → reader |
| Expand | `.mango-continue-more` | When rest length > 3, toggle show more / less (`LIST_PREVIEW = 3`) |

Rules:

- Only one continue item → hero only; do not render empty list
- Do **not** use `PosterCard` / `.mango-poster-rail` for continue
- No secondary Open/book-detail button on hero in this layout
- Comic: sharp corners on `.mango-continue-hero` and `.mango-continue-row` (same thick border/shadow pattern as other cards)
- Source: `frontend/src/pages/HomePage.tsx` (`ContinueSection`), styles in `frontend/src/styles/shell.css`

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
| Continue-reading uses poster rail like start/recent | Hero + compact list (see above) |
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
- [ ] home continue-reading: hero + list (not poster rail); expand when >3 secondary rows
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

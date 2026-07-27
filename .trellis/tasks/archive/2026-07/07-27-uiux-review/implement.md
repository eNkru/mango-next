# Implement — UI/UX polish: ease-of-use, mobile, page-load

Ordered execution plan for the requirements in `prd.md` and the design in
`design.md`. Tiered: Tier 1 = required acceptance criteria; Tier 2 = strong
polish, do if time permits; Tier 3 = minor / deferred. Run validation after
each tier.

## Validation commands (run after each tier)

```bash
npm run typecheck          # tsc --noEmit
npm run build              # tsc + vite build + check-react-outputs.mjs
npm run check              # check-react-outputs.mjs alone
```
Manual smoke (after Tier 1 + before reporting done):
- Four theme combos: flat×light, flat×dark, comic×light, comic×dark.
- Three languages: en, zh-cn, zh-tw (verify `<html lang>` on first paint).
- Widths: 360 / 414 / 768 / desktop.
- Flows: login (incl. language switch before session) → home → library grid
  → title-detail → reader (continuous + paged) → exit. Admin scan/thumb
  cards. Missing-items table.

## Pre-flight (before editing)

- [ ] Read `scripts/check-react-outputs.mjs` to learn what it asserts about
      asset paths. **Do this before touching `vite.config.ts` output names**
      — if it hardcodes `assets/main.js`/`assets/main.css`, the cache-busting
      choice (D4) must keep those names or update the script.
- [ ] Confirm each page's export style (named `export function X` vs default)
      so the `React.lazy` adapter is correct. Known: `HomePage`, `LoginPage`,
      `LibraryPage` use `export function` → need `.then(m => ({ default: m.X }))`.

## Tier 1 — required acceptance criteria

### T1.1 Route-level code-splitting (D1)
- [ ] In `frontend/src/App.tsx`: convert heavy route pages to `React.lazy`
      with the named-export adapter: `TitleDetailPage`, `TagDetailPage`,
      `TagsIndexPage`, `AdminPage`, `UserListPage`, `UserEditPage`,
      `MissingItemsPage`, `ReaderPage`.
- [ ] Keep eager: `HomePage`, `LoginPage`, `LibraryPage`, `AppShell`,
      `ErrorState`, `LoadingState`, `UnknownPage`.
- [ ] Add a `RouteFallback` using `LoadingState` and wrap `<Routes>` in
      `<Suspense fallback={<RouteFallback />}>`.
- [ ] `npm run build`; confirm the reader (and admin) land in separate
      chunks under `go/web/public/react/assets/` and are **absent** from
      `main.js` (grep the built `main.js` for a reader-only symbol, e.g.
      `readerPageImagePath` or `useFocusTrap` — should be gone).
- [ ] Verify login route network payload excludes reader code (build output
      chunk list / source map).

### T1.2 Mobile reader top bar (D2.1)
- [ ] `shell.css`: add `flex-wrap: wrap` + a `@media (max-width: 560px)`
      block for `.mango-reader-topbar` (gap, padding, center full-width/order).
- [ ] `ReaderTopBar.tsx`: wrap the visible labels on the controls button and
      exit button in `<span className="mango-reader-topbar__label
      mango-reader-topbar__label--sm-hide">`; add the `--sm-hide` CSS rule.
      Ensure both buttons keep an `aria-label` (exit has one; add to controls).
- [ ] Smoke: 360px reader top bar, no horizontal scroll, title legible or
      hidden with progress in controls modal.

### T1.3 App topbar mobile height (D2.2)
- [ ] `shell.css` `@media (max-width: 820px)`: `.mango-topbar__tools {
      flex-wrap: nowrap; }`.
- [ ] `shell.css` `@media (max-width: 560px)`: tighten `.mango-topbar`
      padding/gap.
- [ ] Smoke: 360px topbar ≤2 visual rows, all controls reachable.

### T1.4 Touch targets ≥44px (D2.3)
- [ ] `shell.css`: add `@media (pointer: coarse)` block raising
      `.mango-btn--icon` to `2.75rem` and `.mango-tag-pill .mango-btn--icon`
      to `2.5rem`.
- [ ] Smoke on a touch device / DevTools touch mode: icon buttons and
      tag-remove are comfortably tappable; desktop (mouse) density unchanged.

### T1.5 `<html lang>` before first paint (D3)
- [ ] `go/web/views/react-shell.tmpl`: add inline `<script>` in `<head>`
      (after the theme FOUC script) that reads `localStorage['mango-language']`
      and sets `documentElement.lang` via the same mapping as
      `applyDocumentLanguage` (`en→en`, `zh-tw→zh-Hant`, else `zh-Hans`).
- [ ] `frontend/index.html`: add the same inline script (dev parity).
- [ ] Add a sync-comment in both files pointing at `i18n.tsx:applyDocumentLanguage`.
- [ ] Smoke: with `localStorage['mango-language']='en'`, hard-reload `/login`/
      `/` and inspect `<html lang>` in DevTools **before**/at first paint —
      must be `en`, not `zh-CN`. Repeat for `zh-tw` (→ `zh-Hant`).

### T1.6 Cache-busting (D4) — **Option A chosen** (content-hash + Go manifest)
- [ ] Update `vite.config.ts`: hashed `entryFileNames: 'assets/[name]-[hash].js'`,
      `chunkFileNames: 'assets/[name]-[hash].js'`, `assetFileNames:
      'assets/[name]-[hash][extname]'`. Drop the CSS-forcing branch so CSS
      becomes `assets/main-[hash].css`. Enable `build.manifest = true` so Vite
      writes `manifest.json` (entry → hashed filename) into the outDir.
- [ ] **Update `check-react-outputs.mjs`** to match the new hashed/globbed
      paths (it currently asserts fixed `assets/main.js` / `assets/main.css`).
      Use `fs.glob`/regex over `go/web/public/react/assets/` instead of exact
      filenames. Run pre-flight first to learn what it asserts.
- [ ] Go side: read `manifest.json` (embed it at build, or read from the
      served `react/` dir at startup) and inject the hashed entry URLs into
      `react-shell.tmpl` via template fields (e.g. `{{.MainJS}}` /
      `{{.MainCSS}}`) instead of hardcoded `assets/main.js` / `assets/main.css`.
      Locate the serve handler + template render in `go/web`.
- [ ] Update `react-shell.tmpl`: replace the two hardcoded asset URLs with the
      template-injected hashed paths.
- [ ] Verify: after a rebuild, the served `main.js`/`main.css` URLs carry a
      content hash and change when content changes; `npm run build` (incl.
      `check-react-outputs.mjs`) passes; the Go server serves the right files.

### T1.7 Reduced-motion / blur (D2.4)
- [ ] `shell.css`: add `@media (pointer: coarse), (max-width: 560px)` to
      drop `backdrop-filter` on `.mango-topbar` with an opaque fallback.
- [ ] Smoke: scroll a long library page on a coarse-pointer/small viewport;
      confirm no blur jank; topbar stays opaque and readable in all themes.

### T1.8 Final Tier 1 validation
- [ ] `npm run typecheck` passes.
- [ ] `npm run build` passes (incl. `check-react-outputs.mjs`).
- [ ] No new inline styles except dynamic width/margin (grep `style={{`).
- [ ] Manual smoke matrix above passes; no regression to acknowledged
      strengths (focus-visible, reduced-motion, skeletons, lazy poster imgs,
      semantic header/main, ARIA menus, i18n, four theme combos).

## Tier 2 — strong polish (do if time permits)

- [ ] **Route-bundle prefetch** (D1 follow-up): add `<link rel="modulepreload">`
      for the likely-next chunk (Home → TitleDetail; TitleDetail → Reader).
      Needs Vite to expose chunk URLs (manifest or post-build read). If
      non-trivial, record as a follow-up task instead.
- [ ] **Fredoka preload** (D5): add `<link rel="preload" as="font"
      crossorigin>` for the two woff2 in `react-shell.tmpl` (under Option B
      fixed names). Confirm comic first paint is faster / no FOUT regression.
- [ ] **Skip-to-content link**: add a visually-hidden-until-focus "Skip to
      content" link at the top of `AppShell` targeting `<main id="...">`.
      Add `id` to `<main>`.
- [ ] **Route-change focus management**: on client-side navigation, move
      focus to the page `<h1>` (or `<main>`) so keyboard/SR users keep
      orientation. Implement in `AppLink`/`App.tsx` via a location-change
      effect.
- [ ] **Title-detail density**: collapse the 4 entry-card action buttons
      into a primary "Continue/Begin" + an overflow ("⋯") menu for
      Download / Mark read / Edit (admin). Reduces mobile clutter.

## Tier 3 — minor / deferred

- [ ] **Quick-jump select respects saved progress**: `TitleDetailPage.tsx:94`
      navigates to `.../1`; use the entry's saved `page` if > 0.
- [ ] **Missing-items table → mobile cards**: at ≤560px, render rows as
      stacked cards instead of a horizontal-scroll table.
- [ ] **CSS split** (low priority): `shell.css` is one 1825-line file. Under
      the single-`main.css` Go-embed model it's not easily splittable; leave
      unless bundle size becomes a concern.

## Review gates

1. **After pre-flight** (T1.0): confirm export styles + `check-react-outputs.mjs`
   behavior before any code edit.
2. **Before T1.6**: confirm Option A vs B with Howard (touches Go scope).
3. **After Tier 1** (T1.8): full validation + manual smoke before reporting
   done. Tier 2/3 are reported as "available, not done" unless explicitly
   picked up.

## Rollback points

- After T1.1 (code-split): revert `App.tsx` → eager imports restore.
- After T1.2–T1.5/T1.7 (CSS/TSX/template): revert individual files.
- After T1.6 (cache-busting): Option B is header-only (revert handler);
  Option A touches vite+template+script+Go (revert the set).
- Each tier is independently revertible; commit per tier if desired.
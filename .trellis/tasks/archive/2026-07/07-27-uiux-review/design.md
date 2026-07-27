# Design — UI/UX polish: ease-of-use, mobile, page-load

Companion to `prd.md`. Covers the technical approach, contracts, tradeoffs,
and rollout/rollback for the requirements in the PRD. Lower-severity findings
deferred to `implement.md` tiers are not designed here in depth.

## Scope boundaries

- **In**: `frontend/src/**`, `frontend/index.html` (dev shell), the production
  shell `go/web/views/react-shell.tmpl`, and `vite.config.ts` (bundle output).
- **Out**: Go API behavior, LESS legacy pages, content/copy rewrites, the
  reader's image-fetch pipeline (only its chrome + viewport are in scope).

## D1. Route-level code-splitting (page-load)

### Problem
`App.tsx` statically imports every page, including the Reader
(`ReaderPage` + `ReaderControls` + `ReaderViewport` + `useReaderNavigation`
+ `useFocusTrap` + `readerMath` + `useReaderBootstrap` + `useReaderProgress`
+ `useReaderPrefs`). Vite emits one `main.js`; a user on `/login` downloads
the entire reader. No `React.lazy`/`Suspense`/dynamic `import` exists
(verified by grep — only `preloadLookahead` reader pref matches).

### Approach
Lazy-load the heavy, route-gated pages with `React.lazy` + a single
`<Suspense>` boundary. Keep the lightweight, first-paint-critical pages
(Home, Login) eager so the initial route stays fast.

```tsx
// App.tsx
import { lazy, Suspense } from 'react';
import { LoadingState } from './shell/StatePanels';
import { useI18n } from './lib/i18n';

const TitleDetailPage = lazy(() => import('./pages/TitleDetailPage'));
const TagDetailPage   = lazy(() => import('./pages/TagDetailPage'));
const TagsIndexPage   = lazy(() => import('./pages/TagsIndexPage'));
const AdminPage       = lazy(() => import('./pages/AdminPage'));
const UserListPage    = lazy(() => import('./pages/UserListPage'));
const UserEditPage    = lazy(() => import('./pages/UserEditPage'));
const MissingItemsPage= lazy(() => import('./pages/MissingItemsPage'));
const ReaderPage      = lazy(() => import('./pages/reader/ReaderPage'));

function RouteFallback() {
  const { t } = useI18n();
  return <LoadingState message={t('loading')} />;
}
```

Wrap `<Routes>` (or each lazy `<Route element=...>`) in `<Suspense fallback={<RouteFallback />}>`.

**Eager (kept static):** `HomePage`, `LoginPage`, `LibraryPage`, `AppShell`,
`ErrorState`, `LoadingState`, `UnknownPage` (small, error-route). These keep
the login + home + library first paint off the critical path.

**Chunking expectation:** Vite auto-splits each `lazy()` import into its own
chunk under `assets/` (e.g. `TitleDetailPage-<hash>.js`, `ReaderPage-<hash>.js`).
With the cache-busting change in D3, these get content hashes automatically.

### Contracts / gotchas
- `LoadingState` must remain eagerly imported (it's the Suspense fallback and
  the page-level loading state) — do **not** lazy-import it.
- Each lazy page must keep its default export (already true: `export function
  HomePage` etc. are named; `React.lazy` needs a default export, so wrap:
  `lazy(() => import('./pages/X').then(m => ({ default: m.X })))`). **Verify
  each page's export style before finalizing** — `HomePage`/`LoginPage`/
  `LibraryPage` use `export function`, so the `.then(m => ({ default: m.X }))`
  adapter is required.
- `UnknownPage` uses `AppShell` + `ErrorState` (both eager) — keep it eager so
  a 404 never waits on a chunk.
- The reader's `document.title` effect lives in `ReaderPage` — fine, it runs
  after the chunk loads.

### Tradeoffs
- One extra round-trip on first navigation to a heavy route (mitigated by
  prefetch, see below). Worth it: login/home/library shed the reader payload.
- **Optional prefetch**: add `<link rel="modulepreload">` for the most-likely
  next chunk (e.g. Home → prefetch `TitleDetailPage` chunk; TitleDetail →
  prefetch `ReaderPage` chunk). This requires the build to emit a manifest
  of chunk URLs. Vite can do this via `build.manifest` or by reading the
  emitted `<link>` tags. **Defer prefetch to a follow-up** unless trivial;
  the split alone is the main win. Record in `implement.md` as Tier 2.

## D2. Mobile: reader top bar + app topbar + touch targets + blur

### D2.1 Reader top bar overflow (`ReaderTopBar.tsx`, `shell.css:1404-1474`)
Three groups in a `space-between` flex with `gap:1rem`, no `flex-wrap`, no
mobile breakpoint. On ~360px the center title (`max-width: min(48vw,28rem)`,
`white-space: nowrap`, `text-overflow: ellipsis`) is crushed between the
left group (brand + "Reading settings" button with label) and right group
(LanguageSelect + "Exit reader" button with label).

**Approach:**
1. Add `flex-wrap: wrap` to `.mango-reader-topbar` and a mobile breakpoint.
2. On narrow widths, **drop the visible labels** on the two labeled controls
   (reader-controls button, exit button) and keep them icon-only with
   `aria-label` (already present on exit; add to controls button). The center
   title collapses to a single line; if still too tight, hide the progress
   sub-line on the smallest widths and surface it in the controls modal
   (already shown there: `ReaderControls.tsx:99-101`).
3. Reduce `gap` to `0.5rem` under ~480px.

```css
@media (max-width: 560px) {
  .mango-reader-topbar { gap: 0.5rem; padding: 0.4rem 0.6rem; }
  .mango-reader-topbar__center { flex: 1 1 100%; order: 3; align-items: flex-start; }
  /* hide visible labels via a utility class toggled in ReaderTopBar */
  .mango-reader-topbar__label--sm-hide { display: none; }
}
```
In `ReaderTopBar.tsx`, give the two labeled buttons a span that carries
`mango-reader-topbar__label--sm-hide` so the text drops on mobile while the
icon + aria-label remain. Keep desktop labels intact.

**Acceptance**: fits 360px without horizontal scroll; title legible or hidden
with progress available in the controls modal; all controls reachable.

### D2.2 App topbar height on mobile (`shell.css:1334-1355`)
At ≤820px the nav wraps to `order:3; width:100%; overflow-x:auto` (good), but
the 6-button tool cluster (`mango-topbar__tools`, `flex-wrap:wrap`) can wrap
to its own line → 3-row topbar.

**Approach:** At ≤820px, keep the tool cluster on one row by making it
`flex-wrap: nowrap` and allowing it to shrink (icons are fixed-size; the
cluster is ~6 × 36px ≈ 216px + gaps, fits most phones). Only if a future
control is added does it overflow; at that point a "more" menu is the answer
(out of scope now). Also reduce topbar vertical padding at ≤560px so a
2-row topbar (brand+tools row / nav row) is compact.

```css
@media (max-width: 820px) {
  .mango-topbar__tools { flex-wrap: nowrap; }
}
@media (max-width: 560px) {
  .mango-topbar { padding: 0.45rem 0.75rem; gap: 0.4rem 0.75rem; }
}
```

### D2.3 Touch targets ≥44px (`shell.css:287-292`, `1248-1256`)
`.mango-btn--icon` is `min-width/height: 2.25rem` = 36px; the tag-remove
button overrides to `1.5rem` = 24px. Below the 44px (2.75rem) Apple/Google
guidance.

**Approach:** Raise the **mobile** hit area without changing desktop density,
using a touch-pointer media query so mouse users keep the compact look:

```css
@media (pointer: coarse) {
  .mango-btn--icon { min-width: 2.75rem; min-height: 2.75rem; }
  .mango-tag-pill .mango-btn--icon { min-width: 2.5rem; min-height: 2.5rem; }
}
```
`pointer: coarse` targets touch devices broadly (phones + tablets) without
a hard px breakpoint, and doesn't enlarge controls for trackpad/mouse users.
The tag-remove stays a bit smaller (2.5rem) because it sits inside a pill
with tight padding, but is now comfortably tappable.

**Gotcha:** `pointer: coarse` also matches tablets in pointer mode; that's
fine — larger targets there are acceptable. Do **not** use `min-width` media
queries for this, or desktops in a narrow window get oversized buttons.

### D2.4 Sticky-topbar blur scroll-jank (`shell.css:36`)
`backdrop-filter: blur(10px)` on `.mango-topbar` (sticky). Expensive on
low-end mobile, can cause scroll jank.

**Approach:** Keep the blur on desktop (it looks good and is cheap there),
drop it on coarse-pointer / small screens:

```css
@media (pointer: coarse), (max-width: 560px) {
  .mango-topbar {
    backdrop-filter: none;
    background: var(--mango-bg-surface); /* opaque fallback */
  }
}
```
The comic topbar already uses an opaque `--mango-bg-surface`, so only flat
needs the fallback (already covered by the rule above).

## D3. `<html lang>` before first paint (ease-of-use / a11y)

### Problem
`react-shell.tmpl:4` and `index.html:3` hardcode `lang="zh-CN"`. `i18n.tsx:280-282`
sets it on module load and `setLanguage` updates it, but the pre-JS paint
(and no-JS) is wrong for `en`/`zh-tw` users — screen readers get the wrong
language on first paint.

### Approach
Mirror the existing FOUC theme script: read `localStorage['mango-language']`
in an inline `<script>` in the shell head and set `documentElement.lang`
before paint. `i18n.tsx` already maps `en→en`, `zh-tw→zh-Hant`, `zh-cn→zh-Hans`
(`applyDocumentLanguage`, `i18n.tsx:252-255`); replicate that mapping in the
inline script so the two never disagree.

```html
<!-- react-shell.tmpl, in <head> after the theme FOUC script -->
<script>
  (function () {
    var lang = localStorage.getItem('mango-language');
    var map = { 'en': 'en', 'zh-tw': 'zh-Hant' };
    document.documentElement.lang = map[lang] || 'zh-Hans';
  })();
</script>
```
Apply the same inline script to `frontend/index.html` (dev shell) so dev
matches production. The module-load `applyDocumentLanguage` call
(`i18n.tsx:280-282`) stays as the source of truth after hydration; the inline
script only fixes the pre-hydration paint. They use the same mapping so no
flicker.

**Gotcha:** Keep the inline script tiny and dependency-free (no helpers) — it
runs before any module. The mapping must stay in sync with
`applyDocumentLanguage`; add a comment in both pointing at the other.

## D4. Cache-busting for `main.js` / `main.css` (page-load correctness)

### Problem
`vite.config.ts:31-39` forces fixed names: `entryFileNames: 'assets/main.js'`,
CSS forced to `assets/main.css`. `react-shell.tmpl:11,14` hardcodes those
exact URLs. After a Mango upgrade the URLs are identical, so browser/CDN
caches can serve **stale** `main.js`/`main.css` — a real correctness risk for
a self-hosted app users upgrade in place.

### Approach — content-hash + manifest resolution
Two viable shapes; **Option A** is recommended, **Option B** is the fallback.

**Option A — content-hashed names + Go-resolved manifest:**
1. Vite: switch to hashed output:
   ```ts
   entryFileNames: 'assets/[name]-[hash].js',
   chunkFileNames: 'assets/[name]-[hash].js',
   assetFileNames: 'assets/[name]-[hash][extname]',
   ```
   (drop the CSS-forcing branch; CSS becomes `main-[hash].css`).
2. Emit a manifest Vite can produce: `build.manifest = true` writes
   `manifest.json` mapping logical entry → hashed filename. Or read the
   emitted filenames from the build output.
3. Go side: read the manifest at startup (or embed it) and inject the hashed
   URLs into `react-shell.tmpl` via template fields (e.g.
   `{{.MainJS}}`, `{{.MainCSS}}`) instead of hardcoded `assets/main.js`.

**Option B — keep fixed names, force revalidation via headers:**
If the Go embed/serve path can't easily resolve a manifest, set
`Cache-Control: no-cache` (or `max-age=0, must-revalidate`) + an `ETag` on
`/react/assets/main.js` and `main.css` so the browser always revalidates.
Simpler, but every navigation revalidates the bundle (a conditional GET, not
a full re-download if unchanged). Acceptable, less optimal than A.

### Tradeoffs
- **A**: best cache behavior (immutable hashed assets, long `max-age`), but
  requires a Go-side change to consume the manifest — touches `go/web` serve
  code, slightly beyond "frontend only." This is the right long-term answer.
- **B**: frontend-adjacent only (Go cache headers), no manifest plumbing, but
  loses the "immutable asset" win and adds a revalidation hop per load.

### Decision
**Option A — content-hashed names + Go-resolved manifest** (confirmed with
Howard at the review gate, 2026-07-27). Rationale: best cache behavior
(immutable hashed assets, long `max-age`) and it is the right long-term
answer; Howard accepted the larger blast radius (Go serve + template +
`check-react-outputs.mjs` + vite config). Option B (revalidation headers)
is dropped. Implementation must update `check-react-outputs.mjs` to match
the new hashed paths (see `implement.md` pre-flight + T1.6).

## D5. Fredoka font preload (page-load, Tier 2)

`fonts.css` `@font-face` has `font-display: swap` (no FOIT — good), but the
woff2 is only discovered after CSS parses. A `<link rel="preload" as="font"
crossorigin>` in the shell starts the fetch earlier for comic-theme users.

```html
<link rel="preload" href="{{.BaseURL}}react/assets/fredoka-*.woff2"
      as="font" type="font/woff2" crossorigin>
```
**Complication:** the woff2 filename is hashed under D3 Option A, or fixed
under B. Under B (fixed names), Vite currently emits fonts as
`assets/[name][extname]` → `Fredoka-Regular.woff2`. Preload both Regular +
Bold. **Defer to Tier 2** — it's a nice-to-have; `font-display: swap` already
avoids the blocking failure mode. Only worth it if comic is the default
(which it is — the FOUC script defaults to `comic`).

## Rollout / rollback

- All changes are CSS + a few TSX edits + shell template + vite config. No
  data migrations, no API changes.
- **Rollback**: revert the commit(s). No state to clean up. The lazy-import
  adapter and CSS media queries are additive; reverting restores eager
  loading and the old mobile layout.
- **Validation gates** (see `implement.md`): `npm run typecheck`,
  `npm run build` (which runs `check-react-outputs.mjs`), manual smoke at
  360/414/768px across four theme combos + three languages.
- `check-react-outputs.mjs` likely asserts the fixed `assets/main.js` /
  `assets/main.css` paths exist — **this script must be updated** if we move
  to hashed names (Option A). Under Option B it's unchanged. Check the
  script before changing vite output names.

## Out of scope (deferred to implement.md tiers or future tasks)
- Title-detail page density reduction (4 action buttons + admin checkbox per
  entry card) — Tier 2/3.
- Missing-items table → mobile cards — Tier 3 (acceptable horizontal scroll).
- Skip-to-content link + route-change focus management — Tier 2 (a11y).
- Quick-jump select opening page 1 ignoring saved progress — Tier 3 (minor).
- Route-bundle prefetch (`modulepreload`) — Tier 2.
# React reader migration

## Goal

Migrate Mango's comic/manga reader from the legacy Alpine/jQuery/UIkit page to
the existing React + Vite shell, closing the authenticated browse → read loop
while preserving direct page URLs, reading modes, progress, and BaseURL behavior.

## Confirmed Facts

- Routes are `/reader/{title}/{entry}` (redirects to page 1) and
  `/reader/{title}/{entry}/{page}` (renders `reader.tmpl` today).
- Existing React shell already serves login, home, library, title, tags, admin
  users, and missing-items through route-specific `pageId` values; no `reader`
  pageId is registered yet.
- Existing reader APIs already cover the core data plane:
  - `GET /api/dimensions/{tid}/{eid}` → `{ success, dimensions: [{width,height}] }`
  - `GET /api/page/{tid}/{eid}/{page}` → raw page image bytes
  - `PUT /api/progress/{tid}/{page}?eid=` → `{ success: true }`
- Legacy template injects `base_url`, `page`, `tid`, `eid`, plus adjacent-entry
  and exit URLs; page images are never inlined into HTML.
- Legacy reader features (`go/web/public/js/reader.js` + `reader.tmpl`):
  - continuous (long-strip) and paged modes
  - paged fit: height / width / original
  - continuous page margin
  - RTL reading direction
  - flip animation toggle
  - preload lookahead (0–5)
  - keyboard (←/→, j/k) and click-zone navigation in paged mode
  - click image → control modal (jump page, mode, prefs, entry switch, prev/next/exit)
  - progress save with throttle (≥5 pages away, long-page titles, first/last)
  - `history.replaceState` keeps the URL page in sync
  - next-entry / exit buttons mark progress complete then navigate
  - preferences persisted in `localStorage`
- Parent task `07-17-frontend-react-vite` recommends reader as the next child
  after browse, with a bootstrap contract for entry identity, page count,
  dimensions, progress, adjacent entries, exit URL, and reading prefs; image
  bytes stay on the existing page endpoint.
- Browse React pages already deep-link into `/reader/{title}/{entry}` and
  `/reader/{title}/{entry}/1`.
- Unmigrated admin/settings/subscriptions/plugin-download/OPDS stay out of this
  child.

## Requirements

- Render both reader routes through the existing React shell without changing
  public URLs or the authentication boundary.
- Provide a stable JSON bootstrap (or equivalent) contract for reader metadata
  needed on first paint; keep page image bytes on `GET /api/page/...`.
- Preserve continuous and paged reading modes, RTL, fit options, margin,
  preload, flip animation preference, keyboard/touch/click navigation, progress
  saving, adjacent-entry navigation, exit-to-title, and direct page deep links.
- Preserve loading, empty/corrupt entry, error, and successful read states.
- Keep all routes and API calls BaseURL-aware; final Go binary must not require
  Node at runtime.
- Reuse shared React i18n/theme providers and alert/error UI; translate the
  reader chrome in this task (zh-CN / zh-TW / en).
- Use a dedicated immersive reader chrome rather than the shared `AppShell`
  topbar/page-header layout. Provide an auto-hiding reader bar (brand/exit,
  progress, language) that reveals on pointer near the top edge (and on
  intentional open from the control panel / keyboard), with a short show/hide
  animation, then auto-hides again after idle so pages stay full-viewport.
- Isolate reader navigation/progress state so URL page, displayed page, and
  persisted progress cannot drift independently.
- Keep `reader.tmpl`, `reader-error.tmpl`, and `reader.js` as route-local
  rollback assets until the child passes build and browser smoke.
- Add focused Go contract/route tests and pure frontend tests for navigation
  math / progress rules in proportion to the migrated interactions.

## Acceptance Criteria

- [ ] `/reader/{title}/{entry}` redirects (or resolves) to a React reader at a
      concrete page URL without breaking existing deep links.
- [ ] `/reader/{title}/{entry}/{page}` renders React and opens on the requested
      one-based page when valid.
- [ ] Continuous and paged modes work; mode and related prefs persist across
      reloads under namespaced `mango.reader.*` localStorage keys (no requirement
      to migrate legacy unprefixed keys).
- [ ] Keyboard, click-zone, and control-panel navigation update the visible page
      and the browser URL without full reload.
- [ ] Progress is saved through the existing progress API with legacy-equivalent
      throttling rules (including first/last/long-page cases).
- [ ] Next-entry, previous-entry, entry jump, and exit-to-title flows work under
      root and non-root BaseURL.
- [ ] Missing/corrupt entries surface a clear error state instead of a blank
      reader.
- [ ] Reader chrome is available in Simplified Chinese, Traditional Chinese, and
      English through the shared language selector.
- [ ] Reader uses immersive full-viewport chrome (not shared AppShell). An
      auto-hide bar appears on top-edge hover / intentional open with animation
      and hides again after idle without blocking page navigation.
- [ ] Unmigrated routes remain on Go templates; OPDS and download links are
      unchanged.
- [ ] Focused Go and frontend tests cover bootstrap/error contracts and
      navigation/progress math; production build still embeds static assets only.

## Out of Scope

- Admin settings, subscriptions, plugin download, OPDS, or further page
  migrations.
- Rewriting image serving, archive decoding, or thumbnail generation.
- Pixel-perfect UIkit reproduction; feature-equivalent React layout is enough.
- Product removal of legacy reader assets before smoke/validation passes.
- Changing public reader URL shapes.

## Decisions

- Reader chrome: immersive full-viewport reader (not shared `AppShell`).
  Auto-hide top bar on edge hover / intentional open, with short animation;
  default state is hidden so reading stays immersive. Control panel for mode,
  prefs, page/entry jump remains available as in legacy (click image / hotkey).
- Reader metadata bootstrap: add a dedicated lightweight endpoint
  `GET /api/reader/{tid}/{eid}` that returns entry identity/name, page count,
  dimensions (or a clear pointer), current progress, sibling entries, adjacent
  entry URLs, and exit URL in one response. Page image bytes remain on
  `GET /api/page/{tid}/{eid}/{page}`. Progress writes stay on existing
  `PUT /api/progress/{tid}/{page}?eid=`.
- Reader preferences: use namespaced `mango.reader.*` localStorage keys
  (`mode`, `margin`, `fitType`, `preloadLookahead`, `enableFlipAnimation`,
  `enableRightToLeft`). Do not read or migrate legacy unprefixed keys; users
  may need to re-set prefs once after the cutover.

## Open Questions

- (none blocking)

# Fix remaining theme consistency findings

## Goal

Close remaining **P1** (and safe **P2**) consistency gaps from `07-11-ui-theme-style-review/findings.md` after P0 F01/F02 (`07-11-fix-p0-flat-dark`).

## User value

comic/flat × light/dark should feel intentional on login, forms, cards, tags, mobile chrome, and first paint — not half-themed.

## Source findings (in scope)

| id | priority | summary |
|----|----------|---------|
| F03 | P1 | Login: wire comic card classes so Kirby styles apply |
| F04 | P1 | Login: comic+dark bg aligns with dark paper (not loud light Kirby only) |
| F06 | P1 | comic+dark form labels readable |
| F08 | P1 | Title select2 has comic-scoped look |
| F09 | P1 | Mobile navbar comic |
| F10 | P1 | Mobile offcanvas comic |
| F11 | P1 | Anti-FOUC: early theme/style classes on html/body |
| F12 | P1 | Admin uiStyleChanged also ensures theme classes consistent (`setTheme` after style) |
| F14 | P1 | Card progress uses comic progress language under comic |
| F15 | P1 | Title `#select-bar` comic-scoped |
| F18 | P2 | Wire dead selectors we fix via markup (login/progress) rather than leave dead |

## Out of scope

| id | reason |
|----|--------|
| F01/F02 | Done in `07-11-fix-p0-flat-dark` |
| F05 | Product: no new login FAB/style UI this task; storage+ready still apply style |
| F07 | Flat tag pills intentional fallback; only F08 needs comic select2 |
| F13 | Reader dark shell intentional; no reader FAB |
| F16 | Full `mango-app-shell` purge too large; leave dead block |
| F17 | Full token unification across mango/tags/hardcodes — separate |
| F19 | FAB already OK |
| F20 | Spec write via finish / update-spec, not code |

## Decisions

| Decision | Choice |
|----------|--------|
| Login comic (F03) | Add `comic-login-card`, `comic-login-header`, `comic-login-title`, `comic-login-form` classes alongside existing `login-*` (no markup redesign) |
| Login comic dark (F04) | Under `body.comic-theme-dark .comic-login-page` use dark paper bg; keep Kirby gradient for light comic only |
| Progress (F14) | Prefer CSS: style `.comic-card .card-progress-bar/fill` under comic; avoid dual class churn if possible. Optional dual classes if CSS alone insufficient |
| FOUC (F11) | Early inline in `head` after common.js load is too late for body; apply classes to `document.documentElement` immediately and body when present; layout/login call `setUIStyle(); setTheme();` |
| Admin (F12) | `uiStyleChanged` → `setUIStyle` then `setTheme()` |
| Mobile chrome | comic overrides for `.mango-navbar` and `#mobile-nav .uk-offcanvas-bar` under `body.comic-theme*` (+ dark variants) |
| select2 / select-bar | comic-theme.less scopes under `body.comic-theme*` |

## Requirements

- R1–R10: Each in-scope finding fixed with intentional look in relevant states.
- R11: No third theme axis; reuse existing classes/tokens.
- R12: Recompile assets (`npm run uglify` / gulp).
- R13: Do not break flat light/dark or P0 login flat-dark fix.

## Acceptance criteria

1. Login comic light: Kirby card (border/shadow/header) visible.
2. Login comic dark: dark paper page bg; card still readable comic language.
3. Login flat light/dark: unchanged relative to P0 (dark still `#121212` page).
4. comic+dark form labels readable (not near-black on dark paper).
5. Title page select2 under comic uses Kirby/comic palette (not only coral).
6. Mobile navbar + offcanvas under comic match comic chrome language.
7. First paint: html/body get style/theme classes without waiting only on jQuery ready where possible; layout/login call setUIStyle+setTheme.
8. Admin style change re-applies theme classes correctly in dark.
9. Card progress under comic looks comic (stripes/border language), flat unchanged coral.
10. Title select-bar under comic looks comic; flat unchanged.
11. gulp build succeeds.

## Complexity

Complex multi-surface CSS/JS/markup → needs `design.md` + `implement.md` before start… **User already approved “continue the rest” as implementation consent for remaining findings.** Treat as execution of audit backlog; PRD + implement checklist required.

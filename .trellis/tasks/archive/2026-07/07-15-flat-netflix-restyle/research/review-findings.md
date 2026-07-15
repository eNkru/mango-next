# Self-review findings (2026-07-15)

## Fixed in this pass

1. **FOUC / layout flash** — Critical layout rules only used `body.flat-theme`, but head FOUC only sets classes on `html` until body mounts → sidebar + left margin could flash. Now `html.flat-theme` also hides sidebar / zeroes content margin / shows top bar.
2. **Style toggle without re-theming** — `toggleUIStyle` only called `setUIStyle`, not `setTheme`, so `uk-light` / dark markers could desync. Now calls `setTheme(loadTheme())` and loads comic fonts when switching to comic.
3. **Admin style change** — same setTheme + comic font path.
4. **Dead selector** — `html.flat-theme.uk-light` never matched (`uk-light` is body-only); cleaned up.
5. **Home horizontal overflow** — `overflow-x: clip` on `.home-page` for negative margin full-bleed.

## Remaining risks (not all fixed)

| Issue | Severity | Notes |
|-------|----------|--------|
| `_variables.less` Flat coral→red but **mango.css not recompiled** | Medium | Shared mango rules still use old compiled coral until full less rebuild; flat-theme.css overrides many surfaces with !important red |
| Dual sources: flat-theme.less vs flat-theme.css hand-synced | Medium | Drift risk; css is runtime truth |
| Many components still rely on comic class names + flat overrides | Low | Works but brittle |
| Select2 base colors in tags.css still coral for unscoped rules | Low | Flat dark path overridden; light flat may still see coral until select2 block wins |
| Reader has no top bar (standalone page) | By design | Shell polish only |
| Active nav highlight on flat-topbar not wired | Low | Links work; no current-page class |
| Login forces dark-first then light override | Low | OK if cascade order preserved |
| Manual smoke incomplete | — | User should verify comic/flat/dark/light matrix |

## Suggested user smoke matrix

1. Comic + Light / Dark — sidebar present, no top bar, comic fonts
2. Flat + Dark — top bar, no sidebar, red accent, home continue single row, tall side cards
3. Flat + Light — light chrome, still red CTAs
4. Toggle style mid-session (FAB + Admin select)
5. Title tags Select2
6. Reader open + modal

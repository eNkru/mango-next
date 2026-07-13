# Align GitHub Float Button with Selected Theme

## Goal
Make the floating GitHub shortcut (`.github-float`) visually consistent with the user's active theme and UI style, instead of looking like a generic white pill that only partially reacts to dark mode.

## User value
The GitHub button should feel like part of the chrome (sidebar / comic UI), not a bolted-on default control that clashes with light/dark or comic/flat.

## Confirmed facts (from codebase)

1. **Markup** — `src/views/layout.html.ecr`: fixed top-right link to `https://github.com/eNkru/mango-next`, class `github-float`, Font Awesome brands icon `fab fa-github`. Brands webfonts + brands.less import already wired in `gulpfile.js` / `mango.less`.

2. **Theme axes (runtime)**  
   - **Theme** (`localStorage.theme`): `light` | `dark` | `system` → body `uk-light` when dark (`public/js/common.js` `setTheme`).  
   - **UI style** (`localStorage.ui-style`): `comic` (default) | `flat` → `comic-theme` / `comic-theme-dark` on html/body (`setUIStyle`).

3. **Current CSS** (`public/css/mango.less` ~4240–4328)  
   - Base: white circle, `#333` icon, soft shadow (flat light baseline).  
   - `.uk-light .github-float`: translucent dark treatment (flat dark baseline).  
   - `body.mango-app-shell` (+ light variant): dead rules for this control (class never applied at runtime).  
   - **Gap**: no rules under `body.comic-theme` / `body.comic-theme-dark`.

4. **Reader page**: `reader.html.ecr` uses its own document (no layout float) — no hide/restyle work.

5. **i18n**: `data-i18n-title="view_on_github"` present; key missing in `public/js/i18n.js` (`en` / `zh-cn` / `zh-tw`). Layout also has hardcoded English `title` / `aria-label` as fallbacks.

6. **Related separate task**: `07-09-fix-reader-blank` — out of scope.

## Decisions

| Decision | Choice | Notes |
|----------|--------|--------|
| Comic visual language | **A — full Kirby pop** | Thick black outline, hard offset shadow, comic palette. |
| Dead shell GitHub rules | **A — delete** | Only `.github-float` under `mango-app-shell` (+ light variant). Leave rest of shell theme block. |
| i18n `view_on_github` | **A — include** | Add keys for `en`, `zh-cn`, `zh-tw` near other layout chrome keys (`theme_toggle` / `ui_style_toggle`). |
| Style file split | **A — comic in `comic-theme.less`** | Flat base + flat dark in `mango.less`; comic light/dark in `comic-theme.less`. |
| Comic hover fill | **A — `@comic-red`** | Default rest state: paper/white + black border + hard shadow; hover: red fill, white icon (primary CTA language). |

## Requirements

- R1: Intentional look in comic+light, comic+dark, flat+light, flat+dark.
- R2: Class-driven only via existing body/html classes (no new JS).
- R3: Preserve URL, `target="_blank"`, `rel`, focus-visible, mobile offset under top bar.
- R4: Reuse existing design tokens (`@accent` / dark flat tokens; `@comic-*` in comic file).
- R5: Comic = full Kirby pop; hover uses `@comic-red`.
- R6: Delete dead shell-scoped `.github-float` rules.
- R7: Add `view_on_github` for en / zh-cn / zh-tw; keep sensible HTML fallbacks for default lang / pre-i18n.
- R8: Comic styles in `public/css/comic-theme.less`; flat base/dark refinements in `public/css/mango.less`.

## Acceptance criteria

1. **Comic + light**: Kirby chrome — thick black border, hard offset shadow, comic paper/white rest fill; hover red fill + white icon.
2. **Comic + dark**: Same structure with dark comic fills/borders; remains readable on dark paper.
3. **Flat + light / flat + dark**: Soft circle using flat surface/accent language; dark uses `.uk-light` (or equivalent) refinements already in mango.less, tuned if needed for token consistency.
4. No layout regression on existing desktop (`top: 16px; right: 16px`) / mobile (`top: 64px` under navbar) positions and sizes.
5. Toggling theme or UI style updates the button immediately without reload.
6. No remaining `.github-float` selectors under `body.mango-app-shell`.
7. When language is `en` or `zh-tw`, title attribute resolves via i18n (`view_on_github`); `zh-cn` fallback text in markup remains correct Chinese (or key-backed Chinese).
8. Comic rules scoped under `body.comic-theme` / `body.comic-theme-dark` in `comic-theme.less`.

## Out of scope

- Changing the GitHub URL or removing the float.
- Full layout redesign.
- Full `mango-app-shell` theme cleanup (beyond GitHub float rules).
- Reader blank fix (`07-09-fix-reader-blank`) or reader float visibility.
- New theme systems beyond comic/flat × light/dark.
- Markup structure redesign (keep single `.github-float` anchor unless a minimal a11y fix is required for aria-label i18n).

## Suggested copy (implementation default)

| Locale | `view_on_github` |
|--------|------------------|
| en | View on GitHub |
| zh-cn | 在 GitHub 上查看 |
| zh-tw | 在 GitHub 上檢視 |

## Complexity

Lightweight: **PRD-only** is sufficient. Pure CSS + small i18n dict update; no backend, no new components, no data flow changes.

## Open questions

None — planning complete pending user review / approval to start implementation.

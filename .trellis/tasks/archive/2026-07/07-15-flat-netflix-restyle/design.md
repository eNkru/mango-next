# Design: Flat UI Netflix-inspired restyle

## Intent

Re-skin and re-chrome **Flat** only so it feels like a Netflix-style content browser. **Comic** layout and styles remain the existing path.

## Dual layout strategy

| UI style | Chrome | CSS body/html markers |
|----------|--------|------------------------|
| `comic` | Current left sidebar (`top.tmpl` comic path) | `comic-theme` / `comic-theme-dark` |
| `flat` | **Top nav** + full-width main | e.g. `flat-theme` + optional `flat-theme-dark` / light via existing `uk-light` or dedicated classes |

Implementation approach:

1. **`top.tmpl` / layout shell**: structure both navs in DOM or branch with classes; show sidebar only under comic; show top bar only under flat (CSS `display` + `setUIStyle` class toggles). Prefer CSS visibility over server-side branch so toggle stays client-side.
2. **`setUIStyle` / `setTheme`**: add/remove `flat-theme` (name TBD); flat dark uses CSS variables, not only inline `#121212`.
3. **Utility FAB**: restyle for flat; keep language/theme/ui-style actions.

## Design tokens (Flat Netflix)

Extend `_variables.less` (flat section only):

### Dark

| Token | Direction |
|-------|-----------|
| bg base | ~`#141414` / `#000` family |
| bg elevated | ~`#181818`–`#1f1f1f` |
| text primary | `#fff` / `#e5e5e5` |
| text secondary | `#b3b3b3` |
| accent | **`#E50914`** (+ hover darker) |
| border | low-contrast `#2a2a2a` |
| radius | small–medium (less “bubble” than comic) |
| shadow | soft black glow on hover cards |

### Light (also Netflix-structured)

| Token | Direction |
|-------|-----------|
| bg base | off-white / light gray canvas |
| surfaces | white cards, subtle border |
| text | near-black / gray hierarchy |
| accent | same red family |
| chrome | light top bar, dark logo text or red accents |

Do **not** change comic variables.

## Component language (Flat)

1. **Top bar**: logo left; Home / Library / Tags / Admin links; right cluster (or FAB) for tools.
2. **Poster card**: cover-forward; title/meta on hover or under; slight scale + shadow on hover.
3. **Rail / section**: section title + horizontal row (home) or dense grid (library) with consistent gutters.
4. **Buttons**: red primary, ghost/secondary for secondary actions; less thick comic borders.
5. **Forms** (login, admin, select2): flat-scoped dark/light fields matching canvas.
6. **Reader shell**: dark chrome, red accents on controls; page image area unchanged functionally.

## Page mapping

| Page | Flat treatment |
|------|----------------|
| Global shell | Top bar, main padding under bar, no sidebar |
| Home | Featured + rails (continue / start / recent) |
| Library / Tag | Poster grid, filters as slim bar |
| Title | Large cover + episode/chapter grid rails |
| Tags | Chip/list consistent with tokens |
| Login | Dark/light stream-style panel (no Netflix branding) |
| Admin + subpages | Same top bar + denser cards/tables with tokens |
| Reader | Shell/modal/buttons; keep reading modes |
| Plugin/download/subscription | Same chrome + cards |

## CSS architecture

Recommended layering:

```
_variables.less     # flat Netflix tokens + keep comic tokens
flat-theme.less     # NEW: body.flat-theme … all flat overrides
mango.less          # shared structure; reduce flat-specific hacks over time
comic-theme.less    # unchanged responsibility
tags.less           # flat select2 under flat-theme, keep non-comic guards
```

Compile pipeline: match existing less → css workflow used for comic-theme/mango.

Scope rule: **all Netflix rules under `html.flat-theme` / `body.flat-theme`** (and dark variant) so comic never inherits.

## JS touch points

- `common.js`: `setUIStyle`, `setTheme`, class application, remove flat dark-only inline bg if CSS owns it
- `admin.js`: ui style labels unchanged (Comic / Flat)
- i18n: only if new strings; prefer reuse

## Delivery batches (3 PRs)

| Batch | Deliverable | Exit criteria |
|-------|-------------|---------------|
| **1 Shell + tokens** | flat-theme tokens, top bar layout toggle, global bg/type/buttons/FAB, dual theme base | Toggle flat/comic; dark/light canvas correct; comic sidebar OK |
| **2 Browse** | Home, Library, Title, Tags poster/rails | Main browse feels Netflix; data/actions work |
| **3 Rest of app** | Login, Admin*, Reader shell, plugin/download/subscription | Full-site flat consistency |

## Risks

| Risk | Mitigation |
|------|------------|
| Templates share comic class names | Flat styles target structure + dual classes (`comic-card` still styled under flat-theme) |
| Toggle FOUC | Early class on html in head script (like font load) |
| Select2 / UIkit leaks | Scope under flat-theme; extend tags.css pattern |
| Huge CSS | flat-theme.less modular sections; batch PRs |
| Light Netflix awkward | Dedicated light tokens; don’t invert dark poorly |

## Non-goals

- Netflix assets/branding
- Recommendation engine
- Changing default ui-style away from comic unless product asks later

## Rollback

- Revert flat-theme + shell branches; comic path independent
- Feature is CSS/layout class based; no DB migration

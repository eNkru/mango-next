# Design: remaining theme consistency

## Boundaries

- Touch: `login.html.ecr`, `card.html.ecr` (only if CSS insufficient), `layout.html.ecr` scripts, `head`/common.js FOUC, `admin.js`, `comic-theme.less`, possibly `tags.less` or comic select2 in comic-theme.less, `mango.less` only if needed for shared hooks.
- Do not rewrite theme system; class axes stay comic/flat × light/dark.

## Approaches by finding

### F03 Login card
Markup: add comic structural classes next to existing ones on login form shell so existing `.comic-login-page .comic-login-card` rules match.

### F04 Login comic dark
CSS only: `body.comic-theme-dark.comic-login-page` or `body.comic-theme-dark .comic-login-page` → `@comic-dark-gray` (or paper-dark) background; light path keeps Kirby gradient on `.comic-login-page` when not dark body.

Note: body has both `comic-login-page` and theme classes.

### F06 Form labels
In `body.comic-theme-dark`: `.comic-form-label { color: @comic-white or light gray }`; inputs text/bg if needed.

### F08 select2
Under `body.comic-theme, body.comic-theme-dark` style `.select2-container--default` borders/choices with comic black/yellow/red; dark variant under comic-theme-dark. tags.less stays flat baseline.

### F09/F10 Mobile chrome
`body.comic-theme .mango-navbar` thick border / paper / yellow accents; dark comic dark surfaces. Same for `#mobile-nav .uk-offcanvas-bar`.

### F11 FOUC
1. `setUIStyle`/`setTheme` already use jQuery `$('html')`/`$('body')` — when body missing, body classes no-op.
2. Harden: apply to `document.documentElement` via classList; for body use `document.body` if available.
3. layout + login inline: `setUIStyle(); setTheme();` instead of only setTheme.
4. Keep ready re-apply.

### F12 Admin
`setUIStyle(newStyle); setTheme();`

### F14 Progress
`body.comic-theme .comic-card .card-progress-bar` / fill / percent → comic stripe + black border (mirror `.comic-progress-*`).

### F15 select-bar
`body.comic-theme #select-bar` Kirby border/shadow; dark comic variant.

## Compatibility

- Flat remaps and flat-dark P0 rules must remain.
- Reader unchanged except global FOUC improvements.

## Rollback

Revert LESS/JS/ECR changes; re-run gulp.

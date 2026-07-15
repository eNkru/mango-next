# Research: Netflix aesthetic vs Mango Flat

Date: 2026-07-15  
Reference: https://www.netflix.com/nz/ (marketing/browse visual language)

## Netflix visual language (observed)

| Trait | Netflix |
|-------|---------|
| Base | Near-black canvas, full-bleed dark |
| Accent | Saturated red CTAs (`#E50914` family) |
| Chrome | Minimal; content first |
| Typography | Clean geometric sans; high contrast white/gray |
| Content | Poster/card rows, large hero, hover scale + gradient overlays |
| Motion | Subtle hover lift/scale, row scroll |
| Nav | Typically top bar (browse app); marketing has logo + Sign in |

## Mango Flat today

| Trait | Flat (`ui-style=flat`) |
|-------|------------------------|
| Switch | `localStorage ui-style`; default **comic**; `setUIStyle` removes `comic-theme*` |
| CSS | `mango.css` (UIkit + app); comic rules live in `comic-theme.css` and mostly no-op when flat |
| Palette | Warm coral accent `#D96A4B`; light `#FAFAFA` / dark `#121212` (`_variables.less`) |
| Layout | Left **sidebar** + main; home has continue-reading hero + carousels already |
| Classes | Many templates use `comic-*` class names even in flat (styles only apply under comic body classes) |

Key files:

- `go/web/public/js/common.js` — comic/flat toggle, dark bg `#121212` inline for flat dark
- `go/web/public/css/_variables.less` — flat vs comic tokens
- `go/web/public/css/mango.less` → `mango.css`
- `go/web/views/home.tmpl`, `library.tmpl`, `title.tmpl`, `top.tmpl`, `login.tmpl`

## Mapping Netflix → Flat (candidates)

1. **Token shift**: flat dark → Netflix blacks + red accent; flat light optional softer gray or deprecate emphasis
2. **Cards**: larger posters, hover scale, bottom gradient title (home/library already card-based)
3. **Sections**: stronger “row” feel on home rails
4. **Chrome**: dark sidebar/top bar denser, less border noise
5. **Out of scope traps**: cloning Netflix marketing landing, legal copy, real Netflix logo/assets

## Constraints

- Comic UI style must remain intact
- Theme Dark/Light/System still exists; Netflix is naturally dark-first
- Prefer CSS/token + scoped flat rules; avoid rewriting all templates unless layout must change
- Branch intent: `ui/restyling`

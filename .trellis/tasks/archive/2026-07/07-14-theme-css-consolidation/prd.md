# Consolidate theme CSS: shared variables, dead code removal, and DRY

## Goal

Eliminate the three-way duplication of color variables, remove unused comic
decorative CSS, and consolidate the flat-fallback section so that each theme
file has a single, well-scoped responsibility.

## Background

The project ships two UI styles — **Flat** (`mango.less` + `tags.less`) and
**Comic** (`comic-theme.less`) — each supporting light and dark modes. A code
review found that the accent color `#D96A4B` is defined in three independent
locations, ~30 comic decorative classes (sound effects, speech bubbles,
halftone overlays, panel tilts, etc.) are defined but never referenced in any
template or Go handler, a ~450-line "flat fallback" block in `comic-theme.less`
duplicates the entire flat palette with hardcoded literals, and Google Fonts
for the comic style are unconditionally loaded even in flat mode.

## Requirements

- A single shared variables file must be the source of truth for all palette
  colors used across `mango.less`, `tags.less`, and `comic-theme.less`.
- No hardcoded color literals that duplicate a shared variable (e.g. `#D96A4B`,
  `#E8845F`, `#121212`) may remain in `comic-theme.less` flat-fallback or
  dark-flat-override sections.
- Unused comic decorative classes and their associated keyframe animations must
  be removed. "Unused" means the class name does not appear in any `.tmpl`
  template, any Go source file, or any other LESS file that is actually
  referenced.
- The flat-fallback and dark-flat-override sections in `comic-theme.less` must
  be removed or drastically reduced. Structural `comic-*` classes must
  degrade to UIKit defaults or sensible base styles defined in `mango.less`
  — not to a parallel flat implementation inside `comic-theme.less`.
- Google Fonts (Bangers + Fredoka One) must only load when the comic UI style
  is active.
- `tags.less` Select2 rules must be scoped to non-comic mode so they cannot leak
  into comic theme.
- The flat dark background value must be consistent between JS inline style and
  CSS body gradient (pick one value, use everywhere).
- After changes, the compiled CSS must produce visually identical results for
  both flat and comic themes in both light and dark modes.
- Existing functionality (theme switching, UI style toggling, all page
  rendering) must continue to work without regressions.

## Acceptance Criteria

- [ ] A `_variables.less` (or equivalent shared file) is imported by
      `mango.less`, `tags.less`, and `comic-theme.less`.
- [ ] `grep -rn '#D96A4B\|#E8845F\|#121212\|#1A1A1A' comic-theme.less` returns
      zero results (all replaced by shared variables).
- [ ] At least 25 unused comic decorative classes and their keyframes are
      removed from `comic-theme.less`.
- [ ] The flat-fallback section in `comic-theme.less` is removed or reduced to
      under 30 lines of truly necessary structural defaults.
- [ ] Google Fonts `<link>` in `head.tmpl` is conditionally loaded based on UI
      style.
- [ ] `tags.less` Select2 rules are scoped to `body:not(.comic-theme)` to
      prevent comic-mode leakage.
- [ ] Flat dark background is a single consistent value across JS and CSS.
- [ ] No visual regression in flat light, flat dark, comic light, or comic dark
      (verified by code inspection of compiled CSS or visual check).

## Out of Scope

- Introducing CSS Custom Properties (runtime variables) — separate future task.
- Reducing the UIKit bundle size — separate future task.
- Reducing the comic color palette count — separate aesthetic task.
- Changing any template HTML structure or Go handler logic.
- Redesigning the overall theme architecture.

## Confirmed Facts

- `@accent: #D96A4B` is independently defined in `mango.less:21`, `tags.less:4`,
  and hardcoded as `#D96A4B` ~20 times in `comic-theme.less` flat-fallback.
- `@dark-accent: #E8845F` follows the same pattern.
- 627 `!important` in `comic-theme.less`; 215 in `mango.less`.
- ~30 comic classes (sound-boom, speech-bubble, halftone-overlay,
  comic-panel-tilt, comic-border, comic-flip, comic-title-3d, etc.) are defined
  in `comic-theme.less` but never referenced in templates or Go code.
- ~450 lines of flat-fallback CSS in `comic-theme.less` (lines ~1630-2077)
  duplicate the mango.less flat palette with hardcoded literals.
- `head.tmpl` loads Bangers + Fredoka One unconditionally via Google Fonts.
- `tags.less` Select2 rules are globally scoped; `comic-theme.less` also styles
  Select2 under `body.comic-theme`, creating potential conflicts.
- JS sets `html.background = '#121212'` but `mango-app-shell` body gradient
  starts at `#101116`.
- LESS files are pre-compiled to CSS; the compiled `.css` is what's embedded.

## Notes

- This is a CSS-only refactor — no Go code, no template HTML structure changes
  (except `head.tmpl` font loading).
- The compiled `.css` files must be regenerated after LESS changes. If a LESS
  compiler is not available in the environment, the LESS changes must be
  manually reflected in the compiled CSS.
- Keep the PRD focused on requirements; add design.md if the approach needs
  further explanation.

# Design: Theme CSS consolidation

## Constraint

No LESS compiler is available in the build or runtime environment. The
Dockerfile copies pre-compiled `.css` files; `//go:embed public/*` serves
them directly. Therefore every change that has runtime impact must be
applied to BOTH the `.less` source (for future maintainability) AND the
compiled `.css` (for immediate runtime effect).

Changes that only restructure source without changing values (e.g. creating
`_variables.less` and replacing literals with variable references) produce
byte-identical compiled CSS, so only the `.less` needs editing.

## Architecture

### Shared variables file

Create `go/web/public/css/_variables.less` with all palette colors currently
duplicated across `mango.less`, `tags.less`, and `comic-theme.less`:

```less
// Flat palette
@accent: #D96A4B;
@accent-hover: #C2594A;
// ... (all flat light/dark variables)

// Comic palette
@comic-red: #E23636;
// ... (all comic variables)
```

`mango.less`, `tags.less`, and `comic-theme.less` each `@import` it and
delete their inline duplicates. Compiled CSS output is identical because
LESS resolves variables at compile time.

### Dead code removal

Remove ~30 unused comic decorative classes from `comic-theme.less` and
`comic-theme.css`. The classes are identified by grepping all `.tmpl`
templates and Go source — if a class name appears nowhere except the CSS
definition, it is dead.

Affected classes (confirmed unused):
- Sound effects: `sound-boom`, `sound-pow`, `sound-crash`, `sound-wham`,
  `sound-effect` (base)
- Explosion: `explosion-burst`, `click-explode`
- Bubbles: `speech-bubble`, `thought-bubble`
- Halftone: `halftone-overlay`, `halftone-heavy`
- Action lines: `action-lines-bg`, `radial-action-lines`
- Panel transforms: `comic-panel-tilt`, `comic-panel-skew`,
  `comic-panel-irregular`, `comic-panel-diagonal`
- Borders: `comic-border`, `comic-border-thick`, `comic-border-thin`
- Misc: `comic-flip`, `comic-title-3d`, `comic-heading-entrance`,
  `comic-alert`, `comic-breadcrumb`
- Unused button variants: `comic-btn-info`, `comic-btn-success`,
  `comic-btn-warning`
- Unused keyframes: `boom-enter`, `pow-float`, `explosion-burst`,
  `click-explosion`, `comic-flip`, `comic-heading-bam`

Also remove dark-comic overrides for these classes.

### Flat-fallback removal

Remove the "FLAT MODE COMPATIBILITY" section (~lines 1630-1850 in
`comic-theme.less`) and "DARK FLAT MODE OVERRIDES" section (~lines 1990-2077)
from both `.less` and `.css`.

Structural `comic-*` classes used in templates (e.g. `comic-btn`, `comic-card`,
`comic-hero`) already have UIKit fallbacks or `mango.less` styles. The
flat-fallback was duplicating those with hardcoded literals.

If any structural class truly needs a flat-mode base style, add it to
`mango.less` where it belongs, not to `comic-theme.less`.

### Conditional Google Fonts

Replace the unconditional `<link>` in `head.tmpl` with a JS-injected approach:

```html
<script>
  if (localStorage.getItem('ui-style') !== 'flat') {
    var l = document.createElement('link');
    l.rel = 'stylesheet';
    l.href = 'https://fonts.googleapis.com/css2?family=Bangers&family=Fredoka+One&display=swap';
    document.head.appendChild(l);
  }
</script>
```

This runs synchronously before body render, avoiding FOUC for comic mode
while eliminating the render-blocking request in flat mode.

### tags.less scoping

Wrap the `.uk-light` Select2 dark-mode rules in `tags.less`/`tags.css` with
`body:not(.comic-theme):not(.comic-theme-dark)` to prevent leakage into
comic mode.

### Dark background consistency

Change `mango-app-shell` body dark gradient base from `#101116` to `#121212`
to match the JS inline `#121212`. Update both `.less` and `.css`.

## Data Flow

```
_variables.less  ← shared by all three LESS files
    ↓
mango.less       → mango.css (flat theme + structural base styles)
tags.less        → tags.css (Select2, scoped to non-comic)
comic-theme.less → comic-theme.css (comic theme only, no flat fallback)
    ↓
head.tmpl        → conditional font loading
    ↓
//go:embed       → served CSS
```

## Trade-offs

- Editing both .less and .css manually doubles the effort, but ensures
  immediate runtime effect and source maintainability.
- Removing the flat-fallback may reveal edge cases where a `comic-*` class
  is used in flat mode without a UIKit equivalent. These are identified and
  patched in mango.less as needed.
- The JS font-loading approach adds a small inline script to head.tmpl,
  but eliminates render-blocking for the majority (flat) use case.

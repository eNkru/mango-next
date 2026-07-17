# Frontend asset dependency pipeline design

## Scope and boundaries

This delivery establishes a deterministic, repository-local asset pipeline for
Mango's existing server-rendered frontend. It does not introduce a JavaScript
bundle, change the Go template architecture, broadly upgrade runtime libraries,
or complete the planned jQuery removal.

The pipeline owns three distinct asset classes:

1. **Project source**: Mango templates, project JavaScript, LESS sources, images,
   manifest, and other hand-authored files.
2. **Managed copies**: pinned third-party JavaScript, CSS, fonts, and framework
   images copied from npm packages through an explicit allowlist.
3. **Generated outputs**: the four committed runtime stylesheets compiled from
   LESS: `mango.css`, `comic-theme.css`, `flat-theme.css`, and `tags.css`.

Generated and copied outputs remain committed because `go:embed` consumes
`go/web/public` at Go compile time. Node is a build-time prerequisite only; the
Mango binary remains self-contained and makes no runtime package-manager or CDN
requests.

## Package and command contract

Use a root `package.json` and committed `package-lock.json`. npm is the only
required package manager, and clean installs use `npm ci`.

The package scripts expose these stable contracts:

- `assets:copy`: copy an explicit map of pinned browser files from
  `node_modules` into declared `go/web/public` destinations.
- `assets:css`: compile the four LESS entrypoints with the repository-local
  compiler and one deterministic npm import root.
- `assets:licenses`: produce or copy the third-party notices required by the
  managed runtime assets.
- `assets:build`: run copy, CSS, and license generation.
- `assets:check`: build declared outputs in an isolated temporary directory and
  byte-compare them with committed outputs without dirtying the worktree.

A small checked-in Node script should implement explicit file copying and output
comparison. It must fail if an expected upstream path is missing and must never
recursively replace directories containing Mango-owned files.

## Dependency policy

The first migration pins known current versions to reproduce behavior rather
than modernize it. Known managed inputs include jQuery 3.2.1, jQuery UI 1.12.1,
Alpine 2.8.0 regular build, UIkit 3.5.9, Moment 2.24.0, Select2 4.1.0-beta.1,
Font Awesome Free 5.15.4, LESS, and locally hosted comic fonts.

Dependencies with unresolved exact provenance, such as `jquery.inview.min.js`,
must remain committed and be listed as retained manual vendor files until their
bytes and behavior can be matched to a specific upstream release. The inventory
must not claim certainty that the repository cannot prove.

Broad runtime upgrades are follow-up work. Each retained old version receives a
documented compatibility/security decision in the inventory.

## LESS source reconciliation

The current committed CSS is not reproducible from the current LESS. Before
enabling drift checks, existing runtime behavior must be reconciled back into the
LESS source of truth:

- Normalize UIkit and Font Awesome imports so all packages resolve from the root
  npm installation.
- Port CSS-only rules from `comic-theme.css`, `flat-theme.css`, and `tags.css`
  into their corresponding LESS files.
- Compile all four entrypoints with one pinned LESS version and one formatting
  contract.
- Treat regenerated CSS as the output; do not continue hand-synchronizing CSS.

Because reconciliation can change cascade order or formatting, theme smoke
checks must cover comic and flat styles in light and dark modes, card title
height, navigation chrome, library cards, and tag controls.

## Removing dotdotdot without jQuery

`dotdotdot.js` 4.0.11 is removed rather than copied into the new pipeline. Its
CC-BY-NC-4.0 license is unnecessarily restrictive, and the supported browsers
already implement the CSS mechanism used by Mango's current styles.

Card-title truncation uses the existing `display: -webkit-box`,
`-webkit-line-clamp`, `-webkit-box-orient`, and `overflow: hidden` rules. Relevant
templates add native `title` attributes containing the complete title. The
`dots.tmpl` script partial, `dots.js`, and `dotdotdot.js` are removed, and pages
stop including that partial.

`jquery.inview.min.js` is not deleted in this delivery because
`reader.js` still uses its `inview` event in continuous reading mode. The future
jQuery-removal design should replace that behavior with `IntersectionObserver`,
then remove the plugin. No new code in this task may depend on jQuery.

## Local fonts and offline operation

Pin and copy WOFF2 assets for Bangers and the Fredoka family under
`go/web/public/webfonts`, with SIL OFL 1.1 notices. Add local `@font-face`
declarations while preserving the family names consumed by the theme tokens, or
update the token deliberately after visual comparison if the selected Fredoka
package exposes a different family name.

Remove Google Fonts preconnects and dynamic stylesheet injection from both
`head.tmpl` and `common.js`. A repository search for `fonts.googleapis.com` and
`fonts.gstatic.com` must return no runtime references.

## IE11 removal

Delete `alpine-ie11.min.js` and remove its `nomodule` script tag. Keep the normal
Alpine 2.8.0 build because templates still rely on Alpine directives. This task
does not add transpilation, polyfills, or alternative IE compatibility paths.

## Build integration

Make targets that compile or run Mango must generate assets before invoking Go:

- `make build`, `make static`, and `make run` depend on the asset build.
- Add explicit asset install/build/check targets for developers and CI.
- `make check` includes the non-mutating asset drift check in addition to Go vet.
- Go tests may remain independent of Node generation unless a test requires
  regenerated embedded files; the full validation sequence runs assets first.

The production Dockerfile uses a pinned Node image stage to run `npm ci` and
`npm run assets:build`. Its generated `go/web/public` tree is then copied into the
Go builder before `go build`, ensuring the binary embeds lockfile-derived assets.
The final scratch image remains unchanged in runtime composition.

## Inventory and licenses

Commit a human-readable inventory that records every third-party browser asset,
including version or unresolved provenance, source/package, license, runtime
consumer, ownership class, and retain/remove/follow-up decision. Commit the
licenses/notices needed for redistributed managed assets and locally hosted
fonts.

Unreferenced files are not silently deleted unless provenance and reachability
are sufficiently established. Disabled download-manager routes, template,
script, and translations remain present and are marked as unsupported retained
code. Their product-level removal belongs to a separate task.

## Compatibility and rollback

The supported browser target is the current and previous major Chromium,
Firefox, and Safari releases. CSS line clamping and WOFF2 local fonts are valid
within this target.

The safest rollback unit is the complete pipeline change: package metadata,
scripts, reconciled LESS/CSS, copied assets, Makefile, Dockerfile, templates, and
documentation must move together. Partial rollback could leave templates
referencing deleted files or Go embedding stale outputs.

## Validation contract

Required checks are:

```bash
npm ci
npm run assets:build
npm run assets:check
make test
make check
make build
docker build -t mango-asset-check .
```

Additional assertions:

- No Google Fonts or IE11 Alpine runtime references remain.
- Every declared generated/copied output matches the pipeline output.
- The resulting Go binary serves CSS, JS, fonts, and framework images under root
  and non-root BaseURL mounting.
- Comic/flat light/dark smoke checks show no material theme regression.
- Long titles clamp to the intended line count and expose full text through the
  native `title` attribute.
- The continuous reader still updates history as images enter the viewport.
- Unsupported download-manager files remain present but are not treated as a
  supported success flow.

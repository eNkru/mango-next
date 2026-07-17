# Research: Minimal npm pipeline, local fonts, risks, and implementation evidence

- **Query**: 为 design.md 和 implement.md 提供最小 npm 流水线、字体本地托管、IE11 路径影响、drift check、风险和 checklist 的证据。
- **Scope**: mixed
- **Date**: 2026-07-17

## Findings

### Minimal pipeline boundary

The smallest pipeline consistent with the PRD is a copy-and-compile pipeline, not a JS bundle/framework migration.

#### npm-managed dependencies

Pin the versions already committed wherever provenance is known, so the first delivery reproduces behavior rather than broadly upgrading it:

| Dependency | Initial pinned version / purpose |
|---|---|
| `less` | One pinned 4.x version as repository-local compiler; exact selected version becomes part of the lockfile contract. |
| `uikit` | `3.5.9`; supplies JS and LESS/theme assets. |
| `@fortawesome/fontawesome-free` | `5.15.4`; supplies LESS and required fonts/licenses. |
| `jquery` | `3.2.1`; copied to `public/js/jquery.min.js`. |
| `jquery-ui-dist` | `1.12.1`; copied to `public/js/jquery-ui.min.js`. |
| `alpinejs` | `2.8.0`; copy only regular `dist/alpine.js`/minified equivalent; do not copy the IE11 build. |
| `moment` | `2.24.0`; copied to `public/js/moment.min.js`. |
| `select2` | `4.1.0-beta.1`; copy paired JS/CSS. |
| `jquery-inview` | Version requires provenance decision; 1.1.2 is the published candidate but should be byte/behavior verified before declaring exact replacement. |
| `dotdotdot-js` | Pin `4.0.11` only if exact reproduction is required; its CC-BY-NC-4.0 license is a release/legal decision. A version change would be a separate compatibility/license change, not silent pipeline work. |
| Local font package/source | Bangers and Fredoka/Fredoka One package choice must preserve current family naming/appearance; licenses are OFL-1.1. |

Whether runtime libraries appear in `dependencies` or `devDependencies` does not affect the browser (files are copied and embedded), but they must be installed by production asset stages. If Docker uses `npm ci` without `--omit=dev`, either section works; recording browser inputs under `dependencies` and build-only tools under `devDependencies` makes ownership clearer.

#### Keep as project source

- All unminified Mango JS listed in `asset-inventory.md` (`common.js`, `i18n.js`, page scripts, etc.).
- All LESS files and project images/icons.
- Templates and `manifest.json`/`robots.txt`.
- Unsupported retained `download-manager.js` and its template/routes/translations, unchanged in scope.

#### Generated/copied output allowlist

- Compile: `mango.css`, `comic-theme.css`, `flat-theme.css`, `tags.css`.
- Copy: known third-party JS/CSS listed in the inventory, Font Awesome font files actually referenced by generated CSS, UIkit internal SVGs referenced by `uikit.less`, and local Bangers/Fredoka WOFF2 files.
- Do not recursively overwrite `go/web/public/js` or `img`: those directories also contain project source.

### Script shape

A minimal deterministic script set can be expressed without a bundler:

```text
assets:copy     copy an explicit file map from node_modules to go/web/public
assets:css      run local lessc for the four entries
assets:licenses copy/generate a committed third-party notice/license set
assets:build    assets:copy + assets:css + assets:licenses
assets:check    build in an isolated temporary tree, then compare declared outputs
```

Important properties evidenced by the repository:

- Use Node filesystem APIs or a small checked-in script rather than platform-specific `cp`/`diff` if Windows developer support matters; otherwise Make/Docker currently assume Unix tools.
- The copy map must be explicit and fail when an upstream file path disappears.
- The CSS compiler must see one consistent npm root. `mango.less`'s current mixed import paths need a deterministic resolution rule during implementation.
- A drift check should compare only declared generated/copied outputs. Comparing the entire `public` directory would incorrectly treat project-edited JS/images as generated.
- `assets:check` should not leave a dirty worktree; build to a temp directory or snapshot/restore outputs, then byte-compare.

### Local Bangers and Fredoka hosting

Current runtime requests occur in two places:

- `go/web/views/head.tmpl:8-9,24-29` preconnects to Google and dynamically injects the stylesheet when comic is active.
- `go/web/public/js/common.js:185-191` dynamically injects the same Google Fonts URL when switching to comic.

Current CSS names:

- `_variables.less:66`: `'Fredoka One'`
- `_variables.less:67`: `'Bangers'`
- `comic-theme.less` repeatedly consumes `@font-comic`; `@font-sound` is defined but no direct use was found in the inspected matches.

Local-hosting asset contract:

1. Copy pinned WOFF2 files into `go/web/public/webfonts/`.
2. Add project-owned `@font-face` declarations in an always-loaded compiled stylesheet (or a dedicated local CSS referenced from `head.tmpl`).
3. Preserve the CSS family names expected by `_variables.less`, or deliberately update the token if the chosen Fredoka package no longer exposes “Fredoka One”.
4. Remove both Google preconnects and both dynamic Google stylesheet injection paths; removing only `head.tmpl` leaves `common.js` network access.
5. Include OFL-1.1 license text with redistributed fonts.

License evidence:

- Google Fonts `ofl/bangers/OFL.txt` identifies Bangers as SIL Open Font License 1.1.
- Current Google Fonts repository uses `ofl/fredoka/Fredoka[wdth,wght].ttf` and `ofl/fredoka/OFL.txt`, also OFL-1.1. The current UI requests the older Google Fonts family name `Fredoka One`; the exact static file/package preserving that name/version is not determined by this repository.

### IE11 Alpine removal impact

Scoped files/lines:

- Delete managed artifact `go/web/public/js/alpine-ie11.min.js`.
- Remove `go/web/views/head.tmpl:43` (`nomodule` script).
- Keep `go/web/public/js/alpine.min.js` and `head.tmpl:42` because Alpine directives remain pervasive.
- `head.tmpl:4` has `<meta http-equiv="X-UA-Compatible" content="IE=edge">`; it is IE-oriented metadata but is not an Alpine asset path. Whether to remove it is presentation cleanup rather than required dependency removal.

License effect: the IE11 file embeds multiple polyfills/notices beyond Alpine. Removing the file removes the need to redistribute those notices for that artifact, while Alpine's MIT notice remains required for the regular build.

### Concrete risks for design.md

| Risk | Repository evidence | Validation / containment evidence |
|---|---|---|
| CSS source/output divergence changes UI when pipeline first runs | Measured diffs for comic, flat, tags; spec permits hand sync | Before automation, reconcile current runtime CSS back into LESS or explicitly designate another source. Visual/browser smoke across comic/flat light/dark is needed. |
| Main LESS cannot resolve dependencies | `mango.less:5-7` and `uikit.less:1` use inconsistent node_modules paths; clean compile failed | One npm root and deterministic include/import paths; run CSS build from clean checkout/container. |
| Third-party copy silently changes bytes | Existing assets came from jsDelivr/package dist and some lack banners | Pin exact versions/lockfile; explicit paths; compare known headers/hashes in first migration. |
| dotdotdot license conflicts with expected distribution policy | `dotdotdot.js:9-10` is CC-BY-NC-4.0 | Record explicit retain/replace/legal decision before release; preserve attribution meanwhile. |
| Font family mismatch changes comic visual intent | CSS asks for `Fredoka One`; current Google Fonts repo/package ecosystem may expose `Fredoka` | Inspect font metadata and browser computed family; screenshot comic headings/cards before/after. |
| Runtime CDN remains after font work | Two independent injection paths (`head.tmpl`, `common.js`) | `rg 'fonts\.googleapis\.com|fonts\.gstatic\.com' go/web` must be empty; browser network check offline. |
| Removing IE11 file but retaining template reference causes 404 | `head.tmpl:43` references exact file | Remove copy and reference atomically; static asset request smoke. |
| Docker embeds stale committed assets | Current Docker copies only `go/` and builds immediately | Node lockfile stage must precede Go build and pass generated public tree forward. |
| `go:embed` masks generation order errors | Embed occurs at compile time | Build test should modify/regenerate then compile; inspect served asset/hash from resulting binary. |
| Unsupported download page is accidentally “validated” as supported or deleted | Archived queue PRD says disabled/unrepaired; current files/routes exist | Inventory label “unsupported retained code”; existence checks only, no success-flow acceptance in this task. |
| Font Awesome generated CSS references absent formats | `mango.css:17144-17145` references missing solid EOT/SVG | Generate from package consistently and copy only/reference existing formats, then verify icons in supported browsers. |
| BaseURL static/PWA behavior regresses | Templates use `.BaseURL`; manifest uses root-absolute URLs | Build/run under non-root BaseURL and request CSS/JS/fonts/icons explicitly. |

### Design checklist

- [ ] Define the three ownership classes: project source, npm-managed copies, generated CSS.
- [ ] List exact first-pass dependency pins and defer broad upgrades.
- [ ] Resolve `mango.less` npm import root deterministically.
- [ ] Declare four CSS entrypoints and an explicit copy allowlist.
- [ ] Decide how current CSS-only changes are migrated back to LESS before turning on drift checks.
- [ ] Define clean install (`npm ci`), build, and non-mutating drift-check commands.
- [ ] Define license/NOTICE output, including Font Awesome composite terms, font OFL texts, and dotdotdot's CC-BY-NC caveat.
- [ ] Define local Bangers/Fredoka file/family mapping and offline network acceptance.
- [ ] Define IE11 artifact/reference removal while retaining normal Alpine 2.8.0.
- [ ] Mark download-manager route/template/script/translations as unsupported retained code.
- [ ] Put Node asset generation before Go compile in Make and Docker.
- [ ] Preserve committed generated assets so ordinary binary consumers do not need Node at runtime.

### Implementation checklist

- [ ] Add package manifest and committed npm lockfile; verify `npm ci` on clean checkout.
- [ ] Add deterministic copy/compiler scripts with explicit input/output maps.
- [ ] Reconcile LESS/CSS drift, then reproduce all four CSS files byte-for-byte under the chosen formatting contract.
- [ ] Copy pinned third-party JS/CSS/fonts from packages; retain project JS untouched.
- [ ] Add/copy third-party license texts and inventory metadata.
- [ ] Vendor Bangers and Fredoka WOFF2; add `@font-face`; remove Google preconnect/injection in both template and JS.
- [ ] Remove `alpine-ie11.min.js` and its `nomodule` tag only.
- [ ] Keep unsupported download-manager assets/routes/translations and label them in inventory/docs.
- [ ] Wire asset build before `make build`, `make static`, `make run` as chosen by design; expose a check target.
- [ ] Add Node stage to Docker, use `npm ci`, generate assets, then copy the final Go tree/public outputs into Go builder.
- [ ] Update README and frontend guide with Node prerequisite, generated-file ownership, command order, and offline behavior.
- [ ] Run drift, Go tests/vet/build, Docker build, BaseURL static requests, offline network inspection, and theme smoke.

## External References

- [Bangers OFL](https://github.com/google/fonts/blob/main/ofl/bangers/OFL.txt) — SIL OFL 1.1 and copyright notice.
- [Fredoka files and OFL](https://github.com/google/fonts/tree/main/ofl/fredoka) — current Google Fonts family files/license.
- npm registry references listed in `asset-inventory.md` — pinned package provenance/licenses.

## Related Specs

- `.trellis/tasks/07-17-frontend-asset-pipeline/prd.md` — authoritative scope: pipeline first, local fonts, no IE11, no broad runtime upgrades, retain disabled download code.
- `.trellis/spec/frontend/ui-theme-layout.md` — theme runtime behavior and smoke checklist.
- `.trellis/tasks/archive/2026-07/07-17-frontend-browser-smoke/prd.md` — supported browser baseline; implementation was closed/deferred, so this task cannot rely on an existing browser runner.

## Caveats / Not Found

- The exact npm/static-font package and version that reproduces the currently requested “Fredoka One” glyphs is not provable from the repository; current Google Fonts has evolved to the `Fredoka` family.
- No legal/product decision on retaining CC-BY-NC dotdotdot is present.
- No CI workflow exists to host drift checks yet; the sibling CI task owns shared workflow composition.
- Exact desired Make target semantics (automatic build on every `make run` versus separate explicit asset prerequisite) are design decisions, not facts recoverable from current code.

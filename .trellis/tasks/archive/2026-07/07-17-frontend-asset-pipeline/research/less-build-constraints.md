# Research: LESS graph, generated drift, and build-order constraints

- **Query**: 分析 LESS import 图、应生成的 CSS 入口，以及 Makefile、Dockerfile、文档和 `go:embed` 对构建顺序的约束。
- **Scope**: internal
- **Date**: 2026-07-17

## Findings

### LESS import graph

```text
mango.less                         (runtime entry -> mango.css)
├── ./uikit.less
│   └── node_modules/uikit/src/less/uikit.theme.less
├── ../../node_modules/@fortawesome/fontawesome-free/less/fontawesome.less
├── ../../node_modules/@fortawesome/fontawesome-free/less/solid.less
├── ../../node_modules/@fortawesome/fontawesome-free/less/brands.less
└── ./_variables.less

comic-theme.less                  (runtime entry -> comic-theme.css)
└── ./_variables.less

flat-theme.less                   (runtime entry -> flat-theme.css)
└── ./_variables.less

tags.less                         (runtime entry -> tags.css)
└── ./_variables.less
```

Evidence: `mango.less:2,5-7,18`, `uikit.less:1`, `comic-theme.less:15`, `flat-theme.less:3`, `tags.less:3`.

The four CSS outputs loaded by templates are exactly:

1. `mango.css` — `head.tmpl:35`
2. `comic-theme.css` — `head.tmpl:36`
3. `flat-theme.css` — `head.tmpl:37`
4. `tags.css` — `title.tmpl:194`

`select2.min.css` is copied third-party output rather than a LESS output in this repository.

### Import-path constraint

- `uikit.less:1` expects `node_modules/uikit` relative to `go/web/public/css`, resolving to `go/web/public/css/node_modules/...` under ordinary Less resolution unless an include path is supplied.
- Font Awesome imports use `../../node_modules/...` from `go/web/public/css/mango.less`, resolving to `go/web/node_modules/...`.
- These paths do not share one natural `node_modules` root. A root-level npm manifest alone does not satisfy the literal Font Awesome path without either a documented Less include/import strategy or source import-path normalization in implementation.
- A direct temporary compile with npm-exec packages failed at `mango.less:5` because `go/web/node_modules/@fortawesome/...` did not exist. This is concrete evidence that the current source is not compilable from a clean checkout as-is.

### Measured generated-file drift

Using Less 4.2.2 against the entries that do not require external imports:

- `comic-theme.less` does **not** reproduce committed `comic-theme.css`. The committed CSS contains later hand-added rules absent from LESS, including library hide-toggle/card action blocks and moved progress-percent rules. The first observed divergence is around committed CSS line 138.
- `flat-theme.less` does **not** reproduce committed `flat-theme.css`; the diff is very large. The CSS contains extensive later batches and overrides not represented in the 365-line LESS source.
- `tags.less` does **not** reproduce committed `tags.css`. Differences include token values and selector nesting direction (`body ... .uk-light` versus compiled `.uk-light body ...`), beginning at line 1.
- `mango.less` could not be clean-compiled due to the unresolved Font Awesome import path noted above.

This confirms actual drift, not only theoretical risk. The current generated CSS files are partly independent hand-maintained runtime sources.

### Historical evidence

- `.trellis/spec/frontend/ui-theme-layout.md:49` says: `Compile: lessc comic-theme.less comic-theme.css. Prefer lessc when available; flat-theme.css may be hand-synced...`.
- Git history shows CSS and LESS often changed in different commit sets. Recent `comic-theme.css` commits (`d8275c9`, `35193aa`, `7cb5e30`, `261288e`) are absent from `comic-theme.less` history, directly explaining the measured drift.
- `FRONTEND_DEV_GUIDE.md:20` says LESS may exist and compiled CSS is committed, but gives no repository-local compile command.

### Build-order constraints

Current order and missing stages:

| Path | Current behavior | Constraint |
|---|---|---|
| `Makefile:10-17` | `build`, `static`, and `run` invoke Go directly | No npm install, asset generation, or drift validation precedes Go compilation. |
| `Makefile:8` | `all: check test build` | Even the aggregate target does not verify frontend outputs. |
| `Dockerfile:5-12` | Single Go builder; `COPY go/ ./`; then `go build` | Docker context does not copy a package manifest/lockfile and has no Node stage. Generated files must already be committed and correct. |
| `FRONTEND_DEV_GUIDE.md:3,24-34,58-59` | States no Node pipeline; instructs editing public files and rebuilding | Documentation currently treats committed CSS/JS as direct inputs and only documents the embed rebuild. |
| `README.md:30-40,58-71` | Requires only Go; direct `make run` / `go build` | Any npm stage changes prerequisites and must be reflected here. |
| `go/web/embed.go:12-13` | `//go:embed public/*` | Asset generation/copy must finish before `go build`/`go run`; changing files after compilation cannot affect the binary. |

Required dependency order implied by these facts:

```text
npm clean install from lockfile
  -> copy managed third-party browser files/fonts into go/web/public
  -> compile all declared LESS entries into committed CSS
  -> drift check (when validating clean tree)
  -> go test/vet/build
  -> Go embeds final public tree
```

For Docker, the same order requires a Node asset stage before the Go stage, with the generated `go/web/public` tree copied into the Go build context before `go build`.

### Derived-file ownership evidence

The repository currently mixes three ownership classes in `go/web/public`:

1. **Project source** — templates, project JS, LESS, project images.
2. **Managed copies** — minified third-party JS/CSS/fonts and UIkit internal SVGs.
3. **Generated outputs** — four compiled CSS entries.

Because all three are embedded from one directory, a pipeline needs an explicit allowlist/manifest rather than deleting or replacing the whole public tree.

## Suggested validation commands for design evidence

These are command shapes for `design.md`/`implement.md`; exact script names are an implementation choice.

```bash
npm ci
npm run assets:build
npm run assets:check
git diff --exit-code -- go/web/public/css go/web/public/js go/web/public/webfonts go/web/public/img
make test
make check
make build
docker build -t mango-asset-check .
```

Additional repository-specific checks:

```bash
# No runtime CDN font dependency remains
rg 'fonts\.googleapis\.com|fonts\.gstatic\.com' go/web

# No IE11 Alpine path remains after scoped removal
rg 'alpine-ie11|nomodule' go/web

# Every template static reference resolves under public (script can account for BaseURL)
rg 'BaseURL}}(js|css|img|webfonts)/' go/web/views

# Unsupported retained download files still exist
test -f go/web/views/download-manager.tmpl
test -f go/web/public/js/download-manager.js
```

## Related Specs

- `.trellis/spec/frontend/ui-theme-layout.md:42-50` — current theme file map and manual compile guidance.
- `FRONTEND_DEV_GUIDE.md:3-34,56-59` — current frontend build contract.
- `.trellis/tasks/07-17-frontend-asset-pipeline/prd.md:31-37,56-57` — lockfile, deterministic generation, drift, and Docker requirements.

## Caveats / Not Found

- No existing package manifest, lockfile, frontend build script, or CI workflow exists.
- The exact Less version originally used to create each committed CSS file is not recorded. The structural differences are substantive and cannot be explained by formatting/version differences alone.
- A clean `mango.less` reproduction cannot be measured until its external import resolution is made deterministic.

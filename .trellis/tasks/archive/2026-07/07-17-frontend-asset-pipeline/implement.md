# Frontend asset dependency pipeline implementation plan

## 1. Establish the inventory baseline

- [x] Convert the research inventory into a committed dependency/asset inventory
      with ownership class, version/provenance confidence, source, license,
      consumers, and disposition.
- [x] Record unsupported retained download-manager code and unresolved manual
      vendor files without claiming false provenance.
- [x] Define the exact managed copy allowlist and generated CSS output list.

Validation:

```bash
test -f go/web/views/download-manager.tmpl
test -f go/web/public/js/download-manager.js
```

## 2. Add the npm toolchain

- [x] Add root `package.json` and `package-lock.json` with pinned current runtime
      inputs and a pinned LESS compiler.
- [x] Add repository-local Node asset scripts implementing explicit copies,
      license generation, isolated builds, and byte comparison.
- [x] Add `node_modules/` to `.gitignore`.
- [x] Verify `npm ci` succeeds from a clean dependency state.

Validation:

```bash
npm ci
npm run assets:copy
```

Rollback point: package metadata and scripts can be removed before generated
outputs or build entrypoints are switched.

## 3. Reconcile LESS and generated CSS

- [x] Normalize `mango.less` and `uikit.less` imports to the root npm dependency
      tree.
- [x] Port existing CSS-only runtime rules back into `comic-theme.less`,
      `flat-theme.less`, and `tags.less`.
- [x] Compile `mango.css`, `comic-theme.css`, `flat-theme.css`, and `tags.css`
      through `assets:css`.
- [x] Confirm a second isolated build is byte-identical.
- [x] Update the frontend theme spec to prohibit hand-synchronizing generated
      CSS and document the canonical command.

Validation:

```bash
npm run assets:css
npm run assets:check
```

Review gate: compare comic/flat light/dark rendering before continuing. If the
reconciled source changes runtime behavior, fix LESS rather than hand-editing CSS.

## 4. Manage third-party copies and licenses

- [x] Copy known pinned JS/CSS/framework image/font assets through the allowlist.
- [x] Keep unresolved `jquery.inview.min.js` as a documented manual vendor file
      unless exact provenance is proven during implementation.
- [x] Generate or copy required notices for UIkit, Font Awesome, Alpine, jQuery,
      jQuery UI, Moment, Select2, LESS, and managed fonts.
- [x] Ensure generated Font Awesome CSS references only files included in the
      public tree for supported browsers.

Validation:

```bash
npm run assets:build
npm run assets:check
```

## 5. Remove dotdotdot and its jQuery path

- [x] Add native `title` attributes to card titles rendered by home, library,
      tag, and title templates where the full title is available.
- [x] Ensure CSS line clamps reserve the intended two or three lines across both
      themes and supported breakpoints.
- [x] Remove `dots.tmpl`, `dots.js`, and `dotdotdot.js`, and remove template
      inclusions of the partial.
- [x] Stop loading `jquery.inview.min.js` through the removed dots partial, while
      retaining the reader-specific include and behavior.
- [x] Remove dotdotdot from inventory and license outputs; document that the NC
      dependency was eliminated.

Validation:

```bash
rg 'dotdotdot|template "dots"|js/dots\.js' go/web
rg 'jquery\.inview' go/web/views go/web/public/js
```

Expected result: no dotdotdot/dots matches; jquery-inview remains only in the
reader flow.

## 6. Localize fonts and remove IE11 assets

- [x] Add managed Bangers and Fredoka WOFF2 files and OFL notices.
- [x] Add local `@font-face` declarations and preserve the comic theme's intended
      family mapping after visual comparison.
- [x] Remove Google Fonts preconnect and dynamic injection paths from
      `head.tmpl` and `common.js`.
- [x] Delete `alpine-ie11.min.js` and remove the `nomodule` tag; retain regular
      Alpine.

Validation:

```bash
! rg 'fonts\.googleapis\.com|fonts\.gstatic\.com' go/web
! rg 'alpine-ie11|nomodule' go/web
```

## 7. Integrate Make and Docker builds

- [x] Add explicit Make targets for npm install, asset build, and asset check.
- [x] Make `build`, `static`, and `run` generate assets before Go compilation.
- [x] Make `check` run the non-mutating asset check and Go vet.
- [x] Add a Node asset stage to Docker, then copy generated public assets into
      the Go builder before `go build`.
- [x] Preserve the final scratch image and offline runtime behavior.

Validation:

```bash
make check
make test
make build
docker build -t mango-asset-check .
```

Rollback point: Makefile and Dockerfile must be reverted together with package
and generation changes to avoid embedding stale or missing assets.

## 8. Document usage and ownership

- [x] Update README prerequisites and build commands to include Node/npm.
- [x] Rewrite `FRONTEND_DEV_GUIDE.md` around project source, managed copies, and
      generated outputs.
- [x] Document `npm ci`, asset build, drift check, Go embed ordering, Docker
      generation, and the prohibition on hand-editing generated CSS.
- [x] Document retained dependency risks and follow-up work for broad upgrades
      and full jQuery removal.

## 9. Full-scope verification

- [x] Run clean npm installation, deterministic build, and drift check.
- [x] Run Go tests, vet, and build.
- [x] Build the production Docker image.
- [x] Resolved review regressions: the asset script is trackable, standalone
      license generation writes only to `THIRD_PARTY_LICENSES/`, the stale
      comic-font export is removed, generated theme rules are restored to LESS,
      `make -j all` serializes check/test/build, and the manual jquery-inview
      vendor has an ISC notice.
- [ ] Verify static assets and local fonts under root and non-root BaseURL.
- [ ] Smoke-test comic/flat light/dark themes and card truncation/title behavior
      on current Chromium, Firefox, and Safari where available.
- [ ] Verify continuous reader history/progress behavior still works because its
      jquery-inview dependency remains.
- [ ] Verify browser network activity contains no public font/CDN requests.
- [ ] Run the Trellis quality check and resolve all findings before requesting
      implementation completion.

Commands:

```bash
npm ci
npm run assets:build
npm run assets:check
make test
make check
make build
docker build -t mango-asset-check .
```

## Follow-up boundaries

- Broad upgrades of jQuery, Alpine, UIkit, Select2, Moment, and other retained
  runtime libraries are separate compatibility tasks.
- Full jQuery removal should inventory all `$` usage, replace the reader's
  `jquery.inview` behavior with `IntersectionObserver`, and replace or remove
  jQuery UI/Select2 dependencies before deleting global jQuery.
- Product-level removal of disabled download routes, templates, scripts, and
  translations is a separate cleanup task.

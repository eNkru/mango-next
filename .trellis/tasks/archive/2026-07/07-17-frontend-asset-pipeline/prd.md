# Frontend asset dependency pipeline

## Goal

Make committed browser assets traceable, reproducible, maintainable, and locally
served without changing Mango's server-rendered architecture.

## Closure

Closed and redirected before treating the jQuery/LESS inventory pipeline as the
final architecture. Useful outcomes (npm lockfile discipline, Make/Docker asset
generation order, local fonts, offline delivery) are absorbed by
`07-17-frontend-react-vite` / `07-18-frontend-react-foundation`.

## Confirmed Facts

- There is no frontend package manifest or lockfile.
- LESS sources and compiled CSS coexist under `go/web/public/css`.
- Several minified libraries are checked in directly and are years behind current
  releases; version metadata is incomplete for some files.
- `go:embed` consumes `go/web/public` at Go compile time, so generated runtime
  files must exist before the Go build.
- The Makefile and Dockerfile currently build Go directly and do not generate or
  validate frontend assets first.
- The documented theme workflow permits hand-synchronizing compiled CSS, and past
  work records CSS drift caused by the absence of a repository-local `lessc`.
- `head.tmpl` still loads an IE11-specific Alpine build and fetches comic fonts
  from Google Fonts at runtime, contrary to the target browser and self-hosting
  requirements.
- The download-manager route, template, script, and translations remain present
  even though the download feature is disabled; removal requires an explicit
  product-scope decision rather than an asset-only assumption.

## Requirements

- Inventory every vendored JavaScript, CSS, font, and derived asset with version,
  license, source, runtime consumer, and upgrade/remove/retain decision.
- Introduce a repository-local manifest and lockfile for build-time tooling and
  managed browser dependencies.
- Use npm with a committed `package-lock.json`; clean installs use `npm ci` and
  do not require pnpm, Yarn, or globally installed build tools.
- Provide deterministic commands to build committed runtime CSS/JS assets and to
  fail when generated outputs drift.
- Preserve local runtime delivery and offline/self-hosted operation.
- Vendor the Bangers and Fredoka comic-theme fonts as managed local assets,
  preserving the current visual intent while removing Google Fonts runtime
  requests and recording their licenses.
- Upgrade dependencies incrementally, documenting compatibility or security risk
  for anything intentionally retained.
- Prioritize the reproducible pipeline in the first delivery. Do not combine it
  with broad upgrades of jQuery, Alpine, UIkit, Select2, or other runtime
  libraries; record retained-version risks and split upgrades into follow-up
  work.
- Remove `dotdotdot.js` instead of managing or upgrading it. Use the existing
  CSS multi-line clamping contract for card titles, eliminating its restrictive
  CC-BY-NC-4.0 license and one jQuery-dependent behavior without expanding this
  task into a full jQuery migration.
- Preserve full truncated card titles with native HTML `title` attributes; do
  not add JavaScript overflow measurement or replacement tooltip libraries.
- Keep `jquery.inview.min.js` only for the reader's existing continuous-mode
  behavior. Record `IntersectionObserver` as the intended replacement when the
  broader jQuery-removal work is planned.
- Target the current and previous major releases of Chromium, Firefox, and
  Safari; remove the obsolete IE11-specific Alpine build and compatibility path.
- Audit template references and remove unreachable assets or stale MangaDex
  browser calls that are not owned by the queue child task.
- Record disabled download-manager assets as unsupported retained code, but do
  not remove its server route, template, script, or translations in this task;
  complete removal belongs to a separate product-cleanup task.
- Document developer and CI commands, including prerequisites and generated-file
  ownership.
- Make the production Docker build regenerate frontend assets from
  `package-lock.json` in a Node stage before the Go build embeds them.

## Acceptance Criteria

- [ ] A dependency inventory accounts for every committed third-party browser
      asset and its license.
- [ ] A clean checkout installs from a lockfile and reproduces all declared
      generated runtime assets.
- [ ] A drift-check command fails when committed outputs differ from generated
      sources.
- [ ] The Go binary continues to embed and serve the expected assets without
      runtime CDN requests.
- [ ] Legacy dependencies are upgraded, removed, or accompanied by an explicit
      compatibility/security decision.
- [ ] Developer documentation and the Docker/build path describe the correct
      asset generation order.

## Dependencies

- `07-17-frontend-queue-contract` was closed without implementation because the
  download feature is disabled. This task must not preserve its broken browser
  integration as a supported flow.
- Owns frontend asset commands; `07-17-test-ci-baseline` owns shared workflow
  composition.

## Open Questions

- None blocking technical design. Exact retained dependency versions and file
  provenance are research outputs, not product decisions.

## Out of Scope

- A framework rewrite, JavaScript bundling for its own sake, or moving runtime
  assets to a CDN.
- Internet Explorer 11 compatibility, transpilation, or polyfills.
- Unrelated UI redesign.

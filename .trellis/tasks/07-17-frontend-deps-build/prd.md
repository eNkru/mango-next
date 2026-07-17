# Frontend dependencies and build pipeline

## Goal

Repair broken browser/backend contracts, then make frontend assets reproducible,
supportable, and testable while preserving the server-rendered application and
its offline/self-hosted deployment model.

## Requirements

- Inventory every vendored browser dependency, its version, license, runtime use,
  and upgrade/removal path.
- Reconcile the download-manager page with the Go queue API: route names,
  supported actions, request bodies, response envelope/field names, polling or
  push behavior, and error states must have one tested contract.
- Remove or replace stale browser calls to deleted MangaDex endpoints and other
  unreachable runtime assets discovered by a template-to-route/asset audit.
- Upgrade or replace legacy dependencies incrementally, prioritizing security and
  compatibility over a framework rewrite.
- Introduce a locked, documented asset build that produces the embedded CSS/JS
  files from source and detects generated-file drift.
- Preserve local asset serving with no public-CDN runtime dependency.
- Add browser smoke coverage for login, library browsing, title actions, reader
  navigation, admin actions, responsive layouts, and both supported UI themes.
- Improve accessibility issues encountered while touching affected components,
  especially keyboard semantics, labels, image alternatives, and icon buttons.
- Define how localization keys and compiled assets are validated in CI.

## Acceptance Criteria

- [ ] The repository has a dependency manifest and lockfile or an equivalently
      reproducible, documented asset toolchain.
- [ ] Download-manager list/delete/retry behavior works against the current Go
      API, and unsupported pause/resume/WebSocket UI is removed or implemented
      end to end.
- [ ] jQuery 3.2.1, Alpine.js 2.8.0, Moment, UIkit, Select2, and other vendored
      libraries are either upgraded, replaced, or explicitly justified with a
      tracked risk decision.
- [ ] A clean checkout can regenerate runtime assets without fetching from CDNs
      in the browser.
- [ ] CI fails when committed generated assets drift from their sources.
- [ ] Browser smoke tests pass at representative desktop and mobile viewports for
      both UI themes.
- [ ] Existing Go embed/build behavior remains intact.

## Notes

- Dependency: no hard dependency. The broken download-manager contract is the P1
  first milestone; dependency modernization remains P2 and is recommended after
  authentication hardening so browser flows target settled auth behavior.
- Coordination: shared CI wiring belongs to `07-17-test-ci-baseline`; this task
  owns the frontend runner, fixtures, selectors, and asset commands.
- Constraint: do not replace the server-rendered architecture merely to modernize
  tooling.

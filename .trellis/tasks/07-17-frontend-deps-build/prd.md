# Frontend dependencies and build pipeline

## Goal

Coordinate the repair and reproducibility of the embedded frontend while
preserving Mango's server-rendered, self-hosted architecture.

## Confirmed Facts

- Go embeds templates and runtime assets from `go/web/`; there is currently no
  frontend package manifest, lockfile, or CI workflow.
- The download manager calls deleted `/api/admin/mangadex/queue` HTTP and
  WebSocket endpoints, while the Go server exposes `/api/admin/queue` with only
  list, delete, and retry operations.
- Runtime dependencies are committed as minified files, including jQuery 3.2.1,
  Alpine.js 2.8.0, UIkit 3.5.9, jQuery UI 1.12.1, and Select2 4.1.0-beta.1.
- The repository contains LESS sources alongside committed runtime CSS, but no
  documented command that reproduces all generated assets.

## Task Map

- `07-17-frontend-queue-contract` (P1): closed before implementation because the
  download feature is disabled.
- `07-17-frontend-asset-pipeline` (P2): redirected into the React + Vite
  foundation under `07-17-frontend-react-vite`; do not finish the old
  jQuery/LESS inventory end state as the final architecture.
- `07-17-frontend-react-vite` (P1): React + Vite frontend migration parent.

## Requirements

- Each child task must remain independently implementable, testable, and
  reviewable.
- Browser compatibility targets the current and previous major releases of
  Chromium, Firefox, and Safari; Internet Explorer 11 is no longer supported.
- Child dependencies and CI ownership must be explicit in child artifacts.
- Runtime assets must remain locally served; public CDNs must not be required.
- The Go embed/build behavior and server-rendered templates must remain intact.
- Dependency modernization must be incremental rather than a framework rewrite.

## Cross-Task Acceptance Criteria

- [ ] The active asset-pipeline child task satisfies its acceptance criteria.
- [ ] A clean checkout can build the Go binary and reproduce its embedded
      frontend assets using documented commands.
- [ ] Frontend validation covers generated-file drift without duplicating
      ownership of shared CI wiring.
- [ ] No supported runtime flow references deleted MangaDex endpoints or public
      CDN assets.

## Out of Scope

- Replacing Go templates with a client-side application framework.
- Redesigning the application or adding unrelated frontend features.
- Browser automation and end-to-end smoke coverage; the planned child task was
  closed before implementation because the project is not ready for that level
  of testing.
- Owning the repository-wide CI baseline, which remains in
  `07-17-test-ci-baseline`.

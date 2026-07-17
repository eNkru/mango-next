# Configuration deployment and documentation cleanup

## Goal

Make configuration and deployment behavior truthful, reproducible, and safe for
the documented Docker, QNAP, and direct-binary workflows.

## Confirmed Facts

- Unused (loaded but not applied at runtime): `session_secret`, `log_level`,
  `download_timeout_seconds`, `cache_size_mbs`, `cache_log_enabled`.
  `cache_enabled` is loaded but not referenced outside config package.
- `base_url` is normalized and used as a URL prefix in templates/API JSON/cookie
  Path, but routes are registered only at `/` and auth redirects hard-code
  `/login` — non-root values produce broken links.
- Default Compose: obsolete `version: '3'`; `env.example` leaves bind paths empty
  so `docker compose config` fails; host `${PORT}` maps to container `9000`
  without injecting PORT into the process.
- QNAP Compose variants still use obsolete `version` keys; prebuilt docs still
  mention `src/config.cr`.
- `Makefile`: `all` is build-only; `go-all` runs check+test+build. README claims
  `make all` does check/test/build.
- API page (`api.tmpl`) references missing `openapi.json` and
  `js/redoc.standalone.js`; no OpenAPI route.
- Top-level `spec/` still has Crystal `*_spec.cr` with no Crystal implementation.
- Dockerfile is multi-stage scratch with `ENV HOME=/root` (correct).
- Parent review archived at
  `.trellis/tasks/archive/2026-07/07-17-project-wide-review/`.
- Auth/HTTP hardening already documented reverse-proxy cookie/proxy-header trust
  in README and DEPLOY_QNAP.

## Requirements

- Classify and act on config fields:
  - Wire `download_timeout_seconds` into plugin HTTP clients/downloaders.
  - Wire `log_level` into process logging behavior.
  - Keep parsing `session_secret`, `cache_size_mbs`, and `cache_log_enabled` for
    config compatibility but document them as unused/deprecated in Go.
  - Verify `cache_enabled`: honor it in library cache load/save or document and
    treat as deprecated if no safe wiring exists without larger redesign.
- Implement full non-root `base_url` support: mount all routes under the
  configured base path (normalized with trailing slash), keep cookie Path and
  template/API URL prefixes consistent, and fix auth redirects that currently
  hard-code `/login`.
- Align README, QNAP, Makefile, `env.example`, Compose files, and Trellis backend
  specs with the Go-only implementation.
- Remove or hide the broken API documentation page (`/api` ReDoc + missing
  assets). Do not present a non-working OpenAPI UI; document that the JSON API
  is the HTTP surface until a future OpenAPI task.
- Default Compose: fill `env.example` with copyable `./data` and `./config`
  placeholders; document `cp env.example .env`; remove obsolete `version` keys
  from maintained Compose files; keep container listen port 9000 with host
  publish `${PORT}:9000`.
- Delete the dead top-level Crystal `spec/` tree; refresh Trellis backend specs
  that still prescribe Crystal commands.
- Document persistence mounts, ports, first-admin credentials, upgrade/backup/
  rollback for direct binary and container workflows with disposable-data smoke
  guidance.
- Improve Docker layer caching / pin notes only where it materially helps
  reproducibility without making NAS deployment harder.

## Acceptance Criteria

- [x] Every documented configuration key has a tested runtime effect or is clearly
      marked deprecated/removed.
- [x] Non-root `base_url` (e.g. `/mango/`) serves pages, static assets, and API
      routes under that prefix; login redirects and cookies remain consistent;
      covered by focused tests.
- [x] `download_timeout_seconds` and `log_level` affect runtime behavior with
      tests or observable verification.
- [x] `docker compose config` succeeds for the default example (with env.example
      values) and all maintained QNAP variants without obsolete-schema warnings.
- [x] A clean direct-binary and container startup use documented persistent paths
      and ports.
- [x] README and deployment commands reference only current Go paths and behavior.
- [x] The broken ReDoc API docs page is no longer presented as a working feature
      (route/nav removed or replaced with a non-misleading message); missing
      openapi/redoc assets are not required for a successful build.
- [x] `make all` behavior matches its documentation and the planned CI quality
      gate command names (`go-all` alias retained if useful).
- [x] Top-level Crystal `spec/` is removed; docs and `.trellis/spec/backend` no
      longer instruct Crystal checks as the current gate.
- [x] Build, upgrade, backup, rollback, and first-admin credential handling are
      documented; disposable smoke steps are listed where Docker/local sockets
      allow.

## Out of Scope

- Full CI workflow authoring (owned by `07-17-test-ci-baseline`; only align
  Makefile/docs command names with whatever CI will run).
- Frontend dependency modernization (owned by `07-17-frontend-deps-build`).
- Switching away from scratch image or non-root runtime until NAS bind-mount
  permissions are validated.
- Full OpenAPI generation pipeline / embedding ReDoc (deferred after removing
  the broken page).
- Archive decompression size policy and other deferred review items.

## Open Questions

- None currently blocking planning.

## Notes

- Dependency: none for starting. Synchronize final Makefile/README quality
  commands with `07-17-test-ci-baseline` when that task lands CI.
- Constraint: do not switch the scratch image away from current pure-Go behavior
  or force a non-root runtime until NAS bind-mount permissions are validated.

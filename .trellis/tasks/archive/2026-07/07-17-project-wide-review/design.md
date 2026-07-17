# Project-wide review design

## Review boundary

This task produces an evidence-backed assessment and follow-up roadmap. It does
not change product code. Findings are limited to the repository, disposable
local validation, and public build/deployment contracts; production NAS data and
destructive testing are excluded.

## Architecture map

```text
cmd/mango
  -> config + SQLite storage/queue
  -> library scanner/cache + archive/image readers
  -> background runner
       -> thumbnail generation
       -> plugin updater/downloader (goja + HTTP)
  -> chi HTTP server
       -> auth/admin middleware
       -> page and JSON handlers
       -> embedded html/template views and static assets

Deployment: direct binary | scratch Docker image | Compose | QNAP variants
```

The application is a Go-only, server-rendered monolith with well-defined internal
packages. SQLite is the durable source for users, metadata, progress, dimensions,
and the download queue; the library filesystem and JSON metadata remain part of
the data contract.

## Evidence model

Findings are classified as:

- Confirmed: directly traced in code/config or reproduced by a local command.
- Strong: supported by code and known behavior but missing runtime reproduction.
- Candidate: modernization or defense-in-depth work requiring a dedicated task.

Critical and high-priority findings require a concrete file/behavior reference.
Tool output is supporting evidence, not a substitute for tracing the relevant
call path.

## Current strengths

- Go test, vet, build, and race checks pass.
- Library behavior has strong focused coverage and recent regression tests.
- SQL values are generally parameterized, migrations are transactional, and
  SQLite foreign keys are enabled.
- Templates use `html/template`, static assets are local/embedded, and admin
  routes are grouped behind explicit middleware.
- Background workers observe cancellation and the pure-Go build supports a small
  scratch image.
- Recent tasks capture concrete cache, thumbnail, nested-title, and UI contracts.

## Priority model

1. Reliability and security: auth lifecycle, trust boundaries, request limits,
   HTTP lifecycle, and deployment safety.
2. Maintainability: automated gates, handler/migration tests, truthful config,
   current specifications, and reproducible builds.
3. Performance and user experience: dependency modernization, browser coverage,
   accessibility, and focused profiling-driven improvements.

Effort is estimated as S (hours), M (several focused days), or L (multi-stage
work). Confidence is high only for reproduced or directly traced behavior.

## Follow-up boundaries

| Priority | Child task | Main outcome | Estimate |
|---|---|---|---|
| P0 | `07-17-auth-http-hardening` | Correct auth/token lifecycle and bounded HTTP behavior | M |
| P1 | `07-17-test-ci-baseline` | Enforced Go quality gate and critical boundary coverage | M |
| P1 | `07-17-config-deploy-docs-cleanup` | Truthful config and reproducible deployment docs | M |
| P1 | `07-17-frontend-deps-build` | Repair broken browser/API contracts, then modernize assets | L |

The children are independently verifiable. Security is recommended first; CI and
deployment cleanup can proceed independently. The frontend task should repair
the broken download-manager contract before dependency modernization; later
browser/auth work should target settled authentication behavior and integrate
with the CI task.

## Validation limitations

- A disposable startup verified initialization through schema migration 15 and
  background task launch, but the sandbox denied local socket binding.
- Docker Compose files were parsed, but Docker image construction was blocked by
  daemon permissions. Escalation could not be reviewed because the approval
  service returned HTTP 503.
- No browser engine or Playwright installation was available, so interactive and
  visual behavior was not claimed as verified.
- No installed `govulncheck`, `gosec`, `staticcheck`, `golangci-lint`, or frontend
  package manifest was available for a complete dependency vulnerability audit.

These gaps are explicit acceptance work in the relevant child tasks.

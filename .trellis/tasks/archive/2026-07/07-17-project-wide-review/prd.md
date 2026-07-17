# Project-wide review

## Goal

Produce an evidence-backed, project-wide technical review that identifies the
highest-value changes for maintainability, correctness, security, performance,
testability, deployment reliability, and user experience.

The review should help the maintainer decide what to improve next, rather than
collecting low-impact style observations.

## Confirmed Facts

- The repository is a single-repo project with backend and frontend Trellis
  specification layers.
- The current implementation includes a Go application, server-rendered
  templates, static JavaScript/CSS assets, and embedded web resources.
- The repository retains a top-level Crystal `spec/` suite alongside Go tests.
- The project supports several container deployment variants, including QNAP.
- The worktree already contains uncommitted Trellis backend specification files;
  this review must preserve and account for them.
- `go test ./...`, `go vet ./...`, `go build ./...`, and `go test -race
  ./...` pass when the Go build cache is placed in an allowed temporary path.
- Coverage is uneven: `internal/library` is about 80% covered while
  `internal/server` is about 11%; command startup and migrations have no direct
  statement coverage.
- No repository CI workflow or frontend automated test suite is present.
- The current download-manager page targets removed MangaDex HTTP/WebSocket
  routes and expects request/response contracts that do not match the Go queue
  API; the API documentation page references missing OpenAPI and ReDoc assets.
- The browser bundle includes legacy dependencies such as jQuery 3.2.1 and
  Alpine.js 2.8.0, and compiled CSS is maintained without a build pipeline.
- Confirmed authentication issues include an unrestricted post-login callback
  redirect and logout that clears the browser cookie without invalidating the
  persisted token.
- Several documented configuration fields are loaded but unused by runtime
  behavior, including `session_secret`, `log_level`, `download_timeout_seconds`,
  `cache_size_mbs`, and `cache_log_enabled`.
- The default Compose file fails validation when the empty values from
  `env.example` are used; QNAP Compose variants validate with obsolete-version
  warnings and contain stale configuration guidance.
- A disposable startup reached config generation, schema migration 15, initial
  admin creation, library scanning, and background task initialization. Local
  socket binding and Docker daemon access were blocked by the execution sandbox;
  escalation could not be reviewed because the approval service returned 503.
- No local browser engine or Playwright installation is available, so real
  browser smoke validation remains an explicit review limitation.

## Requirements

- Inspect architecture, package boundaries, data flow, persistence, security,
  error handling, concurrency, and resource lifecycle.
- Inspect frontend structure, accessibility, dependency age, client-side safety,
  responsiveness, and maintainability.
- Inspect automated tests, static checks, build tooling, container images,
  deployment configuration, documentation, and dependency health.
- Include local runtime and browser smoke validation using disposable data,
  including key pages/APIs, embedded assets, response headers, Docker image
  construction, and Compose configuration.
- Validate important findings against code, tests, configuration, history, or
  reproducible tooling output.
- Separate confirmed defects and risks from optional modernization ideas.
- Rank recommendations by impact, urgency, effort, and confidence.
- Prefer focused, incremental improvements over speculative rewrites.
- Prioritize reliability and security findings first, maintainability second,
  then performance and user-experience improvements when recommendations have
  otherwise comparable value.
- Do not modify product code as part of the review/planning phase.
- Deliver the recommended roadmap as independently executable follow-up tasks
  for authentication/HTTP hardening, tests/CI, frontend dependencies/build
  tooling, and configuration/deployment/documentation cleanup.

## Acceptance Criteria

- [x] The main runtime architecture and supported deployment paths are mapped.
- [x] Relevant existing specs, task history, and prior decisions are considered.
- [x] Appropriate existing test, lint, vet, build, and dependency checks are run
      or explicitly reported as unavailable.
- [x] Every high-priority finding includes concrete repository evidence.
- [x] Recommendations are grouped into an actionable priority order with likely
      effort and validation guidance.
- [x] Positive project strengths and areas that should remain unchanged are
      identified alongside improvement opportunities.
- [x] Uncertain findings and remaining scope limitations are clearly disclosed.

## Out of Scope

- Implementing the recommended changes during this planning task.
- A wholesale framework or language rewrite without evidence that incremental
  changes cannot address the underlying problem.
- Production-environment penetration testing or destructive data migration tests.
- Connecting to, changing, or testing against the user's NAS or production data.

## Open Questions

- None currently blocking planning.

## Notes

- This is an assessment and prioritization task. Its primary deliverable is a
  review report and an agreed follow-up roadmap, not source-code changes.
- Follow-up work will be split into four child tasks with explicit dependencies
  and independent acceptance criteria.

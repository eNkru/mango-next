# Project-wide review execution plan

## Deliverable

Create `review.md` in this task directory as the final assessment. Product code
must remain unchanged during this parent task.

## Checklist

- [x] Map the Go runtime, frontend assets, persistence, background work, and
      deployment paths.
- [x] Review Trellis specs, recent history, and available cross-session memory.
- [x] Run Go test, vet, build, race, and coverage baselines.
- [x] Inspect authentication, HTTP, upload, plugin, archive, storage, migration,
      configuration, frontend, and deployment boundaries.
- [x] Attempt disposable startup, Compose validation, Docker build, and browser
      tooling discovery; record environmental limitations without bypassing them.
- [x] Write findings in severity order with file/line evidence, impact, confidence,
      effort, and a concrete validation/fix direction.
- [x] Include strengths and explicit non-recommendations to prevent unnecessary
      rewrites.
- [x] Cross-check every P0/P1 finding against its full call path and existing
      tests to eliminate false positives.
- [x] Ensure each accepted recommendation maps to exactly one child task or is
      explicitly deferred.
- [x] Re-run repository status and the available Go quality gate before finalizing
      the report.
- [x] Review `prd.md`, `design.md`, `implement.md`, `review.md`, and all four child
      PRDs for consistency.

## Validation commands

```bash
cd go
GOCACHE=/tmp/mango-next-go-cache go test ./...
GOCACHE=/tmp/mango-next-go-cache go test -race ./...
GOCACHE=/tmp/mango-next-go-cache go vet ./...
GOCACHE=/tmp/mango-next-go-cache go build ./...
GOCACHE=/tmp/mango-next-go-cache go test -cover ./...
```

```bash
docker compose config
docker compose -f docker-compose.qnap.yml config
docker compose -f docker-compose.qnap-prebuilt.yml config
```

Docker build, live HTTP smoke, and browser checks must be rerun in an environment
where Docker daemon, local listening sockets, and a browser engine are available.

## Review gates

- Planning artifacts and child scopes require user review before `task.py start`.
- Starting this parent task authorizes only the assessment report, not child-task
  implementation.
- Each child task requires its own design/implementation review before code
  changes begin.

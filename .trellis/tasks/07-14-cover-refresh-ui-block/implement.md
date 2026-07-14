# Implement: Non-blocking and identity-safe cover refresh

## Ordered Checklist

1. Add read-only, path-aware title/entry identity checks to `internal/storage`.
2. Validate all cache identities before `LoadFromCache` publishes the tree.
   Remove and ignore a confirmed stale cache; preserve it on transient DB
   errors.
3. Refactor `ThumbnailContext` into a synchronized running/progress state with
   atomic start rejection and deferred finish.
4. Snapshot only top-level title pointers under a short library read lock;
   expand entries, generate images, and save thumbnails after releasing
   `Library.mu`.
5. Make `tasks.Runner` gate its first automatic thumbnail run on completion of
   the initial scan attempt instead of a fixed startup delay.
6. Add `running` to the thumbnail progress API and update the admin JavaScript
   to consume it directly.
7. Add focused regressions:
   - cache made from database A is rejected with database B;
   - valid cache still loads and reuses unchanged titles;
   - thumbnail generation does not prevent a queued library writer and new
     readers from completing promptly;
   - duplicate jobs are rejected and lifecycle state resets after failures;
   - initial automatic thumbnail generation starts only after initial scan;
   - progress JSON and admin state distinguish running at 0% from idle.
8. Run formatting and the complete Go quality gate.

## Validation Commands

```bash
cd go && gofmt -w <changed-go-files>
cd go && go test ./internal/library ./internal/storage ./internal/tasks ./internal/server
cd go && go test -race ./internal/library ./internal/tasks
cd go && go test ./... && go vet ./... && go build ./...
```

Run the Docker deployment and verify during an overlapping scan/thumbnail job:

```bash
docker compose up --build -d
docker compose logs --tail=80 backend
```

Manual acceptance checks:

- repeatedly load `/`, `/library`, and the progress endpoint during refresh;
- confirm requests return in seconds rather than waiting for the full job;
- confirm no thumbnail FK 787 errors are emitted;
- confirm progress remains running at 0%, advances, and returns idle;
- confirm a second manual start while active does not create another job.

## Risk And Rollback Points

- Cache rejection affects startup availability on the first boot after a DB
  replacement; verify a successful scan recreates and reloads the cache.
- Runner ordering is concurrency-sensitive; use deterministic channels in tests
  rather than sleeps where possible.
- Keep the progress response additive (`progress` plus `running`) to avoid API
  breakage.
- No migration or data deletion beyond a cache proven inconsistent with the
  configured database.

## Review Gates

- Confirm no archive/image/database work remains inside `Library.mu` scope.
- Confirm cache validation is path-aware and read-only.
- Confirm every job exit executes `Finish()`.
- Confirm focused tests fail against the current implementation and pass after
  the fix.
- Confirm full Go tests, race tests, vet, and build pass before completion.

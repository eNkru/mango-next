# Design: Non-blocking and identity-safe cover refresh

## Problem

Two independent defects combine during startup:

1. A valid gzip cache can belong to a replaced/new SQLite database. Cached IDs
   are copied directly into memory and unchanged titles are reused, so missing
   `ids` parent rows are never recreated. `thumbnails.id -> ids.id` then rejects
   every insert with SQLite extended error 787.
2. `GenerateThumbnails` holds `Library.mu.RLock` for the whole archive loop. If
   a scan finishes and queues for the write lock, Go's writer-preferring
   `RWMutex` blocks all later UI readers behind that writer until thumbnail
   generation releases its long read lock.

The progress API also encodes both "idle" and "running at 0%" as `progress=0`,
which makes the admin UI clear its running state too early.

## Data Flow And Ownership

```text
cache JSON -> library cache decoder -> storage identity validation -> in-memory tree
                                                            |
                                              mismatch -> discard cache -> full scan

initial scan complete -> automatic thumbnail job -> entry snapshot -> archive/image work
                                                         |             (no library lock)
                                                         v
                                               short SQLite thumbnail writes

ThumbnailContext -> {running, progress} API -> admin UI state
```

- `library/cache.go` owns cache structure and traversal.
- `storage` owns read-only checks that an ID maps to the expected library path.
- `Library` owns the immutable entry snapshot and synchronized job/progress
  lifecycle.
- `tasks.Runner` owns startup ordering between the initial scan and first
  automatic thumbnail job.
- The admin API owns serialization of `running` and `progress`; JavaScript only
  consumes that contract.

## Cache Identity Validation

Before `LoadFromCache` publishes any cached tree, validate every cached title
and entry ID against the expected path in `titles` and `ids` respectively.
Include nested title IDs present in the cache payload. Validation is read-only:
cache loading must not create database rows from potentially stale data.

If all identities match, publish the cache as today. If any identity is absent,
unavailable, or mapped to another path, remove the stale cache, leave the
in-memory tree empty, log one actionable message, and let the normal initial
scan rebuild both database IDs and cache. A transient database error returns an
error without deleting the cache.

This deliberately prefers an empty-but-responsive first boot over exposing a
tree whose tags, progress, and thumbnail foreign keys are invalid. Subsequent
boots regain instant cache loading after the successful scan writes a valid
cache.

## Locking And Job Lifecycle

`GenerateThumbnails` acquires `Library.mu.RLock` only long enough to copy the
current top-level `*Title` pointers into a local slice. It releases the lock
before expanding those titles into entries and before any database, archive,
image, sleep, or logging work. A scan may atomically swap the current tree while
the job continues using its retained title/entry objects.

The critical section is therefore O(top-level titles), not O(all books). For the
reported library shape (roughly tens of titles and thousands of entries), the
expected hold time is microseconds to low milliseconds. The concurrency test
uses a conservative 100 ms responsiveness bound to detect accidental disk/DB
work under this lock without relying on a fragile microbenchmark.

Move all thumbnail lifecycle fields behind `ThumbnailContext.mu`:

- `Start(total)` atomically rejects a duplicate job and sets `running=true`.
- `Increment()` updates completed work.
- `Status()` returns a consistent `(progress, running)` snapshot.
- `Finish()` runs via `defer`, clearing running/current/total on every return.

No long operation may hold `Library.mu`. Duplicate manual/scheduled starts are
idempotently skipped and logged.

## Startup Ordering

Replace the fixed "30 seconds after process start" assumption with an explicit
initial-scan completion signal inside `tasks.Runner`. The automatic thumbnail
loop waits for the first `runScan` attempt to finish before its first run. This
prevents it from racing a cache rejection/rebuild on large libraries. Manual
generation remains asynchronous and safe because cache publication and job
snapshotting enforce the same invariants.

Later periodic scans and thumbnail jobs may overlap: entry identity is already
valid, and thumbnail work no longer holds the library lock.

## API Compatibility

`GET /api/admin/thumbnail_progress` keeps the numeric `progress` field and adds
the boolean `running` field. Existing clients remain compatible. The bundled
admin UI uses `running` rather than inferring activity from `progress > 0`.

## Failure Handling

- Per-entry read/generate/save failures remain isolated and logged.
- `defer Finish()` guarantees a terminal job state even when a future code path
  returns early.
- A stale cache produces one cache-level log instead of one FK error per book.
- Initial scan failure still releases the runner signal; the thumbnail job sees
  either a previously validated cache or an empty tree.

## Trade-offs

- Path-aware cache validation adds indexed SQLite lookups at startup. This is
  cheaper and safer than opening every archive, and occurs before HTTP startup.
- Rejecting the complete cache on one identity mismatch sacrifices partial
  cache availability for a simple all-or-nothing invariant.
- SQLite remains single-connection. Individual thumbnail reads/writes can
  briefly serialize DB access, but image/archive work cannot monopolize it.

## Rollback

Changes are isolated to library cache/job code, runner ordering, progress API,
admin JavaScript, and focused tests. Reverting those files restores prior
behavior; no schema migration or destructive database change is required.

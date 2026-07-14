# Eliminate UI blocking during cover refresh

## Goal

Keep the web UI responsive while cover thumbnails are refreshed, including
when cached library entries are stale or no longer have matching database IDs.

## Requirements

- Thumbnail generation must remain a background operation and must not block
  unrelated page or API requests for the duration of a library-wide refresh.
- A cached entry without a valid database parent row must not cause repeated
  SQLite foreign-key failures for every archive in the library.
- Thumbnail progress must remain observable and must reach a terminal state
  even when an individual archive cannot be read or persisted.
- Existing valid thumbnails and normal scheduled/manual refresh behavior must
  remain compatible.
- The first automatic thumbnail refresh must not start against unvalidated
  cached identities before the initial library scan has completed.
- Failures for individual books must be isolated and logged without stopping
  the whole refresh or leaving the generator permanently marked as running.
- The concurrency change must introduce no data races, dangling references,
  cache corruption, or partial database identity writes.

## Acceptance Criteria

- [ ] Browsing and ordinary API requests remain responsive while a full
      thumbnail refresh is running.
- [ ] Thumbnail persistence does not emit `FOREIGN KEY constraint failed (787)`
      for entries loaded from a stale/incomplete cache.
- [ ] Invalid cached identity data is repaired, rejected, or safely reconciled
      before thumbnail rows are written.
- [ ] Concurrent thumbnail-generation requests cannot create duplicate jobs or
      race the progress state.
- [ ] Progress resets after success, per-entry failures, or unexpected job
      termination, and a later refresh can be started.
- [ ] The progress API distinguishes an active job at 0% from an idle job, and
      the admin UI keeps showing the active state until the job finishes.
- [ ] Automated tests cover stale cached IDs, foreground access during refresh,
      and progress/job lifecycle behavior.
- [ ] `go test -race` passes for the affected concurrent packages, and cache/DB
      round-trip tests prove valid data is preserved while stale data is only
      rejected after a read-only identity check.

## Confirmed Facts

- The container loads books from cache before starting cover refresh.
- During refresh, saving thumbnails fails repeatedly with SQLite extended error
  code 787 (`FOREIGN KEY constraint failed`) across unrelated ZIP/RAR archives.
- The UI appears stuck during the same refresh window.
- The current implementation logs the error per entry and continues scanning.
- A recent change deletes a library cache that cannot be parsed, but a
  successfully parsed cache may still contain stale database identities.
- `GenerateThumbnails` holds `Library.mu.RLock()` across the complete library
  iteration, including archive reads, image processing, database writes, and a
  100 ms per-entry delay.
- A concurrent scan builds outside the library lock but eventually waits for
  `Library.mu.Lock()` to publish its result. Go's writer-preferring `RWMutex`
  then prevents new UI readers from acquiring `RLock`, so requests queue until
  the thumbnail job releases its long-held read lock.
- Cache deserialization restores title and entry IDs directly from JSON. The
  incremental scanner reuses an unchanged cached title without calling
  `GetOrCreateTitleID` or `GetOrCreateEntryID`, so a valid cache paired with a
  new/replaced database can retain IDs absent from `titles`/`ids` indefinitely.
- `thumbnails.id` has a foreign key to `ids.id`; saving a thumbnail for one of
  those stale cached entry IDs therefore produces SQLite error 787.
- Storage uses one SQLite connection. Thumbnail image decoding happens outside
  the database, so this can serialize short reads/writes but does not explain
  the multi-minute library-lock freeze.
- The previously accepted non-blocking scan requirement is that the UI responds
  in seconds while background library work continues.

## Out of Scope

- Replacing SQLite or redesigning the full library model without evidence that
  either is required.
- Changing archive/image rendering behavior unrelated to persistence or request
  responsiveness.
- Auditing plugin updates, downloads, and other background jobs that are not
  part of the cover-refresh/scan interaction.

## Open Questions

- None. The user approved limiting this task to cover refresh, including its
  overlap and startup ordering with library scans.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.

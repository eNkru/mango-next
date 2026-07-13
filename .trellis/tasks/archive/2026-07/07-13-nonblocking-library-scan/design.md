# Design: Non-blocking library scan

## Problem

```
Scan() {
  mu.Lock()           // held for entire disk walk
  ScanLibrary(...)    // minutes on NAS
  swap TitleHash
  mu.Unlock()
}
```

Readers (`RLock`) starve until walk completes.

## Solution

```
Scan() {
  result := ScanLibrary(...)   // no lib.mu
  mu.Lock()
  TitleIDs / TitleHash = from result
  mu.Unlock()
  return result
}
```

### Concurrency notes

- Mid-scan readers see **previous** tree or empty (first boot). Acceptable.
- Two concurrent `Scan()` calls: both may build trees; last swap wins. Optional: `sync.Mutex`/`atomic` scan-in-progress to skip overlap — nice-to-have if admin scan + interval collide.
- `GenerateThumbnails` keeps `RLock` for duration; still allows other readers. Do **not** upgrade to write lock.

### Optional hardening (same task if cheap)

- `scanning atomic.Bool` or `TryLock` style: if scan already running, skip/log and return.
- Admin `POST /api/admin/scan` already calls `Scan()`; benefits automatically.

### Startup

No change required to `main.go` / `tasks.Start` if default interval ≥ 1: scan already async. Fix is lock scope only.

If `scan_interval_minutes < 1`, `runScan()` runs synchronously inside runner goroutine only — still not blocking `Listen`. Leave as-is.

### Tests

- Unit: `Scan` updates titles; concurrent reader during mock slow scan (inject delay via interface or test-only hook) observes non-blocking — if too heavy, test that `ScanLibrary` is called outside lock by structure review + regression test that Scan succeeds.
- Prefer: refactor not required for inject; document + existing TestLibraryScan + new test that holds RLock in another goroutine while Scan runs on small lib still completes quickly.

### Rollback

Revert single commit on `fix/nonblocking-library-scan`.

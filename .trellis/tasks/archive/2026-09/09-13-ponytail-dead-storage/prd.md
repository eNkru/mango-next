# Delete dead storage methods + library_cache.go

## Goal

Remove zero-caller storage code: four dead methods plus one whole dead file
(~82 lines). All verified with repo-wide grep (definitions + doc comments
are the only matches).

## Confirmed facts (verified 2026-09-12)

- `BulkMarkTitleRead` — go/internal/storage/progress.go:122-131 (10 lines).
  Zero callers, not even tests (`apiBulkProgress` does not call it).
- `BulkMarkTitleUnread` — go/internal/storage/progress.go:133-139
  (7 lines). Zero callers.
- `MigrateProgressTable` — go/internal/storage/progress.go:232-238
  (7 lines). Not called from the migration runner
  (go/internal/storage/migration/). One-off leftover.
- `CountEntries` — go/internal/storage/title.go:161 (~7 lines). The single
  grep hit besides the definition is its own doc comment.
- `library_cache.go` (51 lines) — contains ONLY `SaveLibraryCache`
  (:15) and `LoadLibraryCache` (:37), both with zero callers (only their
  doc comments). Superseded by the library-package JSON cache:
  main.go:70 calls `lib.LoadFromCache` (library package), a different
  function. Delete the whole file.

## Requirements

1. Delete the four methods listed above.
2. Delete go/internal/storage/library_cache.go entirely.
3. No test updates needed (nothing references the deleted code).

## Acceptance criteria

- [ ] `cd go && go build ./...` passes.
- [ ] `cd go && go test ./...` passes (notably storage + library suites).
- [ ] `grep -rn "BulkMarkTitleRead\|BulkMarkTitleUnread\|MigrateProgressTable\|CountEntries\|SaveLibraryCache\|LoadLibraryCache" go --include="*.go"` returns nothing.
- [ ] Library scan + thumbnail generation still run (cache path untouched).

## Out of scope — DO NOT DELETE (verified live / test-covered)

- `GetAllTitles` / `GetAllEntries` / `TitleRecord` / `EntryRecord` —
  called from library_test.go. Keep.
- `GetEntriesSortTitle`, `CountTitles`, `DeleteThumbnail` — called from
  storage_test.go. Keep.
- `ContinueReadingItem.Percentage` and sibling progress fields —
  `apiContinueReading` reads `item.Percentage` (handlers_api.go:518).
  Deleting breaks compilation. Keep.
- `lib.LoadFromCache` / library-package cache — live. Keep.

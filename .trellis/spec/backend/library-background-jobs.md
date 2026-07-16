# Library Background Jobs

## Scenario: Cache-safe, non-blocking thumbnail generation

### 1. Scope / Trigger

Apply this contract when changing library cache loading, scanning, thumbnail
generation, task startup ordering, or the admin thumbnail-progress endpoint.
It prevents stale cache IDs from violating SQLite foreign keys and prevents a
background job from starving HTTP readers through `sync.RWMutex` writer
preference.

### 2. Signatures

- `Library.LoadFromCache() error`
- `Library.GenerateThumbnails() error`
- `Library.ThumbnailStatus() (progress float64, running bool)`
- `Storage.TitleIdentityMatches(id, absPath string) (bool, error)`
- `Storage.EntryIdentityMatches(id, absPath string) (bool, error)`
- `GET /api/admin/thumbnail_progress`

### 3. Contracts

- Cache identities are checked read-only before the cached tree is published.
- A confirmed mismatch removes the cache and leaves an empty in-memory tree;
  the initial scan rebuilds database IDs and a valid cache.
- A transient database validation error is returned and must not delete cache.
- `GenerateThumbnails` may hold `Library.mu.RLock` only while copying top-level
  title pointers. Archive, image, database, sleep, and entry-expansion work runs
  after unlock.
- The first scheduled thumbnail run starts after the initial scan attempt.
- Progress JSON is additive and stable:

```json
{"success": true, "progress": 0.0, "running": true}
```

`running`, not `progress > 0`, is the source of truth for job activity.

### 4. Validation & Error Matrix

| Condition | Required behavior |
|-----------|-------------------|
| Cache ID/path matches an available DB row | Load cache |
| Cache ID is missing, unavailable, or mapped to another path | Delete cache, return no error, run full scan |
| Identity query fails | Return wrapped error, preserve cache |
| Duplicate thumbnail start | Log and return without starting a second job |
| One archive read/generate/save fails | Log, advance progress, continue |
| Job returns after `Start` | Deferred `Finish` clears running/current/total |

### 5. Good/Base/Bad Cases

- Good: warm cache and matching DB load immediately; the incremental scan may
  reuse unchanged titles.
- Base: no cache produces an empty responsive UI until the scan publishes data.
- Bad: a cache from another DB is never exposed and never seeds DB rows.
- Bad: no disk or database operation occurs while holding `Library.mu`.

### 6. Tests Required

- Build a cache using DB A, load with empty DB B, and assert the cache is
  removed while DB B row counts remain zero.
- Block an entry's `ReadPage`, then assert a library writer and a later reader
  both acquire the lock promptly.
- Start a duplicate job and assert it returns promptly; release the first job
  and assert status becomes `(0, false)`.
- Assert the progress endpoint reports `running=true` at exactly 0%.
- Assert the initial automatic thumbnail creates data only after scan-created
  IDs exist.

### 7. Wrong vs Correct

Wrong:

```go
lib.mu.RLock()
defer lib.mu.RUnlock()
for _, entry := range entries {
    readArchiveAndWriteThumbnail(entry)
}
```

Correct:

```go
lib.mu.RLock()
titles := snapshotTitles(lib.TitleHash)
lib.mu.RUnlock()

for _, title := range titles {
    generateWithoutLibraryLock(title)
}
```

## Scenario: Nested titles in TitleHash

### 1. Scope / Trigger

Apply when changing library scan, `NewTitle`, `applyTitles`, library cache
serialization, `DeepEntries`, or title/book pages that resolve `TitleIDs`.

### 2. Contracts

- `NewTitle` keeps nested titles on `Children` and mirrors IDs in `TitleIDs`
  (sorted by child **name**).
- `Library.TitleIDs` is **top-level only** (library shelf).
- `Library.TitleHash` contains **every** title depth so `TitleHash[subID]` works.
- Library cache (v2) stores a **flat** list of all titles; load rebuilds
  `Children` from `title_ids`.
- `DeepEntries` recurses `Children`. Thumbnail generation iterates **top-level**
  titles only, then `DeepEntries` (do not range all `TitleHash` or nested
  entries are counted twice).
- Cover URL helper `firstEntryID` must return a **entry** id from
  `DeepEntries()`, never a nested title id from `TitleIDs` (breaks `/api/cover`).
  Any first deep entry is acceptable for series-level cover.

### 3. Wrong vs Correct

Wrong: parent only stores child IDs; `applyTitles` hashes top-level only → book
page skips missing nested titles (empty JOJO-style multi-folder trees).

Correct: retain `Children`, flatten into `TitleHash`, cache all depths.

## Scenario: Thumbnail decode formats (JPEG/PNG/GIF/WebP)

### 1. Scope / Trigger

Apply when changing `internal/thumbnail` decode/generate paths, library
image-extension support, or tests that exercise thumbnail generation.

### 2. Contracts

- Production code in `thumbnail` must blank-import decoders for every format
  it claims to generate from: `image/jpeg` (also used for encode), `image/png`,
  `image/gif`, and `golang.org/x/image/webp` (registers with `image.Decode*`).
- Prefer a single `image.Decode` / `image.DecodeConfig` path after registration.
  Do not fallback-decode arbitrary bytes with the WebP decoder: non-WebP input
  yields `riff: missing RIFF chunk header` and masks the real failure.
- Output remains JPEG (size policy 200w portrait / 300h landscape unchanged).
- Thumbnail tests must not import `image/png` or `image/gif` solely to build
  fixtures if that would register decoders process-wide and hide missing
  production imports; use raw fixture bytes or `//go:embed` instead.

### 3. Wrong vs Correct

Wrong: only encode-import `image/jpeg`, then on `image.Decode` failure call
`webp.Decode` on every remaining buffer.

Correct: register PNG/GIF/WebP; decode only via `image.Decode*`; return the
real decode error for unsupported/corrupt data.

# Design: first book open / dimensions performance

## Problem

`GET /api/dimensions/{tid}/{eid}` is on the reader critical path. Today it:

1. Calls `ReadPage` for **every** page (full image bytes).
2. For archives, each `ReadPage` **opens the archive, lists, sorts, reads, closes**.
3. Stores nothing → every open pays the same cost.

Client waits for the full dimensions array before leaving the loading state (out of scope to change).

## Goals

- Cache hit: O(1) DB read, no page I/O.
- Cache miss: compute once with **one** archive open (or one dir pass), then persist.
- Invalidate when entry content signature changes.
- Keep API JSON shape and reader JS unchanged.

## Non-goals

- Progressive first paint / reader.js timing changes.
- Background precompute on scan/thumbnail.
- Full page-image LRU (`cache_enabled`).
- Partial/stream header-only reads in v1 (optional follow-up).

## Architecture

```
reader.js  →  apiDimensions
                 │
                 ├─ findEntry(tid,eid)
                 ├─ storage.GetEntryDimensions(entryID, signature)
                 │     hit  → return cached JSON
                 │     miss → entry.ReadPageDimensions()  [new batch API]
                 │            storage.SaveEntryDimensions(...)
                 └─ { success, dimensions: [{width,height}, ...] }
```

### Persistence

New table (migration **v15**), parallel to `thumbnails` (FK → `ids`):

```sql
CREATE TABLE entry_dimensions (
  id TEXT NOT NULL PRIMARY KEY,          -- entry id
  signature TEXT NOT NULL,             -- content sig at compute time
  dimensions TEXT NOT NULL,            -- JSON array of {width,height}
  page_count INTEGER NOT NULL,
  updated_at INTEGER NOT NULL,
  FOREIGN KEY (id) REFERENCES ids (id) ON UPDATE CASCADE ON DELETE CASCADE
);
```

- **Key**: entry `id` (unique row).
- **Validity**: stored `signature` must equal current `entry.Signature()` (decimal string, same as `ids.signature`).
- On mismatch or missing row → recompute and UPSERT.
- Optional: if `page_count` ≠ `entry.PageCount()` treat as invalid (defense in depth).

Why DB not file: co-located with entry identity, FK cleanup, matches thumbnails pattern, easy tests.

### Compute path (miss)

Add library-level batch method (name flexible):

```go
// Returns width/height per page in reading order (1..PageCount).
func ReadPageDimensions(e Entry) ([]Dim, error)
```

| Entry type | Strategy |
|------------|----------|
| `ArchiveEntry` | Single `archive.Open` → list/filter/sort image entries once → for each page `ReadEntry` + `thumbnail.DecodeConfig` → Close |
| `DirEntry` | Iterate `files` once: read each file + DecodeConfig (same as today but one clear path; no N× open archive) |

`apiDimensions` must call this batch path, **not** a loop of `ReadPage`.

Keep `ReadPage` as-is for `api/page` (per-page open remains; out of scope for image LRU unless cheap shared index later).

### API contract (unchanged)

```json
{ "success": true, "dimensions": [ {"width": w, "height": h}, ... ] }
```

- Page order: same as today (1-based reading order; array index 0 = page 1).
- Decode failure → `{0,0}` (same as today).
- Read failure mid-loop: prefer continue with `{0,0}` for that page (match current) rather than failing entire request, unless entry is unreadable entirely.

### Concurrency

- Two concurrent first-opens of same entry may both miss and both write; last writer wins with same signature → OK.
- No requirement for singleflight in v1; optional if tests show stampedes.

### Compatibility

- New installs: migration 15 runs with others.
- Existing DBs: `user_version` 14 → 15 adds empty table; first open per entry populates.
- Crystal schema was through 14; Go-only table is fine if product is mango-next Go (document in design). If dual-stack DB required later, Crystal ignores unknown tables.

### Failure modes

| Case | Behavior |
|------|----------|
| Cache row, wrong signature | Recompute + UPSERT |
| Cache JSON corrupt | Treat as miss |
| Entry not found | 404 as today |
| Unreadable archive | 500 / error path as today |
| Partial page decode fails | that page 0,0 |

### Trade-offs

| Choice | Benefit | Cost |
|--------|---------|------|
| Persist in SQLite | Simple invalidation, FK | Migration; DB growth small (JSON dims only) |
| Lazy only | Small scope | First open still does full page reads once |
| Full page bytes on miss | Simple, correct DecodeConfig | First open still I/O heavy; header-only later |
| No reader.js change | Low risk | Loading gate still full dims |

### Follow-ups (not this task)

- Header-only / limited read for common codecs.
- Reuse open archive across `api/page` sequential reads.
- Wire `cache_enabled` page LRU.
- Optional precompute during thumbnail job.

## Test plan (design-level)

1. **Cache miss then hit**: first dimensions call populates; second call no archive open / no per-page full re-read (instrument fake Entry or count opens).
2. **Signature change**: update sig or replace content → miss → new dims stored.
3. **Delete entry / FK**: deleting ids row removes dimensions row.
4. **DirEntry + ArchiveEntry** both covered (zip/cbz at minimum; rar/7z if test fixtures allow).
5. **API shape**: JSON field names and 0,0 semantics.

## Rollback

- Revert migration consumer code; table can remain empty/unused.
- Feature is additive; disable path = always recompute if needed via flag is **not** required for v1.

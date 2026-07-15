# Research: first book open path

Date: 2026-07-15

## User-facing flow

1. User navigates to `/reader/{title}/{entry}/{page}` (or `/reader/{title}/{entry}` → redirect to page 1)
2. Server `handleReader` renders `views/reader` with title/entry metadata only (no page dimensions)
3. Client `reader.js` Alpine component `init()`:
   - GET `api/dimensions/{tid}/{eid}` — **blocking** until complete; UI shows loading
   - Builds `items[]` with page URLs + width/height
   - Sets mode from localStorage, `updateMode`, preload next N pages (`preloadLookahead`, default 3)
4. Actual page images load via `api/page/{tid}/{eid}/{page}` when rendered / preloaded

## Critical path (first open)

### Server: `apiDimensions` (`go/internal/server/handlers_api.go:319-347`)

```go
for i := 1; i <= entry.PageCount(); i++ {
  img, err := entry.ReadPage(i)
  w, h, err := thumbnail.DecodeConfig(img.Data)
  dims = append(dims, ...)
}
```

- Reads **every page fully** into memory, then only uses DecodeConfig (header) for width/height
- **No persistence / no cache** of dimensions in storage or memory

### Archive `ReadPage` (`go/internal/library/title.go:181-224`)

Per page call:
1. `archive.Open(path)`
2. `arc.Entries()` list + filter images + sort
3. `arc.ReadEntry` full bytes
4. `arc.Close()`

So for an N-page CBZ/RAR/7z: **N archive open + N full listing/sort + N full page reads** just for dimensions.

### DirEntry

`ReadPage` = full file `os.ReadFile` per page — still N full image reads for dimensions.

### Client after dimensions

- Preloads `preloadLookahead` (default 3) images via `new Image().src`
- Current page image also loads for display
- `loading` only cleared after dimensions return — first paint blocked on full archive scan

## First-open vs subsequent

| Layer | First open | Subsequent |
|-------|------------|------------|
| Dimensions API | Full N-page read, no cache | Same cost every time (no dim cache) |
| Page images | Network + decode cold | Browser HTTP cache may help for same pages |
| Library tree | Library cache exists | Not related to per-entry open |
| Config `cache_enabled` / `cache_size_mbs` | Present in config | **No Go usage found** for page image LRU yet (may be Crystal remnant) |

## Existing related caches

- Library tree cache: `library/cache.go` (titles/entries metadata, not page dims)
- Thumbnails: `storage.GetThumbnail` for covers only
- Client localStorage: mode, margin, preloadLookahead, fitType — not dimensions
- Reader preload: only lookahead pages, not dimensions

## Top optimization candidates (ranked by impact / effort)

1. **Persist page dimensions** (DB or file keyed by entry ID + signature)
   - Fill on first open (or background); later opens O(1) JSON
   - Invalidate when entry signature/mtime changes
   - Files: `handlers_api.go` apiDimensions, `storage`, maybe scan job

2. **Stop full-page reads for dimensions**
   - DecodeConfig only needs image headers; can stream partial reads from zip/rar when possible
   - Or store dims while generating thumbnails / scanning
   - ArchiveEntry still re-opens archive per page — even with partial read, open+list N times hurts

3. **Reuse open archive / page index for ArchiveEntry**
   - Cache sorted image entry list per path; single open for multi-page batch (dimensions + sequential reads)
   - Files: `title.go` ArchiveEntry.ReadPage, archive package

4. **Progressive dimensions / non-blocking first paint**
   - Return page count immediately; stream dims or default 100% sizes; show first page ASAP
   - Client: don't block `loading=false` on all dims (longPages heuristic can wait)
   - Files: `reader.js`, `apiDimensions`

5. **Background precompute dimensions** on library scan / thumbnail pass
   - First open becomes cache hit if user browsed library first
   - Couples to library-background-jobs contracts

## Product questions (code cannot answer)

1. Scope of "first open": first ever for an entry, cold server, or every reader session?
2. Target: cold first paint latency goal? (e.g. <1s to first page)
3. Acceptable tradeoff: slightly wrong layout until dims arrive vs wait for accurate layout?
4. Archive types priority (cbz vs rar vs dir)?
5. Must match Crystal mango behavior exactly for dimensions?

## Key files

- `go/web/public/js/reader.js` — client init, dimensions gate
- `go/web/views/reader.tmpl` — reader shell
- `go/internal/server/handlers_api.go` — apiDimensions, apiPage
- `go/internal/server/handlers_pages.go` — handleReader
- `go/internal/library/title.go` — ArchiveEntry/DirEntry ReadPage
- `go/internal/thumbnail/thumbnail.go` — DecodeConfig
- `go/internal/config/config.go` — cache_enabled (unused for pages?)
- `.trellis/spec/backend/library-background-jobs.md` — library cache contracts

## Complexity assessment

Complex: multi-layer (API + archive + optional storage + client), correctness (signature invalidation), measurable acceptance.

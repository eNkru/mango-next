# Design: library cache + incremental scan

## Why second scan is still slow

Every `Scan()` → `ScanLibrary` → every top-level dir `NewTitle` → recursive dir walk + `ContentsSignature` + per-archive `GetOrCreateEntryID` + open archives for page counts. **No “already scanned → skip”**. DB only stores id/path/signature, not the full tree needed for the reader.

## Architecture

```
Startup
  LoadLibraryCache (gzip JSON)
  if ok → short Lock swap TitleHash   // UI has books
  go Scan()                           // already in tasks.Runner

Scan (incremental)
  list top-level dirs
  for each dir:
    sig = DirSignature(dir)  // cheap: mtime/size walk of names — still disk but lighter than full NewTitle if we only compare to cache
    if cache.Title[path] exists && cache.sig == sig:
      reuse Title pointer/tree from previous memory or deserialized cache
    else:
      NewTitle(dir)  // expensive
  MarkUnavailable for missing ids (existing)
  short Lock swap
  SaveLibraryCache(serialize tree)
```

### Signature choice for skip

- Prefer **`DirSignature`** (uint64, already on Title) for “unchanged directory” — matches storage title signature intent.
- If DirSignature misses pure content rename edge cases, optional second check `ContentsSig` (already computed in NewTitle).

### Cache payload (JSON, gzip)

Minimal fields to rebuild in-memory tree + open files on demand:

```json
{
  "version": 1,
  "library_path": "/root/mango/library",
  "titles": [ { "dir", "id", "parent_id", "name", "signature", "contents_sig", "mtime", "entries": [...], "titles": [nested...] } ]
}
```

Entries: path, id, signature, name, page count if stored, type (archive/dir).  
On load: reconstruct `Title`/`Entry` objects that can `ReadPage` from path (same as live scan).

### Config path

- Use `cfg.LibraryCachePath` when wiring Library/Scanner; fix `Storage.SaveLibraryCache` to accept path or use config (today hardcodes next to library).

### Failure modes

| Case | Behavior |
|------|----------|
| No cache file | empty tree until scan |
| Corrupt gzip/JSON | log, ignore, full scan |
| library_path mismatch | ignore cache, full scan |
| Partial write | write temp + rename |

### Concurrency

- `scanMu` already serializes Scan.
- LoadCache at startup: before first Scan starts, or under scanMu so Scan doesn’t race.

### Performance expectation (user NAS)

| Phase | Today | Target |
|-------|--------|--------|
| Cold start, no cache | ~12m | same (first time) |
| Restart, cache warm | empty until 12m | **seconds** to show books; bg scan minutes or less |
| Interval scan, no changes | ~12m | **seconds–1m** (top-level sig only) |
| One title changed | ~12m | rescan that title only |

## Risks

- Stale cache shows deleted books until scan finishes — acceptable; scan corrects.
- Wrong signature → miss updates — prefer slightly over-invalidating if unsure.

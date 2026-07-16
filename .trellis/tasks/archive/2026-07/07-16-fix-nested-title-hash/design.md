# Design: nested titles in TitleHash + library cache

## Problem

`NewTitle` builds nested titles and only keeps child IDs. `applyTitles` / cache only retain top-level `*Title`, so UI/`TitleHash` lookups for subtitles always miss.

## Approach (recommended)

Keep parent→child IDs as today (`TitleIDs` / `ParentID`). **Retain child objects during scan**, then **flatten every title into `TitleHash`**.

### 1. In-memory model

Add on `Title`:

```go
// Children holds nested titles produced by NewTitle (not serialized as pointers).
// TitleIDs remains the stable ID list used by handlers.
Children []*Title `json:"-"`
```

`NewTitle` when accepting a sub-title:

- append `sub.ID` to `TitleIDs` (unchanged)
- append `sub` to `Children`

### 2. Library apply

```go
func (lib *Library) applyTitles(top []*Title) {
  ids := only top-level IDs
  hash := map[string]*Title{}
  var walk func(*Title)
  walk = func(t *Title) {
    if t == nil || t.ID == "" { return }
    hash[t.ID] = t
    for _, c := range t.Children {
      walk(c)
    }
  }
  for _, t := range top { walk(t) }
  lib.TitleIDs, lib.TitleHash = ids, hash
}
```

- **TitleIDs**: top-level only (library shelf).
- **TitleHash**: all depths.

### 3. Cache format

Use a **flat** `titles[]` array (same `cachedTitle` shape), containing **every** title (top + nested). Links via `parent_id` + `title_ids`.

- `titlesToCache`: DFS/BFS from top-level, emit each title once (`titleToCached` already writes `TitleIDs` + entries).
- `titlesFromCache`:
  1. `titleFromCached` each row into map by ID
  2. Rebuild `Children` from each title’s `TitleIDs`
  3. Return top-level list = titles with empty `ParentID` (or order preserved from scan: those whose dir is direct child of library root)

Order of top-level: preserve existing sort (numeric name). Nested order: existing `TitleIDs` order.

Bump `libraryCacheVersion` to **2** if needed for clarity; v1 files with only top-level still load but nested still broken until rescan — acceptable: after deploy, next scan rewrites cache.

### 4. DeepEntries

```go
func (t *Title) DeepEntries() []Entry {
  out := append([]Entry{}, t.Entries...)
  for _, c := range t.Children {
    out = append(out, c.DeepEntries()...)
  }
  return out
}
```

Handlers that only have `TitleHash` + `TitleIDs` (no Children after partial construction) need a Library-aware deep walk **or** ensure Children is always rebuilt on cache load and scan apply. **Preferred:** always rebuild Children on apply/cache load so `DeepEntries` works on any title pointer in the hash.

When only `TitleIDs` is set (legacy tests constructing bare structs), DeepEntries falls back to direct entries only — same as today for flat titles.

Better: `DeepEntries` using a callback is awkward. Alternative library method:

```go
func (lib *Library) deepEntries(t *Title) []Entry
```

Prefer fixing `Children` so `Title.DeepEntries` is self-contained after apply/load.

### 5. Incremental scan reuse

When top-level signature matches, reuse entire old `*Title` subtree (already has Children if previous apply populated it). Ensure reused titles from **pre-fix** memory still work after first full rebuild post-upgrade.

`snapshotByDir` currently maps **all** TitleHash dirs. Reuse key is top-level dir only in `ScanLibrary` — OK. Nested titles under a reused top title remain via the reused pointer graph.

### 6. Sorting note

`TitleIDs` sort currently uses `compareNumerically` on **IDs** not names — pre-existing. Out of scope unless trivial to fix via sorting `Children` by name then regenerating `TitleIDs`.

## Compatibility

| Layer | Change |
|-------|--------|
| DB IDs | none |
| API JSON | same shapes; more non-empty children |
| Cache | flat list may grow; v1 still readable |
| Root zip | still unsupported |

## Trade-offs

| Option | Pros | Cons |
|--------|------|------|
| **A. Children + flat cache (chosen)** | Minimal API change; fits TitleHash | Must keep Children/TitleIDs in sync |
| B. Nested JSON children only | Natural tree | larger cache format change |
| C. Flatten filesystem on scan | No nest UI | changes product; rejects user layout |

## Rollback

Revert commit; delete library cache file to force rescan with old code.

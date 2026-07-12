# Design: go-signature

## Decision: recommended algorithm (locked)

| Layer | Crystal | Go (this task) | OK? |
|-------|---------|----------------|-----|
| File uint64 | inode | FNV64(path + mtime + size) | Yes — portable; already in prod Go |
| Dir uint64 | sort(inodes) → CRC32 | FNV(walk rel + mtime) | Yes — same role: change detect |
| DirEntry | SHA1(inode strings) | FNV(mtime+size of images) | Yes — change detect for loose dirs |
| Contents | SHA1(joined basenames) | **same** SHA1 + cache | Exact intent parity |

**Why FNV works with DB**

`GetOrCreateTitleID` / `GetOrCreateEntryID`:

1. path + signature exact  
2. **path only** → keep id, `UPDATE signature`  
3. insert  

So algorithm drift or content edit only refreshes signature string; **UUID id stable by path**.

**Why not inode**

- `syscall.Stat_t.Ino` not portable to all targets; Docker volume / copy loses inodes  
- Would churn signatures vs current Go DB rows unnecessarily  
- ID already path-keyed  

## Architecture

```
go/internal/library/signature.go   # public API + ContentsSignature
go/internal/library/title.go       # call sites use helpers (keep or thin wrappers)
```

Prefer **same package `library`** (not new module): signatures are only used by scan/title/entry; avoids import cycles with storage.

### Public API

```go
// FileSignature — uint64 for supported archive/image; 0 if unsupported.
// Implementation: existing fileSignature FNV (not inode).
func FileSignature(path string) (uint64, error)

// DirSignature — uint64 directory signature (FNV walk).
func DirSignature(dirname string) uint64

// ContentsSignature — SHA1 hex of sorted supported basenames (recursive).
// mirrors Crystal Dir.contents_signature
func ContentsSignature(dirname string, cache map[string]string) (string, error)

// DirectoryEntrySignature — keep current dirEntrySignature semantics (FNV);
// document Crystal used SHA1(inodes); Go uses FNV for consistency with FileSignature.
func DirectoryEntrySignature(files []string) uint64  // or (dirname) if re-scanned
```

Internal unexported helpers may remain `fileSignature` / `dirSignature` as thin aliases.

### ContentsSignature algorithm (Crystal parity)

```
ContentsSignature(dir, cache):
  if cache[dir] hit → return
  names := []
  for each entry in dir (skip ".*"):
    if isDir → append ContentsSignature(child, cache)
    else if supported archive|image → append basename (fn only)
  hash = SHA1-hex(join names with "")  // Crystal: signatures.join (no separator)
  cache[dir] = hash
  return hash
```

Crystal sorts **dir.entries** before walk (`dir.entries.sort`); Go: `ReadDir` then sort by name for determinism.

Skip Fiber.yield (Crystal concurrency cooperative); not needed in Go.

### Optional wire (this task if cheap)

- `Title` field `ContentsSignature string` set in `NewTitle`  
- **Not** full examine recursion this task unless time permits  
- Full examine → leave note for library/routes child  

Minimum ship: **API + unit tests**; NewTitle storing contents sig is nice-to-have for future examine.

## Compatibility

- No SQLite migration  
- No change to GetOrCreate* contract  
- Existing tests for scan/title keep passing  

## Trade-offs

| Choice | Benefit | Cost |
|--------|---------|------|
| Keep FNV | Zero ID migration risk | Not byte-identical to Crystal inode |
| SHA1 contents | Rescan-ready, Crystal intent | Extra tree walk if called every scan |
| package library | Simple | title.go grows |

## Rollback

- Delete ContentsSignature + tests; revert wrappers — no DB rollback  

## Tests

1. ContentsSignature stable for fixed tree  
2. Add supported file → hash changes  
3. Add `.hidden` or unsupported ext → hash unchanged  
4. Nested dir: child rename supported file → parent hash changes  
5. Cache: second call same pointer map returns same without recompute (optional)  
6. FileSignature unsupported → 0  

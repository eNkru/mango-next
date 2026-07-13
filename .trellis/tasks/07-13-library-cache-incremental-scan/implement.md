# Implement: library cache + incremental scan

## Ordered steps

1. **Cache DTO + serialize/deserialize** (`go/internal/library/cache.go`)
   - JSON schema v1; map to/from `*Title` tree.
2. **Storage cache path**
   - `SaveLibraryCache`/`LoadLibraryCache` take explicit path (from config) or Library holds path.
3. **Library.LoadFromCache / SaveCache**
   - Load → swap under short lock; validate library_path.
4. **Incremental ScanLibrary**
   - Accept previous tree map path→*Title; skip NewTitle when DirSignature matches.
5. **Wire startup** (`cmd/mango` or `tasks` / `NewLibrary`)
   - After config: `lib.LoadFromCache(cfg.LibraryCachePath)`; then existing runner Scan.
6. **After successful Scan**: `SaveCache`.
7. **Tests**: round-trip; skip unchanged; invalidate on change; corrupt cache.

## Validate

```bash
cd go && go test ./... && go vet ./...
```

## Branch

`feat/library-cache-incremental-scan` (created).

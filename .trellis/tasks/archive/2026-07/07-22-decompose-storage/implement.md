# Implement: Decompose Storage Package

## Checklist

1. Create domain files by **moving** method blocks from `storage.go` (cut/paste, same package).
2. Move types/helpers with their domain (see design map).
3. Leave `storage.go` with only lifecycle + `Storage` type; verify line count < ~300.
4. Fix per-file imports (`goimports` / compile).
5. Run tests:
   ```bash
   cd go && go test ./internal/storage/... && go test ./...
   ```
6. Grep that no domain methods remain in `storage.go` except lifecycle.

## Validation

```bash
cd go && go test ./internal/storage/... && go test ./... && go vet ./internal/storage/...
wc -l go/internal/storage/storage.go   # expect < 300
```

## Risk / rollback

- Pure move; rollback = git revert.
- Do not edit SQL strings or public signatures.

## Files created/touched

- `go/internal/storage/storage.go` (shrink)
- `go/internal/storage/user.go` (new)
- `go/internal/storage/thumbnail.go` (new)
- `go/internal/storage/tag.go` (new)
- `go/internal/storage/title.go` (new)
- `go/internal/storage/identity.go` (new)
- `go/internal/storage/missing.go` (new)
- `go/internal/storage/library_cache.go` (new)

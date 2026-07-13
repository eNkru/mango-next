# Implement: Non-blocking library scan

## Checklist

1. Change `Library.Scan` in `go/internal/library/library.go`:
   - Call `ScanLibrary` **before** `mu.Lock`.
   - Under lock: rebuild `TitleIDs` / `TitleHash` only.
2. Optional: if scan already in progress, log and return early (`atomic` or mutex flag).
3. Tests in `library_test.go`: concurrent RLock + Scan on fixture dir does not deadlock; scan result still correct.
4. `cd go && go test ./... && go vet ./...`

## Validate

```bash
cd go && go test ./... && go vet ./...
```

## Files

- `go/internal/library/library.go` (primary)
- `go/internal/library/library_test.go`
- Possibly `tasks` only if skip-overlap lives there (prefer library-level)

## Done when

PRD AC met; branch ready for review/commit.

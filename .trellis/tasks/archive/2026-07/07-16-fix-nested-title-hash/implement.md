# Implement: nested titles in TitleHash

## Ordered checklist

### 1. Title model + NewTitle

- [x] Add `Children []*Title` on `Title`
- [x] In `NewTitle`, when sub is accepted: append to `Children` and `TitleIDs`
- [x] Sort `Children` by name and rebuild `TitleIDs` from sorted children

### 2. DeepEntries

- [x] Implement recursive `DeepEntries` over `Children`
- [x] Thumbnail generation walks top-level only (avoid double-count)

### 3. Library.applyTitles

- [x] Walk top-level trees; insert every title into `TitleHash`
- [x] Keep `TitleIDs` as top-level only

### 4. Cache

- [x] `titlesToCache`: emit all titles (DFS), not only top slice
- [x] `titlesFromCache`: build map, wire `Children` from `TitleIDs`, return top-level list
- [x] Bump `libraryCacheVersion` to 2 (still load v0/v1)

### 5. Tests

- [x] `TestNestedArchiveTitlesInHash` for `Series/Part/vol.cbz` + cache round-trip
- [x] Existing tests still pass

### 6. Validation

```bash
cd go && go test ./internal/library/... ./internal/server/... ./internal/tasks/... -count=1
cd go && go build ./... && go vet ./...
```
# 73 tests passed in library/server/tasks

## Risky files

| File | Risk |
|------|------|
| `title.go` | NewTitle children lifecycle |
| `library.go` | applyTitles hash completeness |
| `cache.go` | load order / orphans |
| `handlers_*` | should work if TitleHash complete |

## Done when

PRD AC evidenced by tests; nested JOJO-style path works after rescan.

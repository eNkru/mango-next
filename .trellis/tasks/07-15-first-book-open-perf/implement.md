# Implement: first book open / dimensions performance

## Ordered checklist

### 1. Schema + storage API

- [x] Add migration **v15** `CreateEntryDimensions` in `go/internal/storage/migration/migrations.go`
- [x] Bump `LatestVersion()` to 15
- [x] `Storage` methods: Get/SaveEntryDimensions in `dimensions.go`
- [x] Unit tests: save/load, signature mismatch, corrupt JSON, upsert
- [x] Migration test includes `entry_dimensions`

### 2. Batch dimension read on Entry

- [x] `ReadPageDimensions` + shared `sortedImageEntries`
- [x] ArchiveEntry: single open batch path
- [x] DirEntry: one-pass files
- [x] Tests: archive order/sizes + dir entry + storage round-trip

### 3. Wire `apiDimensions`

- [x] Cache hit → return; miss → batch → save → return
- [x] Old per-page ReadPage loop removed

### 4. Regression / quality

- [x] `go test ./...` — 206 passed
- [ ] Manual smoke optional
- [x] `reader.js` untouched

### 5. Spec capture (after check)

- [ ] trellis-update-spec for dimensions cache contract (optional follow-up)

## Validation commands

```bash
cd go && go test ./internal/storage/... ./internal/library/... ./internal/server/... ./internal/thumbnail/...
```

(Adjust packages if tests live elsewhere.)

## Risky files / rollback points

| Area | Risk | Rollback |
|------|------|----------|
| `migrations.go` v15 | version mismatch with old binaries | keep forward-only; old binary ignores table |
| `apiDimensions` | wrong page order | compare against old loop on fixture |
| Archive batch | sort order differs from `ReadPage` | share same filter/sort helper as `ReadPage` |

## Suggested shared helper

Extract image-entry listing/sort from `ArchiveEntry.ReadPage` so batch path and single-page path cannot diverge on order.

## Done when

- PRD acceptance criteria checkboxes can be evidenced by tests
- User approves → `task.py start` then implement

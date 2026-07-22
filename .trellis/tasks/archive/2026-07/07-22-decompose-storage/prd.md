# Task PRD: Decompose Storage Package Repositories

## Goal

Refactor the oversized `go/internal/storage/storage.go` (~1180 lines, ~54 methods) into domain-focused source files **within the same package**, without breaking exported APIs or call sites outside `storage`.

## Target Area

- `go/internal/storage/` (split `storage.go`; leave package name and `*Storage` method set stable)
- Tests: `go/internal/storage/*_test.go` (no API rewrite expected)

## Confirmed facts (from repo)

- Already split: `progress.go` (~238 lines, progress + home lists), `dimensions.go` (~60 lines), `migration/`.
- Remaining in `storage.go`: ~54 methods — lifecycle/migrate, users/auth (~12), thumbnails (3), tags (5), title/entry meta/sort/hidden (~10), identity helpers (~8), missing/list/delete (~9), library cache (2), plus types (`User`, `Image`, `TitleRecord`, `EntryRecord`, `MissingItem`).
- Exported surface is methods on `*Storage` + types + `Open` — callers use package `storage`, not per-repo types.
- Review parent: `07-22-project-review` deep structural item #1.

## Requirements

1. Move `*Storage` methods out of `storage.go` into domain files; keep method signatures and package identity unchanged (no new interfaces required for MVP).
2. Leave `storage.go` for connection lifecycle: `Open`, pragmas, `migrate`/`Version`, `DB`/`Close`, and `Storage` type.
3. Do not change SQL schema or migration versions unless a bug is found.
4. Do not rename exported methods solely for aesthetics.
5. **Mechanical move only** — same code, new file; no logic rewrites, no opportunistic helper extraction.

## Acceptance Criteria

- [x] `storage.go` is under ~300 lines (lifecycle + `Storage` type only). (~124 lines)
- [x] Domain methods live in separate files per the decided map below.
- [x] Exported API of package `storage` is unchanged for external callers (same types/method names).
- [x] `cd go && go test ./internal/storage/...` and `go test ./...` pass.

## Out of scope

- Repository interfaces / DI rewrite for `server.Dependencies`.
- Splitting into multiple Go packages (`storage/user`, …).
- Changing SQLite schema or WAL/busy settings.
- Decomposing `storage_test.go` unless required for compile.
- Progress/dimensions re-split (already separate files).
- Logic cleanup, error-string rewrites, or shared SQL helper extraction.

## Decisions

- **File map (option A):**
  - `storage.go` — Open, pragmas, migrate/Version, DB/Close; `Storage` type only
  - `user.go` — users/auth + InitAdmin + password/username validation helpers
  - `thumbnail.go` — cover/thumb blob CRUD + `Image`
  - `tag.go` — title tags
  - `title.go` — hidden/sort/count title+entry metadata
  - `identity.go` — path helpers, GetOrCreate*ID, identity match/exists + TitleRecord/EntryRecord
  - `missing.go` — mark unavailable, list/delete missing + MissingItem
  - `library_cache.go` — Save/LoadLibraryCache
  - Existing: `progress.go`, `dimensions.go`, `migration/`
- **Types:** colocate with domain files (see above). `Storage` stays in `storage.go`.
- **Helpers:** `InitAdmin`, `hashPassword`, `verifyPassword`, `randomStr`, `validateUsername`, `validatePassword` → `user.go`.
- **Depth:** mechanical-only move.

## Open questions

- None blocking planning.

## Notes

- Complex task: needs `design.md` + `implement.md` before `task.py start`.
- Parent: `07-22-project-review`.

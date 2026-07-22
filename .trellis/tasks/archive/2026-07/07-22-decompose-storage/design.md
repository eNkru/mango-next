# Design: Decompose Storage Package

## Architecture / boundaries

Stay in package `storage`. No new interfaces, no multi-package split.

| File | Owns |
|------|------|
| `storage.go` | `Storage` struct, `Open`, pragmas, `migrate`, `Version`, `DB`, `Close` |
| `user.go` | `User`, auth/user CRUD, `InitAdmin`, bcrypt/uuid/validate helpers |
| `thumbnail.go` | `Image`, Save/Get/DeleteThumbnail |
| `tag.go` | title tag CRUD/list |
| `title.go` | hidden/sort metadata + CountTitles/CountEntries |
| `identity.go` | `TitleRecord`, `EntryRecord`, path helpers, GetOrCreate*ID, identity match/exists |
| `missing.go` | `MissingItem`, MarkUnavailable, GetAll*, List/Delete missing |
| `library_cache.go` | Save/LoadLibraryCache |
| `progress.go` / `dimensions.go` | unchanged |
| `migration/` | unchanged |

## Compatibility

- Exported method set and type names unchanged.
- Call sites in `server` / `library` / `tasks` need no edits.
- SQL and migration versions unchanged.

## Trade-offs

| Choice | Why |
|--------|-----|
| Same package, file split | Lowest risk; Go methods stay on `*Storage` |
| Colocated types | Domain cohesion without a types grab-bag |
| Mechanical only | Reviewable pure move; cleanup is a follow-up |

## Rollback

Git revert the move commit(s). No schema or config change.

## Risks

- Missed unexported helper left in wrong file → compile fix only.
- Import set per file must be trimmed after split (gofmt / compile will catch).

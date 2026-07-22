# Task PRD: Decompose Storage Package Repositories

## Goal

Refactor `go/internal/storage/storage.go` (1100+ lines) by decomposing it into modular, domain-specific repository files (`user_repo.go`, `progress_repo.go`, `book_repo.go`, `tag_repo.go`) within the `storage` package.

## Target Area

- `go/internal/storage/`

## Requirements

1. Move struct methods on `Storage` into logically separated source files without breaking exported package APIs or type definitions.
2. Group user operations in `user_repo.go`, reading progress in `progress_repo.go`, titles/books/covers in `book_repo.go`, and tags in `tag_repo.go`.
3. Keep `storage.go` concise, focusing on database connection lifecycle, driver setup, and schema migrations.

## Acceptance Criteria

- [ ] `storage.go` size is reduced to <300 lines.
- [ ] Methods on `Storage` are cleanly separated by domain into `*_repo.go` files.
- [ ] All unit tests (`go test ./internal/storage/...`) pass without API breaking changes.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.

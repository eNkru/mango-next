# Task PRD: SQLite WAL Mode and Busy Timeout Optimization

## Goal

Improve SQLite database concurrency and prevent "database locked" errors during simultaneous background scans, image extraction, and reading progress updates by configuring WAL mode and busy timeout pragmas.

## Target Area

- `go/internal/storage/storage.go`

## Requirements

1. Enable SQLite Write-Ahead Logging (WAL) mode via `PRAGMA journal_mode=WAL;` upon database connection initialization.
2. Set busy timeout to 5000ms via `PRAGMA busy_timeout=5000;` to allow waiting readers/writers to retry gracefully rather than failing immediately.
3. Ensure backwards compatibility and pass existing unit tests in `storage_test.go`.

## Acceptance Criteria

- [ ] SQLite connection setup in `storage.go` executes `journal_mode=WAL` and `busy_timeout=5000`.
- [ ] Database storage unit tests (`go test ./internal/storage/...`) pass cleanly.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.

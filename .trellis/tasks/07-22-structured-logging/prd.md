# Task PRD: Structured Logging Migration with slog

## Goal

Migrate standard Go `log.Printf` statements in `internal/server` and `internal/storage` to Go 1.21's standard `log/slog` library for structured, level-aware JSON/Text logging.

## Target Area

- `go/internal/server/*.go`
- `go/internal/config/loglevel.go`

## Requirements

1. Replace standard `log.Printf` / `log.Println` calls with contextual `slog.Info`, `slog.Error`, `slog.Debug`, or `slog.Warn`.
2. Configure configurable log level (debug, info, warn, error) controlled via environment/config.
3. Preserve key contextual metadata in log fields (e.g., HTTP request path, user ID, task ID).

## Acceptance Criteria

- [ ] HTTP handlers and storage operations log via `slog`.
- [ ] Log levels filter outputs correctly according to server configuration.
- [ ] Backend test suite passes without breaking existing log capture if any.

## Notes

- Keep `prd.md` focused on requirements, constraints, and acceptance criteria.
- Lightweight tasks can remain PRD-only.
- For complex tasks, add `design.md` for technical design and `implement.md` for execution planning before `task.py start`.

# Task PRD: Structured Logging Migration with slog

## Goal

Migrate standard Go `log.Printf` in `internal/server` to Go `log/slog`, and replace the half-working `ApplyLogLevel` with a real level-filtered slog default logger so operators can control verbosity via existing `log_level` / `LOG_LEVEL`.

## Target Area

- `go/internal/server/*.go` (all `log.Print*` call sites, including HTTP access log middleware)
- `go/internal/config/loglevel.go` (+ related config tests)
- `go/cmd/mango/main.go` only if needed to wire `ApplyLogLevel` / default logger once at boot

## Confirmed facts (from repo)

- Go module is `go 1.26.3` — `log/slog` is available without third-party deps.
- Config already has `LogLevel` (`yaml: log_level`, env `LOG_LEVEL`, default `info`).
- Current `ApplyLogLevel` only tweaks stdlib `log` flags/prefix; it does **not** filter by severity.
- Tests do not assert log message content; `TestApplyLogLevelDoesNotPanic` only checks no panic.
- ~dozens of `log.Print*` live outside server (`library`, `tasks`, `storage`, `upload`, `web`); **out of scope for this task**.

## Requirements

1. Configure a process-wide `slog` default logger in `ApplyLogLevel` with real level filtering: `debug`, `info`, `warn`/`warning`, `error` (unknown → `info`).
2. Replace `log.Printf` / `log.Println` in `internal/server` with `slog.Info` / `slog.Warn` / `slog.Error` / `slog.Debug` as appropriate.
3. HTTP access log (`LoggingMiddleware`) logs at **Info** with structured fields: status, method, path, duration.
4. Error paths in handlers log with level `Error` and an `err` attribute when an error value exists.
5. Default handler is **Text** on stderr (`slog.NewTextHandler`); at `debug` level enable source location (`AddSource`).
6. Bridge stdlib `log` output into slog so unmigrated packages still respect `log_level` (map residual `log.Print*` to Info unless clearly fatal).
7. Do not add external logging libraries.

## Acceptance Criteria

- [x] All `log.Print*` call sites under `go/internal/server/` are gone (prefer zero stdlib `log` imports in server).
- [x] `ApplyLogLevel` configures `slog` so messages below the configured level are not emitted.
- [x] Access middleware emits structured fields: `status`, `method`, `path`, `duration` at Info.
- [x] Handler/server errors use `slog.Error` with `err` attr when an error value exists.
- [x] `go test ./internal/config/... ./internal/server/...` pass.
- [x] Packages outside `server`/`config` may still use `log.Print*` (explicit non-goal).

## Out of scope

- Migrating `library`, `tasks`, `storage`, `upload`, `web` to slog (follow-up).
- Request-scoped logger / username / request-id on every line.
- Status-based access-log levels (4xx Warn / 5xx Error).
- JSON handler or `log_format` config.
- Changing default `log_level` (stays `info`).
- Log aggregation stack.

## Decisions

- **Scope:** server + config only (not library/tasks/storage/upload/web).
- **Format:** Text handler on stderr; JSON out of scope.
- **Residual stdlib log:** bridge into slog (Print* → Info) so `log_level` still filters unmigrated packages.
- **Request context:** out of scope this task.
- **Access log level:** always Info.

## Open questions

- None blocking planning.

## Notes

- Parent: `07-22-project-review` quick win #3.
- Complex: needs `design.md` + `implement.md` before `task.py start`.

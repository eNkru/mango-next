# Implement: Structured Logging Migration with slog

## Checklist

1. **config/loglevel.go**
   - Rewrite `ApplyLogLevel` to set slog default Text handler + level.
   - Bridge stdlib `log` output into slog Info.
   - Keep `LogLevelWriter` usable for tests (document behavior).
2. **config tests**
   - Extend beyond panic-only: assert level filtering if practical (capture handler or buffer).
3. **server middleware**
   - `LoggingMiddleware` → slog Info with status/method/path/duration.
4. **server call sites**
   - Replace every `log.Print*` in `internal/server` (`auth.go`, `handlers_api.go`, `handlers_pages.go`, `middleware.go`, `server.go`).
   - Map errors → `slog.Error(..., "err", err)`; warnings → `slog.Warn`; startup → `slog.Info`.
5. **Verify**
   - `cd go && go test ./internal/config/... ./internal/server/...`
   - Grep: no `log.Print` under `go/internal/server/`.
6. **Spec**
   - After implement+check: update `.trellis/spec/backend/` logging convention if present or add a short note.

## Validation commands

```bash
cd go && go test ./internal/config/... ./internal/server/...
rg 'log\.(Print|Fatal|Panic)' go/internal/server || true
```

## Risk / rollback points

- **Risk:** bridge double-formats timestamps — mitigate with `log.SetFlags(0)`.
- **Risk:** high-volume access logs at Info when `log_level=info` — same as today.
- **Rollback:** git revert; no DB migration.

## Files likely touched

- `go/internal/config/loglevel.go`
- `go/internal/config/config_test.go`
- `go/internal/server/middleware.go`
- `go/internal/server/auth.go`
- `go/internal/server/handlers_api.go`
- `go/internal/server/handlers_pages.go`
- `go/internal/server/server.go`

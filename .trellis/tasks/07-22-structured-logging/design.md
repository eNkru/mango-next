# Design: Structured Logging Migration with slog

## Architecture / boundaries

| Layer | Responsibility |
|-------|----------------|
| `config.ApplyLogLevel` | Own process-wide logging setup: parse level, install `slog` default logger (Text + stderr), bridge stdlib `log` into slog. |
| `internal/server` | Call `slog.*` at call sites; no package-local logger type. |
| Out-of-scope packages | Keep `log.Print*`; filtered only via stdlib→slog bridge. |

No new packages, no third-party log libs.

## Contracts

### Levels

| Config value | `slog.Level` | Notes |
|--------------|--------------|--------|
| `debug` | `LevelDebug` | `HandlerOptions.AddSource = true` |
| `info` / `""` / unknown | `LevelInfo` | default |
| `warn` / `warning` | `LevelWarn` | |
| `error` | `LevelError` | |

### Handler

```go
h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
    Level:     level,
    AddSource: level <= slog.LevelDebug,
})
slog.SetDefault(slog.New(h))
```

### Stdlib bridge

Goal: residual `log.Printf` from library/tasks/storage/upload still respect configured minimum level.

Approach:

1. After setting slog default, point `log.SetOutput` at a writer that parses/forwards lines into `slog.Default().Info(...)` (or a small `io.Writer` adapter).
2. Clear or minimize stdlib flags/prefix so double-timestamp noise is avoided when possible (`log.SetFlags(0)`).
3. Bridge always maps residual Print* → **Info** (cannot recover true severity from free-form strings).

Fatal paths outside server (`web` embed `log.Fatalf`) stay as-is for this task.

### Server call-site mapping

| Situation | API |
|-----------|-----|
| Access log | `slog.Info("request", "status", code, "method", m, "path", p, "duration", d)` |
| Operational info (server start, auth proxy warning) | `slog.Info` / `slog.Warn` |
| Handler/storage failures with `err` | `slog.Error("…", "err", err)` (+ extra attrs if already present as path/id in message) |
| Debug-only detail | `slog.Debug` only if clearly noise today; default most former Printf → Info/Error |

Prefer structured attrs over long interpolated messages; short message string + key attrs.

### Compatibility

- Env/config keys unchanged: `log_level` / `LOG_LEVEL`.
- `LogLevelWriter()` may keep returning the active writer for tests, or adapt to return stderr / bridge writer; tests only require no panic today — strengthen with a level-filter unit test if cheap.
- `main` continues to call `config.ApplyLogLevel(cfg.LogLevel)` once at boot.

## Trade-offs

| Choice | Why |
|--------|-----|
| Server-only migration | Quick win; full-repo later |
| Text not JSON | Human-operated self-host default |
| Bridge stdlib → Info | Approximate filtering without rewriting every package |
| No request-scoped logger | Avoids large middleware/context refactor |

## Rollback

Revert commits; no schema/API surface change. Operators keep same `LOG_LEVEL` env.

## Security

Do not log passwords, session tokens, or full `Authorization` headers. Access log path only (no query string unless already logged today — prefer path only).

# Structured Logging (slog)

## Scope / Trigger

Apply when adding or changing log calls under `go/internal/server/`, or when editing
`go/internal/config/loglevel.go` / `LOG_LEVEL` behavior.

## Signatures

```go
// go/internal/config/loglevel.go
func ApplyLogLevel(level string) // sets slog default + bridges stdlib log
func LogLevelWriter() io.Writer  // stdlib log writer (for tests)

// Levels (case-insensitive): debug | info | warn|warning | error | unknown→info
// Handler: slog.NewTextHandler(os.Stderr, ...); debug enables AddSource.
// Bridge: residual log.Print* → slog.Info (filtered by configured level).
```

## Contracts

| Site | API |
|------|-----|
| Access log | `slog.Info("request", "status", code, "method", m, "path", p, "duration", d)` |
| Handler failures | `slog.Error("…", "err", err)` (+ extra attrs as needed) |
| Startup / ops | `slog.Info` / `slog.Warn` |
| Unmigrated packages | may still use `log.Print*`; filtered via bridge at Info |

Config keys unchanged: `log_level` / `LOG_LEVEL` (default `info`).

## Validation & Error Matrix

| Condition | Behavior |
|-----------|----------|
| `LOG_LEVEL=error` | Info/Warn slog + bridged Print* suppressed; Error emitted |
| `LOG_LEVEL=debug` | All levels + source location |
| Unknown level string | Treated as info |

## Wrong vs Correct

#### Wrong
```go
log.Printf("Save progress error: %v", err)
```

#### Correct
```go
slog.Error("Save progress error", "err", err)
```

## Out of scope (current)

- JSON handler / `log_format`
- Request-id / username on every line
- Migrating library/tasks/storage/upload/web off `log.Print*`

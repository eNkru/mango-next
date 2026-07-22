package config

import (
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
)

// ApplyLogLevel configures the process-wide slog default logger and bridges
// the standard library logger into slog so unmigrated packages still respect
// log_level filtering.
//
// Levels: debug, info, warn/warning, error (case-insensitive). Unknown values
// default to info. Debug enables source location on the text handler.
func ApplyLogLevel(level string) {
	lv := parseLogLevel(level)
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     lv,
		AddSource: lv <= slog.LevelDebug,
	})
	slog.SetDefault(slog.New(h))

	// Residual log.Print* from packages not yet on slog.
	log.SetOutput(slogBridge{})
	log.SetFlags(0)
	log.SetPrefix("")
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info", "":
		return slog.LevelInfo
	default:
		return slog.LevelInfo
	}
}

// slogBridge forwards stdlib log output into slog at Info level.
type slogBridge struct{}

func (slogBridge) Write(p []byte) (int, error) {
	msg := strings.TrimSpace(string(p))
	if msg != "" {
		slog.Info(msg)
	}
	return len(p), nil
}

// LogLevelWriter returns the writer currently used by the standard logger.
// Exposed for tests.
func LogLevelWriter() io.Writer {
	return log.Writer()
}

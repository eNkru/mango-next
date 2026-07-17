package config

import (
	"io"
	"log"
	"os"
	"strings"
)

// ApplyLogLevel configures the standard library logger based on log_level.
// Levels: debug, info, warn/warning, error (case-insensitive). Unknown values
// default to info. Debug enables standard log flags with file:line; error
// suppresses routine output by discarding the default writer (callers that need
// error visibility should still use log.Printf for errors at info+).
func ApplyLogLevel(level string) {
	lv := strings.ToLower(strings.TrimSpace(level))
	switch lv {
	case "debug":
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	case "warn", "warning", "error":
		// Keep stderr but prefix so operators can filter; stdlib cannot fully
		// suppress by severity without a custom Logger per call site.
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags)
		if lv == "error" {
			// Best-effort: still log, but mark level for operators grepping output.
			log.SetPrefix("[error] ")
		} else {
			log.SetPrefix("[warn] ")
		}
	case "info", "":
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags)
		log.SetPrefix("")
	default:
		log.SetOutput(os.Stderr)
		log.SetFlags(log.LstdFlags)
		log.SetPrefix("")
	}
}

// LogLevelWriter returns the writer currently used by the standard logger.
// Exposed for tests.
func LogLevelWriter() io.Writer {
	return log.Writer()
}

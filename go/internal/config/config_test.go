package config

import (
	"bytes"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaults(t *testing.T) {
	c := defaults()
	if c.Port != 9000 {
		t.Errorf("default port = %d, want 9000", c.Port)
	}
	if c.BaseURL != "/" {
		t.Errorf("default base_url = %q, want /", c.BaseURL)
	}
	if !c.CacheEnabled {
		t.Error("cache_enabled default should be true")
	}
}

func TestPrecedenceFileOverEnv(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	os.WriteFile(cfgPath, []byte("port: 1234\nlog_level: debug\n"), 0o644)

	// Env sets a different port; file must win.
	t.Setenv("PORT", "5678")
	t.Setenv("HOST", "1.2.3.4") // not in file → env wins over default

	c, err := Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if c.Port != 1234 {
		t.Errorf("file>env failed: port = %d, want 1234", c.Port)
	}
	if c.Host != "1.2.3.4" {
		t.Errorf("env>default failed: host = %q, want 1.2.3.4", c.Host)
	}
	if c.LogLevel != "debug" {
		t.Errorf("log_level = %q, want debug", c.LogLevel)
	}
}

func TestEnvBool(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	os.WriteFile(cfgPath, []byte("default_username: bob\n"), 0o644)
	t.Setenv("DISABLE_LOGIN", "1")

	c, err := Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if !c.DisableLogin {
		t.Error("DISABLE_LOGIN=1 should set DisableLogin true")
	}
}

func TestBaseURLNormalized(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	os.WriteFile(cfgPath, []byte("base_url: /mango\n"), 0o644)

	c, err := Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if c.BaseURL != "/mango/" {
		t.Errorf("base_url = %q, want trailing slash /mango/", c.BaseURL)
	}
}

func TestPreprocessRejectsBadBaseURL(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yml")
	os.WriteFile(cfgPath, []byte("base_url: mango\n"), 0o644)

	if _, err := Load(cfgPath); err == nil {
		t.Error("expected error for base_url not starting with /")
	}
}

func TestApplyLogLevelDoesNotPanic(t *testing.T) {
	for _, lv := range []string{"debug", "info", "warn", "error", "bogus"} {
		ApplyLogLevel(lv)
	}
	ApplyLogLevel("info")
}

func TestApplyLogLevelFiltersByLevel(t *testing.T) {
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelError})
	slog.SetDefault(slog.New(h))
	// Mirror bridge used by ApplyLogLevel for residual stdlib log.
	log.SetOutput(slogBridge{})
	log.SetFlags(0)

	slog.Info("should-not-appear")
	slog.Error("should-appear")
	log.Print("bridge-info-should-not-appear")

	out := buf.String()
	if strings.Contains(out, "should-not-appear") {
		t.Fatalf("info leaked at error level: %q", out)
	}
	if !strings.Contains(out, "should-appear") {
		t.Fatalf("error missing at error level: %q", out)
	}
	if strings.Contains(out, "bridge-info-should-not-appear") {
		t.Fatalf("bridged info leaked at error level: %q", out)
	}

	// Restore process defaults for other tests in this package.
	ApplyLogLevel("info")
}

func TestParseLogLevel(t *testing.T) {
	cases := map[string]slog.Level{
		"debug":   slog.LevelDebug,
		"INFO":    slog.LevelInfo,
		"":        slog.LevelInfo,
		"bogus":   slog.LevelInfo,
		"warn":    slog.LevelWarn,
		"warning": slog.LevelWarn,
		"error":   slog.LevelError,
	}
	for in, want := range cases {
		if got := parseLogLevel(in); got != want {
			t.Errorf("parseLogLevel(%q)=%v want %v", in, got, want)
		}
	}
}

func TestDumpsDefaultWhenMissing(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "sub", "config.yml")

	c, err := Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if c.Port != 9000 {
		t.Errorf("dumped default port = %d, want 9000", c.Port)
	}
	if _, err := os.Stat(cfgPath); err != nil {
		t.Errorf("config file was not dumped: %v", err)
	}
}

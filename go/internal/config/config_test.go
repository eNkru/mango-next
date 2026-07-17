package config

import (
	"os"
	"path/filepath"
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

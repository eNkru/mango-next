package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	current *Config
	mu      sync.RWMutex
)

// Current returns the process-wide singleton config (Config.current in Crystal).
func Current() *Config {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// SetCurrent installs c as the singleton.
func (c *Config) SetCurrent() {
	mu.Lock()
	defer mu.Unlock()
	current = c
}

// Path returns the config file path this config was loaded from.
func (c *Config) Path() string { return c.path }

// Load reads the config from path, applying env overrides and defaults.
// If path is empty, it uses CONFIG_PATH or ~/.config/mango/config.yml.
// When the file does not exist, it dumps the default config there (matching config.cr).
func Load(path string) (*Config, error) {
	if path == "" {
		if p, ok := os.LookupEnv("CONFIG_PATH"); ok && p != "" {
			path = p
		} else {
			path = "~/.config/mango/config.yml"
		}
	}

	cfgPath, err := expandPath(path)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(cfgPath); err == nil {
		// File exists: file > env > default.
		c := defaults()
		c.applyEnv()
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, c); err != nil {
			return nil, err
		}
		c.path = path
		if err := c.expandPaths(); err != nil {
			return nil, err
		}
		if err := c.preprocess(); err != nil {
			return nil, err
		}
		return c, nil
	}

	// File does not exist: dump defaults (env-applied) and return them.
	c := defaults()
	c.applyEnv()
	c.path = path
	if err := c.expandPaths(); err != nil {
		return nil, err
	}
	fmt.Printf("The config file %s does not exist. Dumping the default config there.\n", cfgPath)
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		return nil, err
	}
	out, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(cfgPath, out, 0o644); err != nil {
		return nil, err
	}
	fmt.Printf("The config file has been created at %s.\n", cfgPath)
	return c, nil
}

// expandPaths expands ~ in the six path options (config.cr expand_paths).
func (c *Config) expandPaths() error {
	for _, p := range []*string{
		&c.LibraryPath, &c.LibraryCachePath, &c.DBPath,
		&c.QueueDBPath, &c.UploadPath, &c.PluginPath,
	} {
		expanded, err := expandPath(*p)
		if err != nil {
			return err
		}
		*p = expanded
	}
	return nil
}

// preprocess validates base_url and login settings (config.cr preprocess).
func (c *Config) preprocess() error {
	if !strings.HasPrefix(c.BaseURL, "/") {
		return fmt.Errorf("base url (%s) should start with `/`", c.BaseURL)
	}
	if !strings.HasSuffix(c.BaseURL, "/") {
		c.BaseURL += "/"
	}
	if c.DisableLogin && c.DefaultUsername == "" {
		return fmt.Errorf("login is disabled, but default username is not set. " +
			"Please set a default username")
	}
	return nil
}

// expandPath expands a leading ~ to the user's home directory.
func expandPath(p string) (string, error) {
	if p == "~" || strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if p == "~" {
			return home, nil
		}
		return filepath.Join(home, p[2:]), nil
	}
	return p, nil
}

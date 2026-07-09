package config

import (
	"os"
	"strconv"
	"strings"
)

// Config mirrors the Crystal Config class (src/config.cr).
// Precedence: config file > environment variable > default value.
type Config struct {
	Host                             string `yaml:"host"`
	Port                             int    `yaml:"port"`
	BaseURL                          string `yaml:"base_url"`
	SessionSecret                    string `yaml:"session_secret"`
	LibraryPath                      string `yaml:"library_path"`
	LibraryCachePath                 string `yaml:"library_cache_path"`
	DBPath                           string `yaml:"db_path"`
	QueueDBPath                      string `yaml:"queue_db_path"`
	ScanIntervalMinutes              int    `yaml:"scan_interval_minutes"`
	ThumbnailGenerationIntervalHours int    `yaml:"thumbnail_generation_interval_hours"`
	LogLevel                         string `yaml:"log_level"`
	UploadPath                       string `yaml:"upload_path"`
	PluginPath                       string `yaml:"plugin_path"`
	DownloadTimeoutSeconds           int    `yaml:"download_timeout_seconds"`
	CacheEnabled                     bool   `yaml:"cache_enabled"`
	CacheSizeMBs                     int    `yaml:"cache_size_mbs"`
	CacheLogEnabled                  bool   `yaml:"cache_log_enabled"`
	DisableLogin                     bool   `yaml:"disable_login"`
	DefaultUsername                  string `yaml:"default_username"`
	AuthProxyHeaderName              string `yaml:"auth_proxy_header_name"`
	PluginUpdateIntervalHours        int    `yaml:"plugin_update_interval_hours"`

	path string
}

// defaults matches the OPTIONS constant in config.cr.
func defaults() *Config {
	return &Config{
		Host:                             "0.0.0.0",
		Port:                             9000,
		BaseURL:                          "/",
		SessionSecret:                    "",
		LibraryPath:                      "~/mango/library",
		LibraryCachePath:                 "~/mango/library.yml.gz",
		DBPath:                           "~/mango.db",
		QueueDBPath:                      "~/mango/queue.db",
		ScanIntervalMinutes:              5,
		ThumbnailGenerationIntervalHours: 24,
		LogLevel:                         "info",
		UploadPath:                       "~/mango/uploads",
		PluginPath:                       "~/mango/plugins",
		DownloadTimeoutSeconds:           30,
		CacheEnabled:                     true,
		CacheSizeMBs:                     50,
		CacheLogEnabled:                  true,
		DisableLogin:                     false,
		DefaultUsername:                  "",
		AuthProxyHeaderName:              "",
		PluginUpdateIntervalHours:        24,
	}
}

// applyEnv overrides defaults with environment variables, matching the
// Crystal macro that reads ENV[k.upcase] for each option.
func (c *Config) applyEnv() {
	setStr := func(env string, dst *string) {
		if v, ok := os.LookupEnv(env); ok {
			*dst = v
		}
	}
	setInt := func(env string, dst *int) {
		if v, ok := os.LookupEnv(env); ok {
			if n, err := strconv.Atoi(v); err == nil {
				*dst = n
			}
		}
	}
	setBool := func(env string, dst *bool) {
		// env_is_true?: true when value is "true" or "1" (case-insensitive).
		if v, ok := os.LookupEnv(env); ok {
			lv := strings.ToLower(strings.TrimSpace(v))
			*dst = lv == "true" || lv == "1"
		}
	}

	setStr("HOST", &c.Host)
	setInt("PORT", &c.Port)
	setStr("BASE_URL", &c.BaseURL)
	setStr("SESSION_SECRET", &c.SessionSecret)
	setStr("LIBRARY_PATH", &c.LibraryPath)
	setStr("LIBRARY_CACHE_PATH", &c.LibraryCachePath)
	setStr("DB_PATH", &c.DBPath)
	setStr("QUEUE_DB_PATH", &c.QueueDBPath)
	setInt("SCAN_INTERVAL_MINUTES", &c.ScanIntervalMinutes)
	setInt("THUMBNAIL_GENERATION_INTERVAL_HOURS", &c.ThumbnailGenerationIntervalHours)
	setStr("LOG_LEVEL", &c.LogLevel)
	setStr("UPLOAD_PATH", &c.UploadPath)
	setStr("PLUGIN_PATH", &c.PluginPath)
	setInt("DOWNLOAD_TIMEOUT_SECONDS", &c.DownloadTimeoutSeconds)
	setBool("CACHE_ENABLED", &c.CacheEnabled)
	setInt("CACHE_SIZE_MBS", &c.CacheSizeMBs)
	setBool("CACHE_LOG_ENABLED", &c.CacheLogEnabled)
	setBool("DISABLE_LOGIN", &c.DisableLogin)
	setStr("DEFAULT_USERNAME", &c.DefaultUsername)
	setStr("AUTH_PROXY_HEADER_NAME", &c.AuthProxyHeaderName)
	setInt("PLUGIN_UPDATE_INTERVAL_HOURS", &c.PluginUpdateIntervalHours)
}

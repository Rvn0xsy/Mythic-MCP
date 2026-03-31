package config

import (
	"fmt"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
)

// TomlConfig is the structure of config.toml
type TomlConfig struct {
	Mythic struct {
		URL           string `toml:"url"`
		APIToken      string `toml:"api_token"`
		Username      string `toml:"username"`
		Password      string `toml:"password"`
		SSL           bool   `toml:"ssl"`
		SkipTLSVerify bool   `toml:"skip_tls_verify"`
		DocsURL       string `toml:"docs_url"`
	} `toml:"mythic"`

	Server struct {
		LogLevel  string `toml:"log_level"`
		Timeout   string `toml:"timeout"`
		TLSCert   string `toml:"tls_cert_file"`
		TLSKey    string `toml:"tls_key_file"`
		AuthToken string `toml:"auth_token"`
	} `toml:"server"`

	FileVending struct {
		Enabled        bool   `toml:"enabled"`
		BaseURL        string `toml:"base_url"`
		StoragePath    string `toml:"storage_path"`
		TokenExpiry    string `toml:"token_expiry"`
		MaxSizeMB      int    `toml:"max_size_mb"`
		CleanupDefault string `toml:"cleanup_interval"`
	} `toml:"file_vending"`
}

// Config holds the MCP server configuration
type Config struct {
	// Mythic connection settings
	MythicURL     string
	APIToken      string
	Username      string
	Password      string
	SSL           bool
	SkipTLSVerify bool

	// Server settings
	LogLevel    string
	Timeout     time.Duration
	TLSCertFile string
	TLSKeyFile  string
	AuthToken   string

	// File vending settings
	FileVendingEnabled  bool
	FileVendingBaseURL  string
	FileStoragePath     string
	FileTokenExpiry     time.Duration
	FileMaxSizeMB       int
	FileCleanupInterval time.Duration
}

// Load loads configuration from config.toml (if present) and environment variables.
// Environment variables take precedence over config.toml values.
func Load() (*Config, error) {
	cfg := &Config{
		// Defaults
		LogLevel:            "info",
		Timeout:             30 * time.Second,
		FileVendingEnabled:   true,
		FileStoragePath:      "/tmp/mythic-files",
		FileTokenExpiry:      5 * time.Minute,
		FileMaxSizeMB:        100,
		FileCleanupInterval:  60 * time.Second,
		SSL:                  true,
	}

	// Load from TOML if config file exists
	configFile := os.Getenv("MCP_CONFIG_FILE")
	if configFile == "" {
		configFile = "config.toml"
	}
	if data, err := os.ReadFile(configFile); err == nil {
		var tc TomlConfig
		if err := toml.Unmarshal(data, &tc); err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", configFile, err)
		}
		applyToml(cfg, &tc)
	}

	// Overlay environment variables (they take precedence)
	applyEnv(cfg)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// applyToml copies TOML config values into cfg (only if non-zero)
func applyToml(cfg *Config, tc *TomlConfig) {
	if tc.Mythic.URL != "" {
		cfg.MythicURL = tc.Mythic.URL
	}
	if tc.Mythic.APIToken != "" {
		cfg.APIToken = tc.Mythic.APIToken
	}
	if tc.Mythic.Username != "" {
		cfg.Username = tc.Mythic.Username
	}
	if tc.Mythic.Password != "" {
		cfg.Password = tc.Mythic.Password
	}
	cfg.SSL = tc.Mythic.SSL
	cfg.SkipTLSVerify = tc.Mythic.SkipTLSVerify
	if tc.Mythic.DocsURL != "" {
		// stored externally, not in Config struct
	}

	if tc.Server.LogLevel != "" {
		cfg.LogLevel = tc.Server.LogLevel
	}
	if tc.Server.Timeout != "" {
		cfg.Timeout, _ = time.ParseDuration(tc.Server.Timeout)
	}
	if tc.Server.TLSCert != "" {
		cfg.TLSCertFile = tc.Server.TLSCert
	}
	if tc.Server.TLSKey != "" {
		cfg.TLSKeyFile = tc.Server.TLSKey
	}
	if tc.Server.AuthToken != "" {
		cfg.AuthToken = tc.Server.AuthToken
	}

	if tc.FileVending.Enabled {
		cfg.FileVendingEnabled = tc.FileVending.Enabled
	}
	if tc.FileVending.BaseURL != "" {
		cfg.FileVendingBaseURL = tc.FileVending.BaseURL
	}
	if tc.FileVending.StoragePath != "" {
		cfg.FileStoragePath = tc.FileVending.StoragePath
	}
	if tc.FileVending.TokenExpiry != "" {
		cfg.FileTokenExpiry, _ = time.ParseDuration(tc.FileVending.TokenExpiry)
	}
	if tc.FileVending.MaxSizeMB != 0 {
		cfg.FileMaxSizeMB = tc.FileVending.MaxSizeMB
	}
	if tc.FileVending.CleanupDefault != "" {
		cfg.FileCleanupInterval, _ = time.ParseDuration(tc.FileVending.CleanupDefault)
	}
}

// applyEnv overlays environment variable values onto cfg
func applyEnv(cfg *Config) {
	if v := os.Getenv("MYTHIC_URL"); v != "" {
		cfg.MythicURL = v
	}
	if v := os.Getenv("MYTHIC_API_TOKEN"); v != "" {
		cfg.APIToken = v
	}
	if v := os.Getenv("MYTHIC_USERNAME"); v != "" {
		cfg.Username = v
	}
	if v := os.Getenv("MYTHIC_PASSWORD"); v != "" {
		cfg.Password = v
	}
	if v := os.Getenv("MYTHIC_SSL"); v != "" {
		cfg.SSL = v == "true" || v == "1" || v == "yes"
	}
	if v := os.Getenv("MYTHIC_SKIP_TLS_VERIFY"); v != "" {
		cfg.SkipTLSVerify = v == "true" || v == "1" || v == "yes"
	}

	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.LogLevel = v
	}
	if v := os.Getenv("TIMEOUT"); v != "" {
		cfg.Timeout, _ = time.ParseDuration(v)
	}
	if v := os.Getenv("MCP_TLS_CERT_FILE"); v != "" {
		cfg.TLSCertFile = v
	}
	if v := os.Getenv("MCP_TLS_KEY_FILE"); v != "" {
		cfg.TLSKeyFile = v
	}
	if v := os.Getenv("MCP_AUTH_TOKEN"); v != "" {
		cfg.AuthToken = v
	}

	if v := os.Getenv("FILE_VENDING_ENABLED"); v != "" {
		cfg.FileVendingEnabled = v == "true" || v == "1" || v == "yes"
	}
	if v := os.Getenv("FILE_VENDING_BASE_URL"); v != "" {
		cfg.FileVendingBaseURL = v
	}
	if v := os.Getenv("FILE_STORAGE_PATH"); v != "" {
		cfg.FileStoragePath = v
	}
	if v := os.Getenv("FILE_TOKEN_EXPIRY"); v != "" {
		cfg.FileTokenExpiry, _ = time.ParseDuration(v)
	}
	if v := os.Getenv("FILE_MAX_SIZE_MB"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.FileMaxSizeMB)
	}
	if v := os.Getenv("FILE_CLEANUP_INTERVAL"); v != "" {
		cfg.FileCleanupInterval, _ = time.ParseDuration(v)
	}
}

// Validate checks that required configuration is present.
// Only MYTHIC_URL is required at startup; credentials are optional because
// the user can authenticate later via the mythic_login MCP tool.
func (c *Config) Validate() error {
	if c.MythicURL == "" {
		return fmt.Errorf("MYTHIC_URL is required")
	}

	return nil
}

// getEnvString returns environment variable value or default
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool returns environment variable as bool or default
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

// getEnvDuration returns environment variable as duration or default
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return duration
}

// getEnvInt returns environment variable as int or default
func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var n int
	if _, err := fmt.Sscanf(value, "%d", &n); err != nil {
		return defaultValue
	}
	return n
}

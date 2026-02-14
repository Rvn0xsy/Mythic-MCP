package config

import (
	"fmt"
	"os"
	"time"
)

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
	LogLevel string
	Timeout  time.Duration

	// File vending settings
	FileVendingEnabled  bool
	FileVendingBaseURL  string
	FileStoragePath     string
	FileTokenExpiry     time.Duration
	FileMaxSizeMB       int
	FileCleanupInterval time.Duration
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() (*Config, error) {
	cfg := &Config{
		MythicURL:     os.Getenv("MYTHIC_URL"),
		APIToken:      os.Getenv("MYTHIC_API_TOKEN"),
		Username:      os.Getenv("MYTHIC_USERNAME"),
		Password:      os.Getenv("MYTHIC_PASSWORD"),
		SSL:           getEnvBool("MYTHIC_SSL", true),
		SkipTLSVerify: getEnvBool("MYTHIC_SKIP_TLS_VERIFY", false),
		LogLevel:      getEnvString("LOG_LEVEL", "info"),
		Timeout:       getEnvDuration("TIMEOUT", 30*time.Second),

		// File vending defaults
		FileVendingEnabled:  getEnvBool("FILE_VENDING_ENABLED", true),
		FileVendingBaseURL:  getEnvString("FILE_VENDING_BASE_URL", ""), // auto-detected if empty
		FileStoragePath:     getEnvString("FILE_STORAGE_PATH", "/tmp/mythic-files"),
		FileTokenExpiry:     getEnvDuration("FILE_TOKEN_EXPIRY", 5*time.Minute),
		FileMaxSizeMB:       getEnvInt("FILE_MAX_SIZE_MB", 100),
		FileCleanupInterval: getEnvDuration("FILE_CLEANUP_INTERVAL", 60*time.Second),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
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

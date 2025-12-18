package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Email    string
	Password string
	Region   string
}

// validRegions contains the list of valid Mazda API regions
var validRegions = map[string]bool{
	"MNAO": true, // Mazda North America Operations
	"MME":  true, // Mazda Europe
	"MJO":  true, // Mazda Japan
}

// Load loads configuration from file and environment variables
// Environment variables take precedence over file values
// configPath can be empty to use default location (~/.config/cx90/config.toml)
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("region", "MNAO")

	// Configure viper
	v.SetConfigType("toml")
	v.SetConfigName("config")

	if configPath != "" {
		// Use specified config file
		v.SetConfigFile(configPath)
	} else {
		// Use default config path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		configDir := filepath.Join(homeDir, ".config", "cx90")
		v.AddConfigPath(configDir)
	}

	// Try to read config file (don't fail if it doesn't exist)
	_ = v.ReadInConfig()

	// Bind environment variables
	v.SetEnvPrefix("MYMAZDA")
	v.AutomaticEnv()

	// Load config
	cfg := &Config{
		Email:    v.GetString("email"),
		Password: v.GetString("password"),
		Region:   v.GetString("region"),
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Email == "" {
		return fmt.Errorf("email is required")
	}
	if c.Password == "" {
		return fmt.Errorf("password is required")
	}
	if !validRegions[c.Region] {
		return fmt.Errorf("invalid region: %s (must be one of: MNAO, MME, MJO)", c.Region)
	}
	return nil
}

// DefaultConfigPath returns the default configuration file path
func DefaultConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "cx90", "config.toml"), nil
}

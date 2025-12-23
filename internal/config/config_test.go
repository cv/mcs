package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cv/mcs/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name       string
		envVars    map[string]string
		wantEmail  string
		wantRegion api.Region
		wantErr    bool
	}{
		{
			name: "load from environment variables",
			envVars: map[string]string{
				"MCS_EMAIL":    "test@example.com",
				"MCS_PASSWORD": "password123",
				"MCS_REGION":   "MNAO",
			},
			wantEmail:  "test@example.com",
			wantRegion: api.RegionMNAO,
			wantErr:    false,
		},
		{
			name: "default region when not specified",
			envVars: map[string]string{
				"MCS_EMAIL":    "test@example.com",
				"MCS_PASSWORD": "password123",
			},
			wantEmail:  "test@example.com",
			wantRegion: api.RegionMNAO,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars using t.Setenv (auto-restores after subtest)
			t.Setenv("MCS_EMAIL", "")
			t.Setenv("MCS_PASSWORD", "")
			t.Setenv("MCS_REGION", "")

			// Set test env vars
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			cfg, err := Load("")
			if tt.wantErr {
				require.Error(t, err, "Load() error = %v, wantErr %v")
			} else {
				require.NoError(t, err, "Load() error = %v, wantErr %v")
			}

			if !tt.wantErr {
				assert.Equal(t, tt.wantEmail, cfg.Email)
				assert.Equal(t, tt.wantRegion, cfg.Region)
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `
email = "file@example.com"
password = "filepassword"
region = "MME"
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err, "Failed to create test config file: %v")

	// Clear env vars to ensure file values are used
	t.Setenv("MCS_EMAIL", "")
	t.Setenv("MCS_PASSWORD", "")
	t.Setenv("MCS_REGION", "")

	cfg, err := Load(configPath)
	require.NoError(t, err, "Load() error = %v")

	assert.Equalf(t, "file@example.com", cfg.Email, "Load() Email = %v, want file@example.com", cfg.Email)
	assert.Equalf(t, api.RegionMME, cfg.Region, "Load() Region = %v, want MME", cfg.Region)
}

func TestEnvironmentOverridesFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	configContent := `
email = "file@example.com"
password = "filepassword"
region = "MME"
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err, "Failed to create test config file: %v")

	// Set test env vars (t.Setenv auto-restores after test)
	t.Setenv("MCS_EMAIL", "env@example.com")
	t.Setenv("MCS_REGION", "MNAO")

	cfg, err := Load(configPath)
	require.NoError(t, err, "Load() error = %v")

	// Env vars should override file values
	assert.Equalf(t, "env@example.com", cfg.Email, "Load() Email = %v, want env@example.com (env should override)", cfg.Email)
	assert.Equalf(t, api.RegionMNAO, cfg.Region, "Load() Region = %v, want MNAO (env should override)", cfg.Region)
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Email:    "test@example.com",
				Password: "password123",
				Region:   api.RegionMNAO,
			},
			wantErr: false,
		},
		{
			name: "missing email",
			config: &Config{
				Password: "password123",
				Region:   api.RegionMNAO,
			},
			wantErr: true,
		},
		{
			name: "missing password",
			config: &Config{
				Email:  "test@example.com",
				Region: api.RegionMNAO,
			},
			wantErr: true,
		},
		{
			name: "invalid region",
			config: &Config{
				Email:    "test@example.com",
				Password: "password123",
				Region:   "INVALID",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err, "Config.Validate() error = %v, wantErr %v")
			} else {
				require.NoError(t, err, "Config.Validate() error = %v, wantErr %v")
			}

		})
	}
}
func TestLoad_propagatesReadError(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Write a valid TOML file
	err := os.WriteFile(configPath, []byte(`email = "test@example.com"`), 0600)
	require.NoError(t, err, "Failed to create config file: %v")

	// Make the file unreadable to force a read error
	err = os.Chmod(configPath, 0o000)
	require.NoError(t, err, "Failed to chmod config file: %v")

	// Clear environment variables so file values are used
	t.Setenv("MCS_EMAIL", "")
	t.Setenv("MCS_PASSWORD", "")
	t.Setenv("MCS_REGION", "")

	// Load should return an error because reading the unreadable file failed
	cfg, err := Load(configPath)
	assert.Nil(t, cfg)
	require.Error(t, err, "expected error when reading unreadable config file")
}

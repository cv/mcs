package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save and restore env vars
	oldEmail := os.Getenv("MCS_EMAIL")
	oldPassword := os.Getenv("MCS_PASSWORD")
	oldRegion := os.Getenv("MCS_REGION")
	defer func() {
		os.Setenv("MCS_EMAIL", oldEmail)
		os.Setenv("MCS_PASSWORD", oldPassword)
		os.Setenv("MCS_REGION", oldRegion)
	}()

	tests := []struct {
		name        string
		envVars     map[string]string
		wantEmail   string
		wantRegion  string
		wantErr     bool
	}{
		{
			name: "load from environment variables",
			envVars: map[string]string{
				"MCS_EMAIL":    "test@example.com",
				"MCS_PASSWORD": "password123",
				"MCS_REGION":   "MNAO",
			},
			wantEmail:  "test@example.com",
			wantRegion: "MNAO",
			wantErr:    false,
		},
		{
			name: "default region when not specified",
			envVars: map[string]string{
				"MCS_EMAIL":    "test@example.com",
				"MCS_PASSWORD": "password123",
			},
			wantEmail:  "test@example.com",
			wantRegion: "MNAO",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars
			os.Unsetenv("MCS_EMAIL")
			os.Unsetenv("MCS_PASSWORD")
			os.Unsetenv("MCS_REGION")

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := Load("")
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if cfg.Email != tt.wantEmail {
					t.Errorf("Load() Email = %v, want %v", cfg.Email, tt.wantEmail)
				}
				if cfg.Region != tt.wantRegion {
					t.Errorf("Load() Region = %v, want %v", cfg.Region, tt.wantRegion)
				}
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
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Clear env vars to ensure file values are used
	oldEmail := os.Getenv("MCS_EMAIL")
	oldPassword := os.Getenv("MCS_PASSWORD")
	oldRegion := os.Getenv("MCS_REGION")
	os.Unsetenv("MCS_EMAIL")
	os.Unsetenv("MCS_PASSWORD")
	os.Unsetenv("MCS_REGION")
	defer func() {
		os.Setenv("MCS_EMAIL", oldEmail)
		os.Setenv("MCS_PASSWORD", oldPassword)
		os.Setenv("MCS_REGION", oldRegion)
	}()

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Email != "file@example.com" {
		t.Errorf("Load() Email = %v, want file@example.com", cfg.Email)
	}
	if cfg.Region != "MME" {
		t.Errorf("Load() Region = %v, want MME", cfg.Region)
	}
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
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Save and clear existing env vars, then set test values
	oldEmail := os.Getenv("MCS_EMAIL")
	oldRegion := os.Getenv("MCS_REGION")
	os.Setenv("MCS_EMAIL", "env@example.com")
	os.Setenv("MCS_REGION", "MNAO")
	defer func() {
		os.Setenv("MCS_EMAIL", oldEmail)
		os.Setenv("MCS_REGION", oldRegion)
	}()

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Env vars should override file values
	if cfg.Email != "env@example.com" {
		t.Errorf("Load() Email = %v, want env@example.com (env should override)", cfg.Email)
	}
	if cfg.Region != "MNAO" {
		t.Errorf("Load() Region = %v, want MNAO (env should override)", cfg.Region)
	}
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
				Region:   "MNAO",
			},
			wantErr: false,
		},
		{
			name: "missing email",
			config: &Config{
				Password: "password123",
				Region:   "MNAO",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			config: &Config{
				Email:  "test@example.com",
				Region: "MNAO",
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
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

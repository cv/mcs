package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TokenCache represents cached authentication credentials.
type TokenCache struct {
	AccessToken             string `json:"access_token"`
	AccessTokenExpirationTs int64  `json:"access_token_expiration_ts"`
	EncKey                  string `json:"enc_key"`
	SignKey                 string `json:"sign_key"`
}

// IsTokenValid checks if a token is present and not expired.
// This is a shared validation function used by both TokenCache and API Client.
func IsTokenValid(accessToken string, accessTokenExpirationTs int64) bool {
	if accessToken == "" {
		return false
	}
	if accessTokenExpirationTs == 0 {
		return false
	}
	if accessTokenExpirationTs <= time.Now().Unix() {
		return false
	}

	return true
}

// IsValid checks if the cached token is still valid.
func (tc *TokenCache) IsValid() bool {
	return IsTokenValid(tc.AccessToken, tc.AccessTokenExpirationTs)
}

// Load reads the token cache from the default location.
func Load() (*TokenCache, error) {
	path, err := getCachePath()
	if err != nil {
		return nil, err
	}

	return LoadFrom(path)
}

// LoadFrom reads the token cache from the specified file path.
func LoadFrom(path string) (*TokenCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No cache file exists yet
		}

		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	var cache TokenCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	return &cache, nil
}

// Save writes the token cache to the default location.
func Save(cache *TokenCache) error {
	path, err := getCachePath()
	if err != nil {
		return err
	}

	return SaveTo(cache, path)
}

// SaveTo writes the token cache to the specified file path.
func SaveTo(cache *TokenCache, path string) error {
	// Create cache directory if it doesn't exist.
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache: %w", err)
	}

	// Write with restricted permissions (owner read/write only).
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// getCachePath returns the path to the token cache file.
func getCachePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	return filepath.Join(homeDir, ".cache", "mcs", "token.json"), nil
}

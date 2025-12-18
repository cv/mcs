package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTokenCache_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		cache *TokenCache
		want  bool
	}{
		{
			name: "valid token",
			cache: &TokenCache{
				AccessToken:             "test-token",
				AccessTokenExpirationTs: time.Now().Unix() + 3600,
				EncKey:                  "test-enc-key",
				SignKey:                 "test-sign-key",
			},
			want: true,
		},
		{
			name: "expired token",
			cache: &TokenCache{
				AccessToken:             "test-token",
				AccessTokenExpirationTs: time.Now().Unix() - 3600,
				EncKey:                  "test-enc-key",
				SignKey:                 "test-sign-key",
			},
			want: false,
		},
		{
			name: "empty token",
			cache: &TokenCache{
				AccessToken:             "",
				AccessTokenExpirationTs: time.Now().Unix() + 3600,
				EncKey:                  "test-enc-key",
				SignKey:                 "test-sign-key",
			},
			want: false,
		},
		{
			name: "zero expiration",
			cache: &TokenCache{
				AccessToken:             "test-token",
				AccessTokenExpirationTs: 0,
				EncKey:                  "test-enc-key",
				SignKey:                 "test-sign-key",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cache.IsValid(); got != tt.want {
				t.Errorf("TokenCache.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	originalHomeDir := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHomeDir)

	testCache := &TokenCache{
		AccessToken:             "test-token-123",
		AccessTokenExpirationTs: time.Now().Unix() + 3600,
		EncKey:                  "test-enc-key-456",
		SignKey:                 "test-sign-key-789",
	}

	// Test Save
	err := Save(testCache)
	if err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify cache file was created with correct permissions
	cachePath := filepath.Join(tmpDir, ".cache", "mcs", "token.json")
	info, err := os.Stat(cachePath)
	if err != nil {
		t.Fatalf("Cache file not created: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("Cache file has incorrect permissions: got %v, want 0600", info.Mode().Perm())
	}

	// Test Load
	loadedCache, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if loadedCache == nil {
		t.Fatal("Load() returned nil cache")
	}

	// Verify loaded data matches saved data
	if loadedCache.AccessToken != testCache.AccessToken {
		t.Errorf("AccessToken mismatch: got %v, want %v", loadedCache.AccessToken, testCache.AccessToken)
	}
	if loadedCache.AccessTokenExpirationTs != testCache.AccessTokenExpirationTs {
		t.Errorf("AccessTokenExpirationTs mismatch: got %v, want %v", loadedCache.AccessTokenExpirationTs, testCache.AccessTokenExpirationTs)
	}
	if loadedCache.EncKey != testCache.EncKey {
		t.Errorf("EncKey mismatch: got %v, want %v", loadedCache.EncKey, testCache.EncKey)
	}
	if loadedCache.SignKey != testCache.SignKey {
		t.Errorf("SignKey mismatch: got %v, want %v", loadedCache.SignKey, testCache.SignKey)
	}
}

func TestLoad_NoCache(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	originalHomeDir := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHomeDir)

	// Load without any cache file
	cache, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}
	if cache != nil {
		t.Errorf("Load() should return nil when no cache exists, got %v", cache)
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	originalHomeDir := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHomeDir)

	// Create cache directory and invalid JSON file
	cachePath := filepath.Join(tmpDir, ".cache", "mcs", "token.json")
	err := os.MkdirAll(filepath.Dir(cachePath), 0700)
	if err != nil {
		t.Fatalf("Failed to create cache directory: %v", err)
	}
	err = os.WriteFile(cachePath, []byte("invalid json"), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid cache file: %v", err)
	}

	// Load should fail with parse error
	_, err = Load()
	if err == nil {
		t.Error("Load() should fail with invalid JSON")
	}
}

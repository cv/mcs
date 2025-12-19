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
	t.Setenv("HOME", tmpDir)

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
	t.Setenv("HOME", tmpDir)

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
	t.Setenv("HOME", tmpDir)

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

func TestIsTokenValid(t *testing.T) {
	tests := []struct {
		name         string
		accessToken  string
		expirationTs int64
		want         bool
	}{
		{
			name:         "valid token",
			accessToken:  "test-token",
			expirationTs: time.Now().Unix() + 3600,
			want:         true,
		},
		{
			name:         "expired token",
			accessToken:  "test-token",
			expirationTs: time.Now().Unix() - 3600,
			want:         false,
		},
		{
			name:         "empty token",
			accessToken:  "",
			expirationTs: time.Now().Unix() + 3600,
			want:         false,
		},
		{
			name:         "zero expiration",
			accessToken:  "test-token",
			expirationTs: 0,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTokenValid(tt.accessToken, tt.expirationTs); got != tt.want {
				t.Errorf("IsTokenValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCachePersistence_MultipleSaveLoad tests save and load cycle multiple times
func TestCachePersistence_MultipleSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// First save
	cache1 := &TokenCache{
		AccessToken:             "token-1",
		AccessTokenExpirationTs: time.Now().Unix() + 1000,
		EncKey:                  "enckey-1",
		SignKey:                 "signkey-1",
	}
	if err := Save(cache1); err != nil {
		t.Fatalf("First Save() failed: %v", err)
	}

	// First load
	loaded1, err := Load()
	if err != nil {
		t.Fatalf("First Load() failed: %v", err)
	}
	if loaded1.AccessToken != "token-1" {
		t.Errorf("First load: expected token-1, got %s", loaded1.AccessToken)
	}

	// Second save (overwrite)
	cache2 := &TokenCache{
		AccessToken:             "token-2",
		AccessTokenExpirationTs: time.Now().Unix() + 2000,
		EncKey:                  "enckey-2",
		SignKey:                 "signkey-2",
	}
	if err := Save(cache2); err != nil {
		t.Fatalf("Second Save() failed: %v", err)
	}

	// Second load
	loaded2, err := Load()
	if err != nil {
		t.Fatalf("Second Load() failed: %v", err)
	}
	if loaded2.AccessToken != "token-2" {
		t.Errorf("Second load: expected token-2, got %s", loaded2.AccessToken)
	}

	// Verify old values are gone
	if loaded2.AccessToken == "token-1" {
		t.Error("Old cache values should be overwritten")
	}
}

// TestCachePersistence_ConcurrentAccess tests concurrent save/load operations
func TestCachePersistence_ConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Save initial cache
	initialCache := &TokenCache{
		AccessToken:             "initial-token",
		AccessTokenExpirationTs: time.Now().Unix() + 3600,
		EncKey:                  "initial-enc",
		SignKey:                 "initial-sign",
	}
	if err := Save(initialCache); err != nil {
		t.Fatalf("Initial Save() failed: %v", err)
	}

	// Try concurrent loads (should all succeed)
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func(id int) {
			cache, err := Load()
			if err != nil {
				t.Errorf("Concurrent load %d failed: %v", id, err)
			}
			if cache == nil {
				t.Errorf("Concurrent load %d returned nil", id)
			}
			done <- true
		}(i)
	}

	// Wait for all loads to complete
	for i := 0; i < 3; i++ {
		<-done
	}
}

// TestCachePersistence_CorruptedData tests handling of corrupted cache file
func TestCachePersistence_CorruptedData(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create cache directory and write corrupted data
	cachePath := filepath.Join(tmpDir, ".cache", "mcs", "token.json")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0700); err != nil {
		t.Fatalf("Failed to create cache directory: %v", err)
	}

	// Write truly corrupted JSON (not valid JSON at all)
	corruptedJSON := `{this is not valid json at all!!!`
	if err := os.WriteFile(cachePath, []byte(corruptedJSON), 0600); err != nil {
		t.Fatalf("Failed to write corrupted cache: %v", err)
	}

	// Load should fail gracefully
	_, err := Load()
	if err == nil {
		t.Error("Expected error when loading corrupted cache")
	}
}

// TestCachePersistence_PartialData tests cache with missing fields
func TestCachePersistence_PartialData(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Create cache with only some fields
	cachePath := filepath.Join(tmpDir, ".cache", "mcs", "token.json")
	if err := os.MkdirAll(filepath.Dir(cachePath), 0700); err != nil {
		t.Fatalf("Failed to create cache directory: %v", err)
	}

	// Write partial JSON (missing signKey)
	partialJSON := `{"accessToken": "partial-token", "accessTokenExpirationTs": 1234567890, "encKey": "partial-enc"}`
	if err := os.WriteFile(cachePath, []byte(partialJSON), 0600); err != nil {
		t.Fatalf("Failed to write partial cache: %v", err)
	}

	// Load should succeed but cache should be invalid (missing signKey)
	cache, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cache.SignKey != "" {
		t.Errorf("Expected empty SignKey, got %s", cache.SignKey)
	}

	// Note: IsValid only checks token validity, not presence of keys
	// So this cache will be considered "valid" even though signKey is missing
	// The actual validation happens when the CLI tries to use the credentials
}

// TestCacheValidation_EdgeCases tests edge cases in cache validation
func TestCacheValidation_EdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		cache *TokenCache
		want  bool
	}{
		{
			name: "token expiring in 1 second (still valid)",
			cache: &TokenCache{
				AccessToken:             "soon-to-expire",
				AccessTokenExpirationTs: time.Now().Unix() + 1,
				EncKey:                  "key",
				SignKey:                 "sign",
			},
			want: true,
		},
		{
			name: "token expired 1 second ago",
			cache: &TokenCache{
				AccessToken:             "just-expired",
				AccessTokenExpirationTs: time.Now().Unix() - 1,
				EncKey:                  "key",
				SignKey:                 "sign",
			},
			want: false,
		},
		{
			name: "missing enc key (still considered valid by IsValid - only checks token)",
			cache: &TokenCache{
				AccessToken:             "token",
				AccessTokenExpirationTs: time.Now().Unix() + 3600,
				EncKey:                  "",
				SignKey:                 "sign",
			},
			want: true, // IsValid only checks token, not keys
		},
		{
			name: "missing sign key (still considered valid by IsValid - only checks token)",
			cache: &TokenCache{
				AccessToken:             "token",
				AccessTokenExpirationTs: time.Now().Unix() + 3600,
				EncKey:                  "key",
				SignKey:                 "",
			},
			want: true, // IsValid only checks token, not keys
		},
		{
			name: "all fields empty",
			cache: &TokenCache{
				AccessToken:             "",
				AccessTokenExpirationTs: 0,
				EncKey:                  "",
				SignKey:                 "",
			},
			want: false,
		},
		{
			name: "very far future expiration",
			cache: &TokenCache{
				AccessToken:             "long-lived",
				AccessTokenExpirationTs: time.Now().Unix() + 31536000, // 1 year
				EncKey:                  "key",
				SignKey:                 "sign",
			},
			want: true,
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

// TestCachePersistence_FilePermissions tests that cache file has correct permissions
func TestCachePersistence_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	cache := &TokenCache{
		AccessToken:             "secure-token",
		AccessTokenExpirationTs: time.Now().Unix() + 3600,
		EncKey:                  "secure-enc",
		SignKey:                 "secure-sign",
	}

	if err := Save(cache); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	cachePath := filepath.Join(tmpDir, ".cache", "mcs", "token.json")
	info, err := os.Stat(cachePath)
	if err != nil {
		t.Fatalf("Failed to stat cache file: %v", err)
	}

	// Verify file permissions are 0600 (read/write for owner only)
	if info.Mode().Perm() != 0600 {
		t.Errorf("Cache file has incorrect permissions: got %v, want 0600", info.Mode().Perm())
	}
}

// TestCachePersistence_DirectoryCreation tests that cache directory is created if it doesn't exist
func TestCachePersistence_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Verify cache directory doesn't exist yet
	cachePath := filepath.Join(tmpDir, ".cache", "mcs")
	if _, err := os.Stat(cachePath); err == nil {
		t.Error("Cache directory should not exist yet")
	}

	// Save cache (should create directory)
	cache := &TokenCache{
		AccessToken:             "new-token",
		AccessTokenExpirationTs: time.Now().Unix() + 3600,
		EncKey:                  "new-enc",
		SignKey:                 "new-sign",
	}

	if err := Save(cache); err != nil {
		t.Fatalf("Save() failed: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(cachePath); err != nil {
		t.Errorf("Cache directory should have been created: %v", err)
	}
}

// TestCachePersistence_EmptyHomeDir tests behavior when HOME is not set
func TestCachePersistence_EmptyHomeDir(t *testing.T) {
	// Note: This test may not work on all systems
	// We can't truly unset HOME in Go tests, so we set it to empty
	t.Setenv("HOME", "")

	cache := &TokenCache{
		AccessToken:             "test-token",
		AccessTokenExpirationTs: time.Now().Unix() + 3600,
		EncKey:                  "test-enc",
		SignKey:                 "test-sign",
	}

	// Save should fail gracefully when HOME is empty
	err := Save(cache)
	if err == nil {
		t.Error("Expected error when HOME is empty, got nil")
	}
}

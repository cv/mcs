package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenCache_IsValid(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			assert.Equal(t, tt.want, tt.cache.IsValid())
		})
	}
}

func TestSaveAndLoad(t *testing.T) {
	t.Parallel()
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "token.json")

	testCache := &TokenCache{
		AccessToken:             "test-token-123",
		AccessTokenExpirationTs: time.Now().Unix() + 3600,
		EncKey:                  "test-enc-key-456",
		SignKey:                 "test-sign-key-789",
	}

	// Test SaveTo
	err := SaveTo(testCache, cachePath)
	require.NoError(t, err, "SaveTo() failed: %v")

	// Verify cache file was created with correct permissions
	info, err := os.Stat(cachePath)
	require.NoError(t, err, "Cache file not created: %v")
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm(), "Cache file has incorrect permissions")

	// Test LoadFrom
	loadedCache, err := LoadFrom(cachePath)
	require.NoError(t, err, "LoadFrom() failed: %v")
	require.NotNil(t, loadedCache, "LoadFrom() returned nil cache")

	// Verify loaded data matches saved data
	assert.Equal(t, testCache.AccessToken, loadedCache.AccessToken)
	assert.Equal(t, testCache.AccessTokenExpirationTs, loadedCache.AccessTokenExpirationTs)
	assert.Equal(t, testCache.EncKey, loadedCache.EncKey)
	assert.Equal(t, testCache.SignKey, loadedCache.SignKey)
}

func TestLoad_NoCache(t *testing.T) {
	t.Parallel()
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "nonexistent.json")

	// Load without any cache file
	cache, err := LoadFrom(cachePath)
	require.NoError(t, err, "LoadFrom() failed: %v")
	assert.Nil(t, cache)
}

func TestLoad_InvalidJSON(t *testing.T) {
	t.Parallel()
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "token.json")

	// Write invalid JSON file
	err := os.WriteFile(cachePath, []byte("invalid json"), 0600)
	require.NoError(t, err, "Failed to write invalid cache file: %v")

	// Load should fail with parse error
	_, err = LoadFrom(cachePath)
	require.Error(t, err, "LoadFrom() should fail with invalid JSON")
}

func TestIsTokenValid(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			assert.Equal(t, tt.want, IsTokenValid(tt.accessToken, tt.expirationTs))
		})
	}
}

// TestCachePersistence_MultipleSaveLoad tests save and load cycle multiple times.
func TestCachePersistence_MultipleSaveLoad(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "token.json")

	// First save
	cache1 := &TokenCache{
		AccessToken:             "token-1",
		AccessTokenExpirationTs: time.Now().Unix() + 1000,
		EncKey:                  "enckey-1",
		SignKey:                 "signkey-1",
	}
	err := SaveTo(cache1, cachePath)
	require.NoError(t, err, "First SaveTo() failed: %v")

	// First load
	loaded1, err := LoadFrom(cachePath)
	require.NoError(t, err, "First LoadFrom() failed: %v")
	assert.Equalf(t, "token-1", loaded1.AccessToken, "First load: expected token-1, got %s", loaded1.AccessToken)

	// Second save (overwrite)
	cache2 := &TokenCache{
		AccessToken:             "token-2",
		AccessTokenExpirationTs: time.Now().Unix() + 2000,
		EncKey:                  "enckey-2",
		SignKey:                 "signkey-2",
	}
	err = SaveTo(cache2, cachePath)
	require.NoError(t, err, "Second SaveTo() failed: %v")

	// Second load
	loaded2, err := LoadFrom(cachePath)
	require.NoError(t, err, "Second LoadFrom() failed: %v")
	assert.Equalf(t, "token-2", loaded2.AccessToken, "Second load: expected token-2, got %s", loaded2.AccessToken)

	// Verify old values are gone
	assert.NotEqual(t, "token-1", loaded2.AccessToken, "Old cache values should be overwritten")
}

// TestCachePersistence_ConcurrentAccess tests concurrent save/load operations.
func TestCachePersistence_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "token.json")

	// Save initial cache
	initialCache := &TokenCache{
		AccessToken:             "initial-token",
		AccessTokenExpirationTs: time.Now().Unix() + 3600,
		EncKey:                  "initial-enc",
		SignKey:                 "initial-sign",
	}
	err := SaveTo(initialCache, cachePath)
	require.NoError(t, err, "Initial SaveTo() failed: %v")

	// Try concurrent loads (should all succeed)
	done := make(chan bool, 3)
	for i := range 3 {
		go func(id int) {
			cache, err := LoadFrom(cachePath)
			assert.NoErrorf(t, err, "Concurrent load %d failed: %v", id, err)
			assert.NotNilf(t, cache, "Concurrent load %d returned nil", id)
			done <- true
		}(i)
	}

	// Wait for all loads to complete
	for range 3 {
		<-done
	}
}

// TestCachePersistence_CorruptedData tests handling of corrupted cache file.
func TestCachePersistence_CorruptedData(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "token.json")

	// Write truly corrupted JSON (not valid JSON at all)
	corruptedJSON := `{this is not valid json at all!!!`
	err := os.WriteFile(cachePath, []byte(corruptedJSON), 0600)
	require.NoError(t, err, "Failed to write corrupted cache: %v")

	// Load should fail gracefully
	_, err = LoadFrom(cachePath)
	require.Error(t, err, "Expected error when loading corrupted cache")
}

// TestCachePersistence_PartialData tests cache with missing fields.
func TestCachePersistence_PartialData(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "token.json")

	// Write partial JSON (missing signKey)
	partialJSON := `{"accessToken": "partial-token", "accessTokenExpirationTs": 1234567890, "encKey": "partial-enc"}`
	err := os.WriteFile(cachePath, []byte(partialJSON), 0600)
	require.NoError(t, err, "Failed to write partial cache: %v")

	// Load should succeed but cache should be invalid (missing signKey)
	cache, err := LoadFrom(cachePath)
	require.NoError(t, err, "LoadFrom() failed: %v")

	assert.Emptyf(t, cache.SignKey, "Expected empty SignKey, got %s", cache.SignKey)

	// Note: IsValid only checks token validity, not presence of keys
	// So this cache will be considered "valid" even though signKey is missing
	// The actual validation happens when the CLI tries to use the credentials
}

// TestCacheValidation_EdgeCases tests edge cases in cache validation.
func TestCacheValidation_EdgeCases(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			assert.Equal(t, tt.want, tt.cache.IsValid())
		})
	}
}

// TestCachePersistence_FilePermissions tests that cache file has correct permissions.
func TestCachePersistence_FilePermissions(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "token.json")

	cache := &TokenCache{
		AccessToken:             "secure-token",
		AccessTokenExpirationTs: time.Now().Unix() + 3600,
		EncKey:                  "secure-enc",
		SignKey:                 "secure-sign",
	}

	err := SaveTo(cache, cachePath)
	require.NoError(t, err, "SaveTo() failed: %v")

	info, err := os.Stat(cachePath)
	require.NoError(t, err, "Failed to stat cache file: %v")

	// Verify file permissions are 0600 (read/write for owner only)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm(), "Cache file has incorrect permissions")
}

// TestCachePersistence_DirectoryCreation tests that cache directory is created if it doesn't exist.
func TestCachePersistence_DirectoryCreation(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	cacheDir := filepath.Join(tmpDir, "subdir", "nested")
	cachePath := filepath.Join(cacheDir, "token.json")

	// Verify cache directory doesn't exist yet
	assert.NoFileExists(t, cacheDir)

	// Save cache (should create directory)
	cache := &TokenCache{
		AccessToken:             "new-token",
		AccessTokenExpirationTs: time.Now().Unix() + 3600,
		EncKey:                  "new-enc",
		SignKey:                 "new-sign",
	}

	err := SaveTo(cache, cachePath)
	require.NoError(t, err, "SaveTo() failed: %v")

	// Verify directory was created
	assert.DirExists(t, cacheDir)
}

// TestCachePersistence_EmptyHomeDir tests behavior when HOME is not set.
// This test specifically verifies the default Save() function which relies on HOME.
// Note: This test cannot use t.Parallel() because it uses t.Setenv.
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
	require.Error(t, err, "Expected error when HOME is empty, got nil")
}

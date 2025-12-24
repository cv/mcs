package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cv/mcs/internal/api"
	"github.com/cv/mcs/internal/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testConfig holds paths for a test's isolated config and cache files.
type testConfig struct {
	configFile string
	cacheFile  string
	ctx        context.Context
}

// newTestConfig creates isolated config and cache files for a parallel test.
// The config file is pre-populated with valid credentials.
func newTestConfig(t *testing.T) *testConfig {
	t.Helper()
	tmpDir := t.TempDir()

	configFile := filepath.Join(tmpDir, "config.toml")
	cacheFile := filepath.Join(tmpDir, "cache", "token.json")

	// Create config file with valid credentials.
	configContent := `
email = "test@example.com"
password = "test-password"
region = "MNAO"
`
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err, "Failed to create config file")

	cfg := &CLIConfig{
		ConfigFile: configFile,
		CacheFile:  cacheFile,
	}

	return &testConfig{
		configFile: configFile,
		cacheFile:  cacheFile,
		ctx:        ContextWithConfig(context.Background(), cfg),
	}
}

// TestCreateAPIClient_WithValidConfig tests creating an API client with valid config.
func TestCreateAPIClient_WithValidConfig(t *testing.T) {
	t.Parallel()
	tc := newTestConfig(t)

	client, err := createAPIClient(tc.ctx)
	require.NoError(t, err, "Failed to create API client")
	require.NotNil(t, client, "Expected client to be created, got nil")
}

// TestCreateAPIClient_WithInvalidRegion tests error handling for invalid region.
func TestCreateAPIClient_WithInvalidRegion(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()

	// Create config file with invalid region.
	configFile := filepath.Join(tmpDir, "config.toml")
	configContent := `
email = "test@example.com"
password = "test-password"
region = "INVALID"
`
	err := os.WriteFile(configFile, []byte(configContent), 0600)
	require.NoError(t, err, "Failed to create config file")

	cfg := &CLIConfig{ConfigFile: configFile}
	ctx := ContextWithConfig(context.Background(), cfg)

	_, err = createAPIClient(ctx)
	require.Error(t, err, "Expected error with invalid region, got nil")
}

// TestCreateAPIClient_WithConfigFile tests loading config from file.
func TestCreateAPIClient_WithConfigFile(t *testing.T) {
	t.Parallel()
	tc := newTestConfig(t)

	client, err := createAPIClient(tc.ctx)
	require.NoError(t, err, "Failed to create API client")
	require.NotNil(t, client, "Expected client to be created, got nil")
}

// TestCreateAPIClient_WithCachedCredentials tests using cached credentials.
func TestCreateAPIClient_WithCachedCredentials(t *testing.T) {
	t.Parallel()
	tc := newTestConfig(t)

	// Create valid cached credentials.
	cachedToken := &cache.TokenCache{
		AccessToken:             "cached-token-12345",
		AccessTokenExpirationTs: time.Now().Unix() + 3600,
		EncKey:                  "cached-enc-key",
		SignKey:                 "cached-sign-key",
	}
	err := cache.SaveTo(cachedToken, tc.cacheFile)
	require.NoError(t, err, "Failed to save cached token")

	// Create API client (should use cached credentials).
	client, err := createAPIClient(tc.ctx)
	require.NoError(t, err, "Failed to create API client")

	// Verify cached credentials were loaded.
	accessToken, expirationTs, encKey, signKey := client.GetCredentials()
	assert.Equal(t, "cached-token-12345", accessToken)
	assert.Equal(t, cachedToken.AccessTokenExpirationTs, expirationTs)
	assert.Equal(t, "cached-enc-key", encKey)
	assert.Equal(t, "cached-sign-key", signKey)
}

// TestCreateAPIClient_WithExpiredCache tests that expired cache is ignored.
func TestCreateAPIClient_WithExpiredCache(t *testing.T) {
	t.Parallel()
	tc := newTestConfig(t)

	// Create expired cached credentials.
	expiredToken := &cache.TokenCache{
		AccessToken:             "expired-token",
		AccessTokenExpirationTs: time.Now().Unix() - 3600, // Expired 1 hour ago.
		EncKey:                  "old-enc-key",
		SignKey:                 "old-sign-key",
	}
	err := cache.SaveTo(expiredToken, tc.cacheFile)
	require.NoError(t, err, "Failed to save expired token")

	// Create API client (should ignore expired cache).
	client, err := createAPIClient(tc.ctx)
	require.NoError(t, err, "Failed to create API client")

	// Verify expired credentials were not loaded.
	accessToken, _, _, _ := client.GetCredentials()
	assert.NotEqual(t, "expired-token", accessToken, "Expired cache should not be loaded")
}

// TestSaveClientCache_ValidCredentials tests that client credentials are saved to cache.
func TestSaveClientCache_ValidCredentials(t *testing.T) {
	t.Parallel()
	tc := newTestConfig(t)

	// Create client with valid credentials.
	client, err := api.NewClient("test@example.com", "password", api.RegionMNAO)
	require.NoError(t, err, "Failed to create client")

	// Set credentials.
	futureTimestamp := time.Now().Unix() + 3600
	client.SetCachedCredentials("test-token", futureTimestamp, "testenckey123456", "testsignkey12345")

	// Save client cache.
	saveClientCache(tc.ctx, client)

	// Verify cache was saved.
	loadedCache, err := cache.LoadFrom(tc.cacheFile)
	require.NoError(t, err, "Failed to load cache")
	require.NotNil(t, loadedCache, "Expected cache to be saved, got nil")

	assert.Equal(t, "test-token", loadedCache.AccessToken)
	assert.Equal(t, "testenckey123456", loadedCache.EncKey)
	assert.Equal(t, "testsignkey12345", loadedCache.SignKey)
}

// TestSaveClientCache_EmptyCredentials tests that empty credentials are not saved.
func TestSaveClientCache_EmptyCredentials(t *testing.T) {
	t.Parallel()
	tc := newTestConfig(t)

	// Create client without credentials.
	client, err := api.NewClient("test@example.com", "password", api.RegionMNAO)
	require.NoError(t, err, "Failed to create client")

	// Save client cache (should not save empty credentials).
	saveClientCache(tc.ctx, client)

	// Verify cache does not exist.
	loadedCache, err := cache.LoadFrom(tc.cacheFile)
	require.NoError(t, err, "Failed to load cache")
	assert.Nil(t, loadedCache)
}

// TestSaveClientCache_PartialCredentials tests that partial credentials are not saved.
func TestSaveClientCache_PartialCredentials(t *testing.T) {
	t.Parallel()
	tc := newTestConfig(t)

	// Create client with partial credentials.
	client, err := api.NewClient("test@example.com", "password", api.RegionMNAO)
	require.NoError(t, err, "Failed to create client")

	// Set only some credentials (missing signKey).
	futureTimestamp := time.Now().Unix() + 3600
	client.SetCachedCredentials("partial-token", futureTimestamp, "partial-enc", "")

	// Save client cache (should not save due to missing field).
	saveClientCache(tc.ctx, client)

	// Verify cache does not exist (missing signKey).
	loadedCache, err := cache.LoadFrom(tc.cacheFile)
	require.NoError(t, err, "Failed to load cache")
	assert.Nil(t, loadedCache)
}

// TestCreateAPIClient_EnvVarOverridesFile tests that env vars override config file.
// This test specifically verifies env var behavior, so it cannot run in parallel.
func TestCreateAPIClient_EnvVarOverridesFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	configPath := filepath.Join(tmpDir, "test-config.toml")

	// Create config file with different region.
	configContent := `
email = "file@example.com"
password = "file-password"
region = "MME"
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err, "Failed to create config file")

	// Set env vars (should override file).
	t.Setenv("MCS_EMAIL", "env@example.com")
	t.Setenv("MCS_PASSWORD", "env-password")
	t.Setenv("MCS_REGION", "MNAO")

	cfg := &CLIConfig{ConfigFile: configPath}
	ctx := ContextWithConfig(context.Background(), cfg)
	client, err := createAPIClient(ctx)
	require.NoError(t, err, "Failed to create API client")
	assert.NotNil(t, client)
}

// TestCreateAPIClient_MissingCredentials tests error when credentials are missing.
// This test specifically verifies env var behavior, so it cannot run in parallel.
func TestCreateAPIClient_MissingCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Clear env vars.
	t.Setenv("MCS_EMAIL", "")
	t.Setenv("MCS_PASSWORD", "")
	t.Setenv("MCS_REGION", "")

	cfg := &CLIConfig{}
	ctx := ContextWithConfig(context.Background(), cfg)
	_, err := createAPIClient(ctx)
	require.Error(t, err, "Expected error with missing credentials, got nil")
}

// TestVehicleInfo_InternalVINType tests that VehicleInfo.InternalVIN uses api.InternalVIN type.
func TestVehicleInfo_InternalVINType(t *testing.T) {
	t.Parallel()
	vehicleInfo := VehicleInfo{
		InternalVIN: api.InternalVIN("test-vin-123"),
		VIN:         "JM3XXXXXXXXXX1234",
		Nickname:    "Test Vehicle",
		ModelName:   "CX-90",
		ModelYear:   "2024",
	}

	// Verify InternalVIN type.
	var _ = vehicleInfo.InternalVIN

	// Verify string conversion.
	vinString := vehicleInfo.InternalVIN.String()
	assert.Equal(t, "test-vin-123", vinString)
	assert.Equal(t, "test-vin-123", string(vehicleInfo.InternalVIN))
}

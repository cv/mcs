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

// testContextWithConfig creates a context with the given config file path.
func testContextWithConfig(configFile string) context.Context {
	cfg := &CLIConfig{ConfigFile: configFile}

	return ContextWithConfig(context.Background(), cfg)
}

// TestCreateAPIClient_WithValidConfig tests creating an API client with valid config.
func TestCreateAPIClient_WithValidConfig(t *testing.T) {
	// Setup: Create temporary environment.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Configure environment variables.
	t.Setenv("MCS_EMAIL", "test@example.com")
	t.Setenv("MCS_PASSWORD", "test-password")
	t.Setenv("MCS_REGION", "MNAO")

	// Test: Create API client.
	ctx := testContextWithConfig("")
	client, err := createAPIClient(ctx)
	require.NoError(t, err, "Failed to create API client: %v")

	require.NotNil(t, client, "Expected client to be created, got nil")
}

// TestCreateAPIClient_WithInvalidRegion tests error handling for invalid region.
func TestCreateAPIClient_WithInvalidRegion(t *testing.T) {
	// Setup: Create temporary environment.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Configure with invalid region.
	t.Setenv("MCS_EMAIL", "test@example.com")
	t.Setenv("MCS_PASSWORD", "test-password")
	t.Setenv("MCS_REGION", "INVALID")

	// Test: Create API client should fail.
	ctx := testContextWithConfig("")
	_, err := createAPIClient(ctx)
	require.Error(t, err, "Expected error with invalid region, got nil")
}

// TestCreateAPIClient_WithConfigFile tests loading config from file.
func TestCreateAPIClient_WithConfigFile(t *testing.T) {
	// Setup: Create temporary environment.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	configPath := filepath.Join(tmpDir, "test-config.toml")

	// Setup: Create config file.
	configContent := `
email = "file@example.com"
password = "file-password"
region = "MNAO"
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err, "Failed to create config file: %v")

	// Clear env vars.
	t.Setenv("MCS_EMAIL", "")
	t.Setenv("MCS_PASSWORD", "")
	t.Setenv("MCS_REGION", "")

	// Test: Create API client from file.
	ctx := testContextWithConfig(configPath)
	client, err := createAPIClient(ctx)
	require.NoError(t, err, "Failed to create API client: %v")

	require.NotNil(t, client, "Expected client to be created, got nil")
}

// TestCreateAPIClient_WithCachedCredentials tests using cached credentials.
func TestCreateAPIClient_WithCachedCredentials(t *testing.T) {
	// Setup: Create temporary environment.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Configure environment variables.
	t.Setenv("MCS_EMAIL", "cached@example.com")
	t.Setenv("MCS_PASSWORD", "cached-password")
	t.Setenv("MCS_REGION", "MNAO")

	// Setup: Create valid cached credentials.
	cachedToken := &cache.TokenCache{
		AccessToken:             "cached-token-12345",
		AccessTokenExpirationTs: time.Now().Unix() + 3600,
		EncKey:                  "cached-enc-key",
		SignKey:                 "cached-sign-key",
	}
	err := cache.Save(cachedToken)
	require.NoError(t, err, "Failed to save cached token: %v")

	// Test: Create API client (should use cached credentials).
	ctx := testContextWithConfig("")
	client, err := createAPIClient(ctx)
	require.NoError(t, err, "Failed to create API client: %v")

	// Verify: Check that cached credentials were loaded.
	accessToken, expirationTs, encKey, signKey := client.GetCredentials()

	assert.Equalf(t, "cached-token-12345", accessToken, "Expected cached access token, got %s", accessToken)
	assert.Equal(t, cachedToken.AccessTokenExpirationTs, expirationTs)
	assert.Equalf(t, "cached-enc-key", encKey, "Expected cached enc key, got %s", encKey)
	assert.Equalf(t, "cached-sign-key", signKey, "Expected cached sign key, got %s", signKey)
}

// TestCreateAPIClient_WithExpiredCache tests that expired cache is ignored.
func TestCreateAPIClient_WithExpiredCache(t *testing.T) {
	// Setup: Create temporary environment.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Configure environment variables.
	t.Setenv("MCS_EMAIL", "expired@example.com")
	t.Setenv("MCS_PASSWORD", "expired-password")
	t.Setenv("MCS_REGION", "MNAO")

	// Setup: Create expired cached credentials.
	expiredToken := &cache.TokenCache{
		AccessToken:             "expired-token",
		AccessTokenExpirationTs: time.Now().Unix() - 3600, // Expired 1 hour ago.
		EncKey:                  "old-enc-key",
		SignKey:                 "old-sign-key",
	}
	err := cache.Save(expiredToken)
	require.NoError(t, err, "Failed to save expired token: %v")

	// Test: Create API client (should ignore expired cache).
	ctx := testContextWithConfig("")
	client, err := createAPIClient(ctx)
	require.NoError(t, err, "Failed to create API client: %v")

	// Verify: Cached credentials should not be loaded (expired).
	accessToken, _, _, _ := client.GetCredentials()

	assert.NotEqual(t, "expired-token", accessToken, "Expired cache should not be loaded")
}

// TestSaveClientCache_ValidCredentials tests that client credentials are saved to cache.
func TestSaveClientCache_ValidCredentials(t *testing.T) {
	// Setup: Create temporary environment.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Create client with valid credentials.
	client, err := api.NewClient("test@example.com", "password", api.RegionMNAO)
	require.NoError(t, err, "Failed to create client: %v")

	// Set credentials.
	futureTimestamp := time.Now().Unix() + 3600
	client.SetCachedCredentials("test-token", futureTimestamp, "testenckey123456", "testsignkey12345")

	// Test: Save client cache.
	saveClientCache(client)

	// Verify: Load cache and check values.
	loadedCache, err := cache.Load()
	require.NoError(t, err, "Failed to load cache: %v")

	require.NotNil(t, loadedCache, "Expected cache to be saved, got nil")

	assert.Equalf(t, "test-token", loadedCache.AccessToken, "Expected access token 'test-token', got %s", loadedCache.AccessToken)

	assert.Equalf(t, "testenckey123456", loadedCache.EncKey, "Expected enc key 'testenckey123456', got %s", loadedCache.EncKey)

	assert.Equalf(t, "testsignkey12345", loadedCache.SignKey, "Expected sign key 'testsignkey12345', got %s", loadedCache.SignKey)
}

// TestSaveClientCache_EmptyCredentials tests that empty credentials are not saved.
func TestSaveClientCache_EmptyCredentials(t *testing.T) {
	// Setup: Create temporary environment.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Create client without credentials.
	client, err := api.NewClient("test@example.com", "password", api.RegionMNAO)
	require.NoError(t, err, "Failed to create client: %v")

	// Test: Save client cache (should not save empty credentials).
	saveClientCache(client)

	// Verify: Cache should not exist.
	loadedCache, err := cache.Load()
	require.NoError(t, err, "Failed to load cache: %v")

	assert.Nil(t, loadedCache)
}

// TestSaveClientCache_PartialCredentials tests that partial credentials are not saved.
func TestSaveClientCache_PartialCredentials(t *testing.T) {
	// Setup: Create temporary environment.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Create client with partial credentials.
	client, err := api.NewClient("test@example.com", "password", api.RegionMNAO)
	require.NoError(t, err, "Failed to create client: %v")

	// Set only some credentials (missing signKey).
	futureTimestamp := time.Now().Unix() + 3600
	client.SetCachedCredentials("partial-token", futureTimestamp, "partial-enc", "")

	// Test: Save client cache (should not save due to missing field).
	saveClientCache(client)

	// Verify: Cache should not exist (missing signKey).
	loadedCache, err := cache.Load()
	require.NoError(t, err, "Failed to load cache: %v")

	assert.Nil(t, loadedCache)
}

// TestCreateAPIClient_EnvVarOverridesFile tests that env vars override config file.
func TestCreateAPIClient_EnvVarOverridesFile(t *testing.T) {
	// Setup: Create temporary environment.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	configPath := filepath.Join(tmpDir, "test-config.toml")

	// Setup: Create config file.
	configContent := `
email = "file@example.com"
password = "file-password"
region = "MME"
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err, "Failed to create config file: %v")

	// Setup: Set env vars (should override file).
	t.Setenv("MCS_EMAIL", "env@example.com")
	t.Setenv("MCS_PASSWORD", "env-password")
	t.Setenv("MCS_REGION", "MNAO")

	// Test: Create API client.
	ctx := testContextWithConfig(configPath)
	client, err := createAPIClient(ctx)
	require.NoError(t, err, "Failed to create API client: %v")

	// Verify: Client should be created (env values should be used).
	// We can't directly verify the internal values, but we can verify no error.
	assert.NotNil(t, client)
}

// TestCreateAPIClient_MissingCredentials tests error when credentials are missing.
func TestCreateAPIClient_MissingCredentials(t *testing.T) {
	// Setup: Create temporary environment.
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Don't set any credentials.
	t.Setenv("MCS_EMAIL", "")
	t.Setenv("MCS_PASSWORD", "")
	t.Setenv("MCS_REGION", "")

	// Test: Create API client should fail.
	ctx := testContextWithConfig("")
	_, err := createAPIClient(ctx)
	require.Error(t, err, "Expected error with missing credentials, got nil")
}

// TestVehicleInfo_InternalVINType tests that VehicleInfo.InternalVIN uses api.InternalVIN type.
func TestVehicleInfo_InternalVINType(t *testing.T) {
	t.Parallel()
	// This test verifies compile-time type safety for InternalVIN.
	// Create a VehicleInfo with api.InternalVIN type.
	vehicleInfo := VehicleInfo{
		InternalVIN: api.InternalVIN("test-vin-123"),
		VIN:         "JM3XXXXXXXXXX1234",
		Nickname:    "Test Vehicle",
		ModelName:   "CX-90",
		ModelYear:   "2024",
	}

	// Verify that InternalVIN is of type api.InternalVIN.
	var _ = vehicleInfo.InternalVIN

	// Verify that we can convert to string using String() method.
	vinString := vehicleInfo.InternalVIN.String()
	assert.Equalf(t, "test-vin-123", vinString, "Expected VIN string 'test-vin-123', got '%s'", vinString)

	// Verify that we can use it directly as string (implicit conversion).
	assert.Equalf(t, "test-vin-123", string(vehicleInfo.InternalVIN), "Expected VIN string 'test-vin-123', got '%s'", string(vehicleInfo.InternalVIN))
}

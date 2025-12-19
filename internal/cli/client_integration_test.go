package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cv/mcs/internal/api"
	"github.com/cv/mcs/internal/cache"
)

// TestCreateAPIClient_WithValidConfig tests creating an API client with valid config
func TestCreateAPIClient_WithValidConfig(t *testing.T) {
	// Setup: Create temporary environment
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Configure environment variables
	t.Setenv("MCS_EMAIL", "test@example.com")
	t.Setenv("MCS_PASSWORD", "test-password")
	t.Setenv("MCS_REGION", "MNAO")

	// Test: Create API client
	ConfigFile = "" // Use environment variables
	client, err := createAPIClient(context.Background())
	if err != nil {
		t.Fatalf("Failed to create API client: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}
}

// TestCreateAPIClient_WithInvalidRegion tests error handling for invalid region
func TestCreateAPIClient_WithInvalidRegion(t *testing.T) {
	// Setup: Create temporary environment
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Configure with invalid region
	t.Setenv("MCS_EMAIL", "test@example.com")
	t.Setenv("MCS_PASSWORD", "test-password")
	t.Setenv("MCS_REGION", "INVALID")

	// Test: Create API client should fail
	ConfigFile = ""
	_, err := createAPIClient(context.Background())
	if err == nil {
		t.Fatal("Expected error with invalid region, got nil")
	}
}

// TestCreateAPIClient_WithConfigFile tests loading config from file
func TestCreateAPIClient_WithConfigFile(t *testing.T) {
	// Setup: Create temporary environment
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	configPath := filepath.Join(tmpDir, "test-config.toml")

	// Setup: Create config file
	configContent := `
email = "file@example.com"
password = "file-password"
region = "MNAO"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Clear env vars
	t.Setenv("MCS_EMAIL", "")
	t.Setenv("MCS_PASSWORD", "")
	t.Setenv("MCS_REGION", "")

	// Test: Create API client from file
	ConfigFile = configPath
	client, err := createAPIClient(context.Background())
	if err != nil {
		t.Fatalf("Failed to create API client: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}
}

// TestCreateAPIClient_WithCachedCredentials tests using cached credentials
func TestCreateAPIClient_WithCachedCredentials(t *testing.T) {
	// Setup: Create temporary environment
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Configure environment variables
	t.Setenv("MCS_EMAIL", "cached@example.com")
	t.Setenv("MCS_PASSWORD", "cached-password")
	t.Setenv("MCS_REGION", "MNAO")

	// Setup: Create valid cached credentials
	cachedToken := &cache.TokenCache{
		AccessToken:             "cached-token-12345",
		AccessTokenExpirationTs: time.Now().Unix() + 3600,
		EncKey:                  "cached-enc-key",
		SignKey:                 "cached-sign-key",
	}
	if err := cache.Save(cachedToken); err != nil {
		t.Fatalf("Failed to save cached token: %v", err)
	}

	// Test: Create API client (should use cached credentials)
	ConfigFile = ""
	client, err := createAPIClient(context.Background())
	if err != nil {
		t.Fatalf("Failed to create API client: %v", err)
	}

	// Verify: Check that cached credentials were loaded
	accessToken, expirationTs, encKey, signKey := client.GetCredentials()

	if accessToken != "cached-token-12345" {
		t.Errorf("Expected cached access token, got %s", accessToken)
	}
	if expirationTs != cachedToken.AccessTokenExpirationTs {
		t.Errorf("Expected cached expiration ts %d, got %d", cachedToken.AccessTokenExpirationTs, expirationTs)
	}
	if encKey != "cached-enc-key" {
		t.Errorf("Expected cached enc key, got %s", encKey)
	}
	if signKey != "cached-sign-key" {
		t.Errorf("Expected cached sign key, got %s", signKey)
	}
}

// TestCreateAPIClient_WithExpiredCache tests that expired cache is ignored
func TestCreateAPIClient_WithExpiredCache(t *testing.T) {
	// Setup: Create temporary environment
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Configure environment variables
	t.Setenv("MCS_EMAIL", "expired@example.com")
	t.Setenv("MCS_PASSWORD", "expired-password")
	t.Setenv("MCS_REGION", "MNAO")

	// Setup: Create expired cached credentials
	expiredToken := &cache.TokenCache{
		AccessToken:             "expired-token",
		AccessTokenExpirationTs: time.Now().Unix() - 3600, // Expired 1 hour ago
		EncKey:                  "old-enc-key",
		SignKey:                 "old-sign-key",
	}
	if err := cache.Save(expiredToken); err != nil {
		t.Fatalf("Failed to save expired token: %v", err)
	}

	// Test: Create API client (should ignore expired cache)
	ConfigFile = ""
	client, err := createAPIClient(context.Background())
	if err != nil {
		t.Fatalf("Failed to create API client: %v", err)
	}

	// Verify: Cached credentials should not be loaded (expired)
	accessToken, _, _, _ := client.GetCredentials()

	if accessToken == "expired-token" {
		t.Error("Expired cache should not be loaded")
	}
}

// TestSaveClientCache_ValidCredentials tests that client credentials are saved to cache
func TestSaveClientCache_ValidCredentials(t *testing.T) {
	// Setup: Create temporary environment
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Create client with valid credentials
	client, err := api.NewClient("test@example.com", "password", api.RegionMNAO)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Set credentials
	futureTimestamp := time.Now().Unix() + 3600
	client.SetCachedCredentials("test-token", futureTimestamp, "testenckey123456", "testsignkey12345")

	// Test: Save client cache
	saveClientCache(client)

	// Verify: Load cache and check values
	loadedCache, err := cache.Load()
	if err != nil {
		t.Fatalf("Failed to load cache: %v", err)
	}

	if loadedCache == nil {
		t.Fatal("Expected cache to be saved, got nil")
	}

	if loadedCache.AccessToken != "test-token" {
		t.Errorf("Expected access token 'test-token', got %s", loadedCache.AccessToken)
	}

	if loadedCache.EncKey != "testenckey123456" {
		t.Errorf("Expected enc key 'testenckey123456', got %s", loadedCache.EncKey)
	}

	if loadedCache.SignKey != "testsignkey12345" {
		t.Errorf("Expected sign key 'testsignkey12345', got %s", loadedCache.SignKey)
	}
}

// TestSaveClientCache_EmptyCredentials tests that empty credentials are not saved
func TestSaveClientCache_EmptyCredentials(t *testing.T) {
	// Setup: Create temporary environment
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Create client without credentials
	client, err := api.NewClient("test@example.com", "password", api.RegionMNAO)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Test: Save client cache (should not save empty credentials)
	saveClientCache(client)

	// Verify: Cache should not exist
	loadedCache, err := cache.Load()
	if err != nil {
		t.Fatalf("Failed to load cache: %v", err)
	}

	if loadedCache != nil {
		t.Error("Expected no cache to be saved with empty credentials")
	}
}

// TestSaveClientCache_PartialCredentials tests that partial credentials are not saved
func TestSaveClientCache_PartialCredentials(t *testing.T) {
	// Setup: Create temporary environment
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Create client with partial credentials
	client, err := api.NewClient("test@example.com", "password", api.RegionMNAO)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Set only some credentials (missing signKey)
	futureTimestamp := time.Now().Unix() + 3600
	client.SetCachedCredentials("partial-token", futureTimestamp, "partial-enc", "")

	// Test: Save client cache (should not save due to missing field)
	saveClientCache(client)

	// Verify: Cache should not exist (missing signKey)
	loadedCache, err := cache.Load()
	if err != nil {
		t.Fatalf("Failed to load cache: %v", err)
	}

	if loadedCache != nil {
		t.Error("Expected no cache to be saved with partial credentials")
	}
}

// TestCreateAPIClient_EnvVarOverridesFile tests that env vars override config file
func TestCreateAPIClient_EnvVarOverridesFile(t *testing.T) {
	// Setup: Create temporary environment
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	configPath := filepath.Join(tmpDir, "test-config.toml")

	// Setup: Create config file
	configContent := `
email = "file@example.com"
password = "file-password"
region = "MME"
`
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Setup: Set env vars (should override file)
	t.Setenv("MCS_EMAIL", "env@example.com")
	t.Setenv("MCS_PASSWORD", "env-password")
	t.Setenv("MCS_REGION", "MNAO")

	// Test: Create API client
	ConfigFile = configPath
	client, err := createAPIClient(context.Background())
	if err != nil {
		t.Fatalf("Failed to create API client: %v", err)
	}

	// Verify: Client should be created (env values should be used)
	// We can't directly verify the internal values, but we can verify no error
	if client == nil {
		t.Error("Expected client to be created with env var values")
	}
}

// TestCreateAPIClient_MissingCredentials tests error when credentials are missing
func TestCreateAPIClient_MissingCredentials(t *testing.T) {
	// Setup: Create temporary environment
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Setup: Don't set any credentials
	t.Setenv("MCS_EMAIL", "")
	t.Setenv("MCS_PASSWORD", "")
	t.Setenv("MCS_REGION", "")

	// Test: Create API client should fail
	ConfigFile = ""
	_, err := createAPIClient(context.Background())
	if err == nil {
		t.Fatal("Expected error with missing credentials, got nil")
	}
}

// TestVehicleInfo_InternalVINType tests that VehicleInfo.InternalVIN uses api.InternalVIN type
func TestVehicleInfo_InternalVINType(t *testing.T) {
	// This test verifies compile-time type safety for InternalVIN
	// Create a VehicleInfo with api.InternalVIN type
	vehicleInfo := VehicleInfo{
		InternalVIN: api.InternalVIN("test-vin-123"),
		VIN:         "JM3XXXXXXXXXX1234",
		Nickname:    "Test Vehicle",
		ModelName:   "CX-90",
		ModelYear:   "2024",
	}

	// Verify that InternalVIN is of type api.InternalVIN
	var _ = vehicleInfo.InternalVIN

	// Verify that we can convert to string using String() method
	vinString := vehicleInfo.InternalVIN.String()
	if vinString != "test-vin-123" {
		t.Errorf("Expected VIN string 'test-vin-123', got '%s'", vinString)
	}

	// Verify that we can use it directly as string (implicit conversion)
	if string(vehicleInfo.InternalVIN) != "test-vin-123" {
		t.Errorf("Expected VIN string 'test-vin-123', got '%s'", string(vehicleInfo.InternalVIN))
	}
}

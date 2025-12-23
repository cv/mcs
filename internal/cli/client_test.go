package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cv/mcs/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAPIClientSetup is a helper that sets up environment for API client creation tests
func mockAPIClientSetup(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("MCS_EMAIL", "test@example.com")
	t.Setenv("MCS_PASSWORD", "test-password")
	t.Setenv("MCS_REGION", "MNAO")
	ConfigFile = ""
	return tmpDir
}

// TestSetupVehicleClient_Success tests successful vehicle client setup
func TestSetupVehicleClient_Success(t *testing.T) {
	// This is an integration test that requires real API interaction
	// Skip if we don't have valid credentials
	t.Skip("Requires real API credentials - integration test")

	mockAPIClientSetup(t)

	ctx := context.Background()
	client, vehicleInfo, err := setupVehicleClient(ctx)

	require.NoError(t, err, "Expected successful setup, got error: %v")

	assert.NotNil(t, client)

	// Verify VehicleInfo fields are populated
	assert.NotEmpty(t, vehicleInfo.InternalVIN, "Expected InternalVIN to be set")
}

// TestSetupVehicleClient_ConfigError tests error handling when config is invalid
func TestSetupVehicleClient_ConfigError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Set invalid region
	t.Setenv("MCS_EMAIL", "test@example.com")
	t.Setenv("MCS_PASSWORD", "test-password")
	t.Setenv("MCS_REGION", "INVALID_REGION")
	ConfigFile = ""

	ctx := context.Background()
	_, _, err := setupVehicleClient(ctx)

	require.Error(t, err, "Expected error with invalid config, got nil")
}

// TestSetupVehicleClient_MissingConfig tests error when config file doesn't exist
func TestSetupVehicleClient_MissingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	// Clear env vars
	t.Setenv("MCS_EMAIL", "")
	t.Setenv("MCS_PASSWORD", "")
	t.Setenv("MCS_REGION", "")

	// Point to non-existent config file
	ConfigFile = filepath.Join(tmpDir, "nonexistent.toml")

	ctx := context.Background()
	_, _, err := setupVehicleClient(ctx)

	require.Error(t, err, "Expected error with missing config, got nil")
}

// TestSetupVehicleClient_ContextCancellation tests context cancellation handling
func TestSetupVehicleClient_ContextCancellation(t *testing.T) {
	mockAPIClientSetup(t)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, _, err := setupVehicleClient(ctx)

	// Should return an error (either context cancelled or connection error)
	require.Error(t, err, "Expected error with cancelled context, got nil")
}

// TestWithVehicleClient_CallbackExecuted tests that callback is executed with client
func TestWithVehicleClient_CallbackExecuted(t *testing.T) {
	t.Skip("Requires real API credentials - integration test")

	mockAPIClientSetup(t)
	ctx := context.Background()

	callbackExecuted := false
	var receivedClient *api.Client
	var receivedVIN api.InternalVIN

	err := withVehicleClient(ctx, func(ctx context.Context, client *api.Client, vin api.InternalVIN) error {
		callbackExecuted = true
		receivedClient = client
		receivedVIN = vin
		return nil
	})

	require.NoError(t, err, "Expected successful execution, got error: %v")

	assert.True(t, callbackExecuted)

	assert.NotNil(t, receivedClient)

	assert.NotEmpty(t, receivedVIN, "Expected VIN to be passed to callback")
}

// TestWithVehicleClient_CallbackError tests that callback errors are propagated
func TestWithVehicleClient_CallbackError(t *testing.T) {
	t.Skip("Requires real API credentials - integration test")

	mockAPIClientSetup(t)
	ctx := context.Background()

	expectedErr := errors.New("callback error")

	err := withVehicleClient(ctx, func(ctx context.Context, client *api.Client, vin api.InternalVIN) error {
		return expectedErr
	})

	assert.Equalf(t, expectedErr, err, "Expected error to be propagated, got: %v", err)
}

// TestWithVehicleClient_SetupError tests that setup errors are propagated
func TestWithVehicleClient_SetupError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("MCS_EMAIL", "")
	t.Setenv("MCS_PASSWORD", "")
	t.Setenv("MCS_REGION", "")
	ConfigFile = ""

	ctx := context.Background()

	callbackExecuted := false

	err := withVehicleClient(ctx, func(ctx context.Context, client *api.Client, vin api.InternalVIN) error {
		callbackExecuted = true
		return nil
	})

	require.Error(t, err, "Expected setup error to be propagated")

	assert.False(t, callbackExecuted)
}

// TestWithVehicleClientEx_CallbackExecuted tests extended callback execution
func TestWithVehicleClientEx_CallbackExecuted(t *testing.T) {
	t.Skip("Requires real API credentials - integration test")

	mockAPIClientSetup(t)
	ctx := context.Background()

	callbackExecuted := false
	var receivedClient *api.Client
	var receivedInfo VehicleInfo

	err := withVehicleClientEx(ctx, func(ctx context.Context, client *api.Client, info VehicleInfo) error {
		callbackExecuted = true
		receivedClient = client
		receivedInfo = info
		return nil
	})

	require.NoError(t, err, "Expected successful execution, got error: %v")

	assert.True(t, callbackExecuted)

	assert.NotNil(t, receivedClient)

	// Verify full VehicleInfo is passed
	assert.NotEmpty(t, receivedInfo.InternalVIN, "Expected InternalVIN to be set")
}

// TestWithVehicleClientEx_CallbackError tests error propagation
func TestWithVehicleClientEx_CallbackError(t *testing.T) {
	t.Skip("Requires real API credentials - integration test")

	mockAPIClientSetup(t)
	ctx := context.Background()

	expectedErr := errors.New("extended callback error")

	err := withVehicleClientEx(ctx, func(ctx context.Context, client *api.Client, info VehicleInfo) error {
		return expectedErr
	})

	assert.Equalf(t, expectedErr, err, "Expected error to be propagated, got: %v", err)
}

// TestWithVehicleClientEx_SetupError tests setup error propagation
func TestWithVehicleClientEx_SetupError(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	t.Setenv("MCS_EMAIL", "")
	t.Setenv("MCS_PASSWORD", "")
	t.Setenv("MCS_REGION", "")
	ConfigFile = ""

	ctx := context.Background()

	callbackExecuted := false

	err := withVehicleClientEx(ctx, func(ctx context.Context, client *api.Client, info VehicleInfo) error {
		callbackExecuted = true
		return nil
	})

	require.Error(t, err, "Expected setup error to be propagated")

	assert.False(t, callbackExecuted)
}

// TestVehicleInfo_StructFields tests VehicleInfo struct field types
func TestVehicleInfo_StructFields(t *testing.T) {
	// Test that VehicleInfo struct has correct field types
	info := VehicleInfo{
		InternalVIN: api.InternalVIN("test123"),
		VIN:         "JM3KKEHC1R0123456",
		Nickname:    "Test Car",
		ModelName:   "CX-90",
		ModelYear:   "2024",
	}

	// Verify InternalVIN is api.InternalVIN type
	var _ = info.InternalVIN

	// Verify string conversion works
	assert.Equalf(t, "test123", string(info.InternalVIN), "Expected InternalVIN to be 'test123', got '%s'", string(info.InternalVIN))

	// Verify String() method works
	assert.Equalf(t, "test123", info.InternalVIN.String(), "Expected InternalVIN.String() to be 'test123', got '%s'", info.InternalVIN.String())

	// Verify other fields
	assert.Equalf(t, "JM3KKEHC1R0123456", info.VIN, "Expected VIN to be 'JM3KKEHC1R0123456', got '%s'", info.VIN)

	assert.Equalf(t, "Test Car", info.Nickname, "Expected Nickname to be 'Test Car', got '%s'", info.Nickname)

	assert.Equalf(t, "CX-90", info.ModelName, "Expected ModelName to be 'CX-90', got '%s'", info.ModelName)

	assert.Equalf(t, "2024", info.ModelYear, "Expected ModelYear to be '2024', got '%s'", info.ModelYear)
}

// TestSetupVehicleClient_ConfigFromFile tests config loading from file
func TestSetupVehicleClient_ConfigFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	configPath := filepath.Join(tmpDir, "test-config.toml")

	// Create config file
	configContent := `
email = "file@example.com"
password = "file-password"
region = "MNAO"
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err, "Failed to create config file: %v")

	// Clear env vars
	t.Setenv("MCS_EMAIL", "")
	t.Setenv("MCS_PASSWORD", "")
	t.Setenv("MCS_REGION", "")

	ConfigFile = configPath

	// This would fail without real API, but should at least get past config loading
	ctx := context.Background()
	_, _, err = setupVehicleClient(ctx)

	// We expect an API error, not a config error
	// This verifies config was loaded successfully
	if err != nil {
		// Check that it's not a config validation error
		// (it will be an API connection error instead)
		if err.Error() == "invalid config: email is required" ||
			err.Error() == "invalid config: password is required" ||
			err.Error() == "invalid config: region is required" {
			t.Fatalf("Config should have been loaded from file, got config error: %v", err)
		}
		// Other errors are expected (API connection, etc.)
	}
}

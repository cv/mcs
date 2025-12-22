package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/cv/mcs/internal/api"
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

	if err != nil {
		t.Fatalf("Expected successful setup, got error: %v", err)
	}

	if client == nil {
		t.Error("Expected client to be created, got nil")
	}

	// Verify VehicleInfo fields are populated
	if vehicleInfo.InternalVIN == "" {
		t.Error("Expected InternalVIN to be set")
	}
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

	if err == nil {
		t.Fatal("Expected error with invalid config, got nil")
	}
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

	if err == nil {
		t.Fatal("Expected error with missing config, got nil")
	}
}

// TestSetupVehicleClient_ContextCancellation tests context cancellation handling
func TestSetupVehicleClient_ContextCancellation(t *testing.T) {
	mockAPIClientSetup(t)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, _, err := setupVehicleClient(ctx)

	// Should return an error (either context cancelled or connection error)
	if err == nil {
		t.Error("Expected error with cancelled context, got nil")
	}
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

	if err != nil {
		t.Fatalf("Expected successful execution, got error: %v", err)
	}

	if !callbackExecuted {
		t.Error("Expected callback to be executed")
	}

	if receivedClient == nil {
		t.Error("Expected client to be passed to callback")
	}

	if receivedVIN == "" {
		t.Error("Expected VIN to be passed to callback")
	}
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

	if err != expectedErr {
		t.Errorf("Expected error to be propagated, got: %v", err)
	}
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

	if err == nil {
		t.Fatal("Expected setup error to be propagated")
	}

	if callbackExecuted {
		t.Error("Expected callback not to be executed when setup fails")
	}
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

	if err != nil {
		t.Fatalf("Expected successful execution, got error: %v", err)
	}

	if !callbackExecuted {
		t.Error("Expected callback to be executed")
	}

	if receivedClient == nil {
		t.Error("Expected client to be passed to callback")
	}

	// Verify full VehicleInfo is passed
	if receivedInfo.InternalVIN == "" {
		t.Error("Expected InternalVIN to be set")
	}
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

	if err != expectedErr {
		t.Errorf("Expected error to be propagated, got: %v", err)
	}
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

	if err == nil {
		t.Fatal("Expected setup error to be propagated")
	}

	if callbackExecuted {
		t.Error("Expected callback not to be executed when setup fails")
	}
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
	if string(info.InternalVIN) != "test123" {
		t.Errorf("Expected InternalVIN to be 'test123', got '%s'", string(info.InternalVIN))
	}

	// Verify String() method works
	if info.InternalVIN.String() != "test123" {
		t.Errorf("Expected InternalVIN.String() to be 'test123', got '%s'", info.InternalVIN.String())
	}

	// Verify other fields
	if info.VIN != "JM3KKEHC1R0123456" {
		t.Errorf("Expected VIN to be 'JM3KKEHC1R0123456', got '%s'", info.VIN)
	}

	if info.Nickname != "Test Car" {
		t.Errorf("Expected Nickname to be 'Test Car', got '%s'", info.Nickname)
	}

	if info.ModelName != "CX-90" {
		t.Errorf("Expected ModelName to be 'CX-90', got '%s'", info.ModelName)
	}

	if info.ModelYear != "2024" {
		t.Errorf("Expected ModelYear to be '2024', got '%s'", info.ModelYear)
	}
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
	if err := os.WriteFile(configPath, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Clear env vars
	t.Setenv("MCS_EMAIL", "")
	t.Setenv("MCS_PASSWORD", "")
	t.Setenv("MCS_REGION", "")

	ConfigFile = configPath

	// This would fail without real API, but should at least get past config loading
	ctx := context.Background()
	_, _, err := setupVehicleClient(ctx)

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

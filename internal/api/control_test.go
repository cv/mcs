package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestControlEndpoints tests all simple control endpoints using table-driven approach
func TestControlEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		method   func(ctx context.Context, client *Client, internalVIN string) error
	}{
		{
			name:     "DoorLock",
			endpoint: EndpointDoorLock,
			method:   func(ctx context.Context, client *Client, vin string) error { return client.DoorLock(ctx, vin) },
		},
		{
			name:     "DoorUnlock",
			endpoint: EndpointDoorUnlock,
			method:   func(ctx context.Context, client *Client, vin string) error { return client.DoorUnlock(ctx, vin) },
		},
		{
			name:     "LightsOn",
			endpoint: EndpointLightOn,
			method:   func(ctx context.Context, client *Client, vin string) error { return client.LightsOn(ctx, vin) },
		},
		{
			name:     "LightsOff",
			endpoint: EndpointLightOff,
			method:   func(ctx context.Context, client *Client, vin string) error { return client.LightsOff(ctx, vin) },
		},
		{
			name:     "EngineStart",
			endpoint: EndpointEngineStart,
			method:   func(ctx context.Context, client *Client, vin string) error { return client.EngineStart(ctx, vin) },
		},
		{
			name:     "EngineStop",
			endpoint: EndpointEngineStop,
			method:   func(ctx context.Context, client *Client, vin string) error { return client.EngineStop(ctx, vin) },
		},
		{
			name:     "ChargeStart",
			endpoint: EndpointChargeStart,
			method:   func(ctx context.Context, client *Client, vin string) error { return client.ChargeStart(ctx, vin) },
		},
		{
			name:     "ChargeStop",
			endpoint: EndpointChargeStop,
			method:   func(ctx context.Context, client *Client, vin string) error { return client.ChargeStop(ctx, vin) },
		},
		{
			name:     "HVACOn",
			endpoint: EndpointHVACOn,
			method:   func(ctx context.Context, client *Client, vin string) error { return client.HVACOn(ctx, vin) },
		},
		{
			name:     "HVACOff",
			endpoint: EndpointHVACOff,
			method:   func(ctx context.Context, client *Client, vin string) error { return client.HVACOff(ctx, vin) },
		},
		{
			name:     "RefreshVehicleStatus",
			endpoint: EndpointRefreshVehicleStatus,
			method: func(ctx context.Context, client *Client, vin string) error {
				return client.RefreshVehicleStatus(ctx, vin)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createControlTestServer(t, "/"+tt.endpoint)
			defer server.Close()

			client := createTestClient(t, server.URL)

			err := tt.method(context.Background(), client, "INTERNAL123")
			if err != nil {
				t.Fatalf("%s failed: %v", tt.name, err)
			}
		})
	}
}

// TestSetHVACSetting tests setting HVAC settings
func TestSetHVACSetting(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/updateHVACSetting/v4")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.SetHVACSetting(context.Background(), "INTERNAL123", 22.0, Celsius, true, false)
	require.NoError(t, err, "SetHVACSetting failed: %v")
}

// TestSetHVACSetting_Fahrenheit tests setting HVAC with Fahrenheit
func TestSetHVACSetting_Fahrenheit(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/updateHVACSetting/v4")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.SetHVACSetting(context.Background(), "INTERNAL123", 72.0, Fahrenheit, false, true)
	require.NoError(t, err, "SetHVACSetting failed: %v")
}

// TestControlError tests error handling for control endpoints
func TestControlError(t *testing.T) {
	server := createErrorServer(t, "500E00", "Internal error")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.DoorLock(context.Background(), "INTERNAL123")
	require.Error(t, err, "Expected error, got nil")

	expectedError := "failed to lock doors: result code 500E00"
	assert.EqualValuesf(t, expectedError, err.Error(), "Expected error '%s', got '%s'")
}

// TestBoolToInt tests the boolToInt helper function
func TestBoolToInt(t *testing.T) {
	tests := []struct {
		input bool
		want  int
	}{
		{true, 1},
		{false, 0},
	}

	for _, tt := range tests {
		if got := boolToInt(tt.input); got != tt.want {
			t.Errorf("boolToInt(%v) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

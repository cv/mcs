package api

import (
	"context"
	"testing"
)

// TestDoorLock tests locking doors
func TestDoorLock(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/doorLock/v4")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.DoorLock(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("DoorLock failed: %v", err)
	}
}

// TestDoorUnlock tests unlocking doors
func TestDoorUnlock(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/doorUnlock/v4")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.DoorUnlock(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("DoorUnlock failed: %v", err)
	}
}

// TestLightsOn tests turning lights on
func TestLightsOn(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/lightOn/v4")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.LightsOn(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("LightsOn failed: %v", err)
	}
}

// TestLightsOff tests turning lights off
func TestLightsOff(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/lightOff/v4")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.LightsOff(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("LightsOff failed: %v", err)
	}
}

// TestEngineStart tests starting the engine
func TestEngineStart(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/engineStart/v4")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.EngineStart(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("EngineStart failed: %v", err)
	}
}

// TestEngineStop tests stopping the engine
func TestEngineStop(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/engineStop/v4")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.EngineStop(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("EngineStop failed: %v", err)
	}
}

// TestChargeStart tests starting charging
func TestChargeStart(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/chargeStart/v4")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.ChargeStart(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("ChargeStart failed: %v", err)
	}
}

// TestChargeStop tests stopping charging
func TestChargeStop(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/chargeStop/v4")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.ChargeStop(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("ChargeStop failed: %v", err)
	}
}

// TestHVACOn tests turning HVAC on
func TestHVACOn(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/hvacOn/v4")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.HVACOn(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("HVACOn failed: %v", err)
	}
}

// TestHVACOff tests turning HVAC off
func TestHVACOff(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/hvacOff/v4")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.HVACOff(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("HVACOff failed: %v", err)
	}
}

// TestRefreshVehicleStatus tests refreshing vehicle status
func TestRefreshVehicleStatus(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/activeRealTimeVehicleStatus/v4")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.RefreshVehicleStatus(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("RefreshVehicleStatus failed: %v", err)
	}
}

// TestSetHVACSetting tests setting HVAC settings
func TestSetHVACSetting(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/updateHVACSetting/v4")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.SetHVACSetting(context.Background(), "INTERNAL123", 22.0, Celsius, true, false)
	if err != nil {
		t.Fatalf("SetHVACSetting failed: %v", err)
	}
}

// TestSetHVACSetting_Fahrenheit tests setting HVAC with Fahrenheit
func TestSetHVACSetting_Fahrenheit(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/updateHVACSetting/v4")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.SetHVACSetting(context.Background(), "INTERNAL123", 72.0, Fahrenheit, false, true)
	if err != nil {
		t.Fatalf("SetHVACSetting failed: %v", err)
	}
}

// TestControlError tests error handling for control endpoints
func TestControlError(t *testing.T) {
	server := createErrorServer(t, "500E00", "Internal error")
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.DoorLock(context.Background(), "INTERNAL123")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedError := "failed to lock doors: result code 500E00"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
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

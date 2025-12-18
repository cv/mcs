package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// createControlTestServer creates a test server for control endpoints
// Uses shared test helpers from test_helpers_test.go
func createControlTestServer(t *testing.T, expectedPath string, expectedBody map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if expectedBody != nil {
			body, _ := io.ReadAll(r.Body)
			if len(body) == 0 {
				t.Error("Expected non-empty body")
			}
		}

		testResponse := map[string]interface{}{
			"resultCode": "200S00",
			"message":    "Success",
		}

		responseJSON, _ := json.Marshal(testResponse)
		encrypted, _ := EncryptAES128CBC(responseJSON, testEncKey, IV)

		response := map[string]interface{}{
			"state":   "S",
			"payload": encrypted,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
}

// Note: createTestClient is defined in test_helpers_test.go

// TestDoorLock tests locking doors
func TestDoorLock(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/doorLock/v4", map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    "INTERNAL123",
	})
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.DoorLock(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("DoorLock failed: %v", err)
	}
}

// TestDoorUnlock tests unlocking doors
func TestDoorUnlock(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/doorUnlock/v4", map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    "INTERNAL123",
	})
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.DoorUnlock(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("DoorUnlock failed: %v", err)
	}
}

// TestLightsOn tests turning lights on
func TestLightsOn(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/lightOn/v4", map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    "INTERNAL123",
	})
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.LightsOn(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("LightsOn failed: %v", err)
	}
}

// TestLightsOff tests turning lights off
func TestLightsOff(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/lightOff/v4", map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    "INTERNAL123",
	})
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.LightsOff(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("LightsOff failed: %v", err)
	}
}

// TestEngineStart tests starting the engine
func TestEngineStart(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/engineStart/v4", map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    "INTERNAL123",
	})
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.EngineStart(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("EngineStart failed: %v", err)
	}
}

// TestEngineStop tests stopping the engine
func TestEngineStop(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/engineStop/v4", map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    "INTERNAL123",
	})
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.EngineStop(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("EngineStop failed: %v", err)
	}
}

// TestChargeStart tests starting charging
func TestChargeStart(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/chargeStart/v4", map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    "INTERNAL123",
	})
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.ChargeStart(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("ChargeStart failed: %v", err)
	}
}

// TestChargeStop tests stopping charging
func TestChargeStop(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/chargeStop/v4", map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    "INTERNAL123",
	})
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.ChargeStop(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("ChargeStop failed: %v", err)
	}
}

// TestHVACOn tests turning HVAC on
func TestHVACOn(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/hvacOn/v4", map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    "INTERNAL123",
	})
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.HVACOn(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("HVACOn failed: %v", err)
	}
}

// TestHVACOff tests turning HVAC off
func TestHVACOff(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/hvacOff/v4", map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    "INTERNAL123",
	})
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.HVACOff(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("HVACOff failed: %v", err)
	}
}

// TestRefreshVehicleStatus tests refreshing vehicle status
func TestRefreshVehicleStatus(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/activeRealTimeVehicleStatus/v4", map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    "INTERNAL123",
	})
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.RefreshVehicleStatus(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("RefreshVehicleStatus failed: %v", err)
	}
}

// TestSetHVACSetting tests setting HVAC settings
func TestSetHVACSetting(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/updateHVACSetting/v4", map[string]interface{}{
		"internaluserid":  "__INTERNAL_ID__",
		"internalvin":     "INTERNAL123",
		"Temperature":     22.0,
		"TemperatureType": 1,
		"FrontDefroster":  1,
		"RearDefogger":    0,
	})
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.SetHVACSetting(context.Background(), "INTERNAL123", 22.0, "c", true, false)
	if err != nil {
		t.Fatalf("SetHVACSetting failed: %v", err)
	}
}

// TestSetHVACSetting_Fahrenheit tests setting HVAC with Fahrenheit
func TestSetHVACSetting_Fahrenheit(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/updateHVACSetting/v4", map[string]interface{}{
		"internaluserid":  "__INTERNAL_ID__",
		"internalvin":     "INTERNAL123",
		"Temperature":     72.0,
		"TemperatureType": 2,
		"FrontDefroster":  0,
		"RearDefogger":    1,
	})
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.SetHVACSetting(context.Background(), "INTERNAL123", 72.0, "f", false, true)
	if err != nil {
		t.Fatalf("SetHVACSetting failed: %v", err)
	}
}

// TestSetHVACSetting_InvalidUnit tests invalid temperature unit
func TestSetHVACSetting_InvalidUnit(t *testing.T) {
	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.SetHVACSetting(context.Background(), "INTERNAL123", 22.0, "k", false, false)
	if err == nil {
		t.Fatal("Expected error for invalid unit, got nil")
	}

	expectedError := "invalid temperature unit: k (must be 'c' or 'f')"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// TestControlError tests error handling for control endpoints
func TestControlError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testResponse := map[string]interface{}{
			"resultCode": "500E00",
			"message":    "Internal error",
		}

		responseJSON, _ := json.Marshal(testResponse)
		encrypted, _ := EncryptAES128CBC(responseJSON, "testenckey123456", IV)

		response := map[string]interface{}{
			"state":   "S",
			"payload": encrypted,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
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

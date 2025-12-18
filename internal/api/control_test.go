package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Helper function to create test server for control endpoints
func createControlTestServer(t *testing.T, expectedPath string, expectedBody map[string]interface{}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify endpoint
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify method
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Verify body if provided
		if expectedBody != nil {
			body, _ := io.ReadAll(r.Body)
			// Body is encrypted, so we just check it's not empty
			if len(body) == 0 {
				t.Error("Expected non-empty body")
			}
		}

		// Return success response
		testResponse := map[string]interface{}{
			"resultCode": "200S00",
			"message":    "Success",
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
}

// Helper function to create test client
func createTestClient(t *testing.T, serverURL string) *Client {
	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.baseURL = serverURL + "/"
	client.encKey = "testenckey123456"
	client.signKey = "testsignkey12345"
	client.accessToken = "test-token"
	client.accessTokenExpirationTs = 9999999999
	return client
}

// TestDoorLock tests locking doors
func TestDoorLock(t *testing.T) {
	server := createControlTestServer(t, "/remoteServices/doorLock/v4", map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    "INTERNAL123",
	})
	defer server.Close()

	client := createTestClient(t, server.URL)

	err := client.DoorLock("INTERNAL123")
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

	err := client.DoorUnlock("INTERNAL123")
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

	err := client.LightsOn("INTERNAL123")
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

	err := client.LightsOff("INTERNAL123")
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

	err := client.EngineStart("INTERNAL123")
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

	err := client.EngineStop("INTERNAL123")
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

	err := client.ChargeStart("INTERNAL123")
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

	err := client.ChargeStop("INTERNAL123")
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

	err := client.HVACOn("INTERNAL123")
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

	err := client.HVACOff("INTERNAL123")
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

	err := client.RefreshVehicleStatus("INTERNAL123")
	if err != nil {
		t.Fatalf("RefreshVehicleStatus failed: %v", err)
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

	err := client.DoorLock("INTERNAL123")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedError := "Failed to lock doors: result code 500E00"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

package api

import (
	"testing"
)

// The current API answers a refresh with MSG200001; that is not a failure.
func TestRefreshVehicleStatusAccepted(t *testing.T) {
	t.Parallel()
	for _, code := range []string{ResultCodeRefreshAccepted, "200S00"} {
		server := createTestServer(t, map[string]any{"resultCode": code, "message": "ok"},
			WithPath("/"+EndpointRefreshVehicleStatus), WithMethod("POST"), WithBodyValidation())
		client := createTestClient(t, server.URL)
		if err := client.RefreshVehicleStatus(t.Context(), "vin"); err != nil {
			t.Fatalf("%s: unexpected error %v", code, err)
		}
		server.Close()
	}
}

// Other control commands still require the success code.
func TestDoorLockStillRequiresSuccess(t *testing.T) {
	t.Parallel()
	server := createTestServer(t, map[string]any{"resultCode": ResultCodeRefreshAccepted, "message": "ok"},
		WithPath("/"+EndpointDoorLock), WithMethod("POST"), WithBodyValidation())
	defer server.Close()
	client := createTestClient(t, server.URL)
	if err := client.DoorLock(t.Context(), "vin"); err == nil {
		t.Fatal("expected an error for a non-success result code")
	}
}

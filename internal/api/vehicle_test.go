package api

import (
	"context"
	"testing"
)

// TestUnmarshalResponse tests the generic unmarshalResponse helper
func TestUnmarshalResponse(t *testing.T) {
	t.Run("successful unmarshal", func(t *testing.T) {
		response := map[string]interface{}{
			"resultCode": "200S00",
			"message":    "Success",
			"count":      float64(42),
		}

		type TestResponse struct {
			ResultCode string  `json:"resultCode"`
			Message    string  `json:"message"`
			Count      float64 `json:"count"`
		}

		result, err := unmarshalResponse[TestResponse](response)
		if err != nil {
			t.Fatalf("unmarshalResponse failed: %v", err)
		}

		if result.ResultCode != "200S00" {
			t.Errorf("Expected resultCode 200S00, got %v", result.ResultCode)
		}

		if result.Message != "Success" {
			t.Errorf("Expected message 'Success', got %v", result.Message)
		}

		if result.Count != 42 {
			t.Errorf("Expected count 42, got %v", result.Count)
		}
	})

	t.Run("unmarshal with nested structures", func(t *testing.T) {
		response := map[string]interface{}{
			"resultCode": "200S00",
			"data": map[string]interface{}{
				"id":   "test123",
				"name": "Test Name",
			},
		}

		type NestedData struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}

		type TestResponse struct {
			ResultCode string     `json:"resultCode"`
			Data       NestedData `json:"data"`
		}

		result, err := unmarshalResponse[TestResponse](response)
		if err != nil {
			t.Fatalf("unmarshalResponse failed: %v", err)
		}

		if result.ResultCode != "200S00" {
			t.Errorf("Expected resultCode 200S00, got %v", result.ResultCode)
		}

		if result.Data.ID != "test123" {
			t.Errorf("Expected data.id 'test123', got %v", result.Data.ID)
		}

		if result.Data.Name != "Test Name" {
			t.Errorf("Expected data.name 'Test Name', got %v", result.Data.Name)
		}
	})

	t.Run("unmarshal with array", func(t *testing.T) {
		response := map[string]interface{}{
			"resultCode": "200S00",
			"items": []interface{}{
				map[string]interface{}{"value": "item1"},
				map[string]interface{}{"value": "item2"},
			},
		}

		type Item struct {
			Value string `json:"value"`
		}

		type TestResponse struct {
			ResultCode string `json:"resultCode"`
			Items      []Item `json:"items"`
		}

		result, err := unmarshalResponse[TestResponse](response)
		if err != nil {
			t.Fatalf("unmarshalResponse failed: %v", err)
		}

		if result.ResultCode != "200S00" {
			t.Errorf("Expected resultCode 200S00, got %v", result.ResultCode)
		}

		if len(result.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(result.Items))
		}

		if result.Items[0].Value != "item1" {
			t.Errorf("Expected items[0].value 'item1', got %v", result.Items[0].Value)
		}

		if result.Items[1].Value != "item2" {
			t.Errorf("Expected items[1].value 'item2', got %v", result.Items[1].Value)
		}
	})
}

// TestGetVecBaseInfos tests getting vehicle base information
func TestGetVecBaseInfos(t *testing.T) {
	responseData := map[string]interface{}{
		"resultCode": "200S00",
		"vecBaseInfos": []map[string]interface{}{
			{
				"vin": "TEST123456789",
				"Vehicle": map[string]interface{}{
					"CvInformation": map[string]interface{}{
						"internalVin": "INTERNAL123",
					},
				},
				"econnectType": 1,
			},
		},
		"vehicleFlags": []map[string]interface{}{
			{"vinRegistStatus": 3},
		},
	}

	server := createSuccessServer(t, "/remoteServices/getVecBaseInfos/v4", responseData)
	defer server.Close()

	client := createTestClient(t, server.URL)

	result, err := client.GetVecBaseInfos(context.Background())
	if err != nil {
		t.Fatalf("GetVecBaseInfos failed: %v", err)
	}

	if result.ResultCode != "200S00" {
		t.Errorf("Expected resultCode 200S00, got %v", result.ResultCode)
	}

	if len(result.VecBaseInfos) != 1 {
		t.Errorf("Expected 1 vehicle, got %d", len(result.VecBaseInfos))
	}
}

// TestGetVehicleStatus tests getting vehicle status
func TestGetVehicleStatus(t *testing.T) {
	responseData := map[string]interface{}{
		"resultCode": "200S00",
		"alertInfos": []map[string]interface{}{
			{
				"OccurrenceDate": "20231201120000",
				"Door": map[string]interface{}{
					"DrStatDrv": 0, "DrStatPsngr": 0, "DrStatRl": 0, "DrStatRr": 0,
					"DrStatTrnkLg": 0, "DrStatHood": 0, "FuelLidOpenStatus": 0,
					"LockLinkSwDrv": 0, "LockLinkSwPsngr": 0, "LockLinkSwRl": 0, "LockLinkSwRr": 0,
				},
				"Pw":         map[string]interface{}{"PwPosDrv": 0, "PwPosPsngr": 0, "PwPosRl": 0, "PwPosRr": 0},
				"HazardLamp": map[string]interface{}{"HazardSw": 0},
			},
		},
		"remoteInfos": []map[string]interface{}{
			{
				"PositionInfo": map[string]interface{}{
					"Latitude": 37.7749, "LatitudeFlag": 0,
					"Longitude": 122.4194, "LongitudeFlag": 1,
					"AcquisitionDatetime": "20231201120000",
				},
				"ResidualFuel":     map[string]interface{}{"FuelSegementDActl": 75.5, "RemDrvDistDActlKm": 350.2},
				"DriveInformation": map[string]interface{}{"OdoDispValue": 12345.6},
				"TPMSInformation":  map[string]interface{}{"FLTPrsDispPsi": 32.5, "FRTPrsDispPsi": 32.0, "RLTPrsDispPsi": 31.5, "RRTPrsDispPsi": 31.8},
			},
		},
	}

	server := createSuccessServer(t, "/remoteServices/getVehicleStatus/v4", responseData)
	defer server.Close()

	client := createTestClient(t, server.URL)

	result, err := client.GetVehicleStatus(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("GetVehicleStatus failed: %v", err)
	}

	if result.ResultCode != "200S00" {
		t.Errorf("Expected resultCode 200S00, got %v", result.ResultCode)
	}

	if len(result.AlertInfos) != 1 {
		t.Errorf("Expected 1 alert info, got %d", len(result.AlertInfos))
	}

	if len(result.RemoteInfos) != 1 {
		t.Errorf("Expected 1 remote info, got %d", len(result.RemoteInfos))
	}
}

// TestGetVehicleStatus_Error tests error handling
func TestGetVehicleStatus_Error(t *testing.T) {
	server := createErrorServer(t, "500E00", "Internal error")
	defer server.Close()

	client := createTestClient(t, server.URL)

	_, err := client.GetVehicleStatus(context.Background(), "INTERNAL123")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedError := "failed to get vehicle status: result code 500E00"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// TestGetEVVehicleStatus tests getting EV vehicle status
func TestGetEVVehicleStatus(t *testing.T) {
	responseData := map[string]interface{}{
		"resultCode": "200S00",
		"resultData": []map[string]interface{}{
			{
				"OccurrenceDate": "20231201120000",
				"PlusBInformation": map[string]interface{}{
					"VehicleInfo": map[string]interface{}{
						"ChargeInfo": map[string]interface{}{
							"SmaphSOC": 85, "SmaphRemDrvDistKm": 245.5,
							"ChargerConnectorFitting": 1, "ChargeStatusSub": 6,
							"MaxChargeMinuteAC": 180, "MaxChargeMinuteQBC": 45,
							"CstmzStatBatHeatAutoSW": 1, "BatteryHeaterON": 0,
						},
						"RemoteHvacInfo": map[string]interface{}{
							"HVAC": 1, "FrontDefroster": 0, "RearDefogger": 0,
							"InteriorTemp": 22, "TargetTemp": 21,
						},
					},
				},
			},
		},
	}

	server := createSuccessServer(t, "/remoteServices/getEVVehicleStatus/v4", responseData)
	defer server.Close()

	client := createTestClient(t, server.URL)

	result, err := client.GetEVVehicleStatus(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("GetEVVehicleStatus failed: %v", err)
	}

	if result.ResultCode != "200S00" {
		t.Errorf("Expected resultCode 200S00, got %v", result.ResultCode)
	}

	if len(result.ResultData) != 1 {
		t.Errorf("Expected 1 result data, got %d", len(result.ResultData))
	}

	// Verify charge info
	chargeInfo := result.ResultData[0].PlusBInformation.VehicleInfo.ChargeInfo

	if chargeInfo.SmaphSOC != float64(85) {
		t.Errorf("Expected battery level 85, got %v", chargeInfo.SmaphSOC)
	}

	if chargeInfo.ChargerConnectorFitting != float64(1) {
		t.Errorf("Expected plugged in (1), got %v", chargeInfo.ChargerConnectorFitting)
	}
}

// TestGetEVVehicleStatus_Error tests error handling
func TestGetEVVehicleStatus_Error(t *testing.T) {
	server := createErrorServer(t, "500E00", "Internal error")
	defer server.Close()

	client := createTestClient(t, server.URL)

	_, err := client.GetEVVehicleStatus(context.Background(), "INTERNAL123")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	expectedError := "failed to get EV vehicle status: result code 500E00"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

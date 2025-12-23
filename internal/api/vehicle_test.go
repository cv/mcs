package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetVecBaseInfos tests getting vehicle base information.
func TestGetVecBaseInfos(t *testing.T) {
	t.Parallel()
	responseData := map[string]any{
		"resultCode": "200S00",
		"vecBaseInfos": []map[string]any{
			{
				"vin": "TEST123456789",
				"Vehicle": map[string]any{
					"CvInformation": map[string]any{
						"internalVin": "INTERNAL123",
					},
				},
				"econnectType": 1,
			},
		},
		"vehicleFlags": []map[string]any{
			{"vinRegistStatus": 3},
		},
	}

	server := createSuccessServer(t, "/"+EndpointGetVecBaseInfos, responseData)
	defer server.Close()

	client := createTestClient(t, server.URL)

	result, err := client.GetVecBaseInfos(context.Background())
	require.NoError(t, err, "GetVecBaseInfos failed: %v")

	assert.Equalf(t, ResultCodeSuccess, result.ResultCode, "Expected resultCode 200S00, got %v", result.ResultCode)

	assert.Lenf(t, result.VecBaseInfos, 1, "Expected 1 vehicle, got %d", len(result.VecBaseInfos))
}

// TestGetVehicleStatus tests getting vehicle status.
func TestGetVehicleStatus(t *testing.T) {
	t.Parallel()
	responseData := map[string]any{
		"resultCode": "200S00",
		"alertInfos": []map[string]any{
			{
				"OccurrenceDate": "20231201120000",
				"Door": map[string]any{
					"DrStatDrv": 0, "DrStatPsngr": 0, "DrStatRl": 0, "DrStatRr": 0,
					"DrStatTrnkLg": 0, "DrStatHood": 0, "FuelLidOpenStatus": 0,
					"LockLinkSwDrv": 0, "LockLinkSwPsngr": 0, "LockLinkSwRl": 0, "LockLinkSwRr": 0,
				},
				"Pw":         map[string]any{"PwPosDrv": 0, "PwPosPsngr": 0, "PwPosRl": 0, "PwPosRr": 0},
				"HazardLamp": map[string]any{"HazardSw": 0},
			},
		},
		"remoteInfos": []map[string]any{
			{
				"PositionInfo": map[string]any{
					"Latitude": 37.7749, "LatitudeFlag": 0,
					"Longitude": 122.4194, "LongitudeFlag": 1,
					"AcquisitionDatetime": "20231201120000",
				},
				"ResidualFuel":     map[string]any{"FuelSegementDActl": 75.5, "RemDrvDistDActlKm": 350.2},
				"DriveInformation": map[string]any{"OdoDispValue": 12345.6},
				"TPMSInformation":  map[string]any{"FLTPrsDispPsi": 32.5, "FRTPrsDispPsi": 32.0, "RLTPrsDispPsi": 31.5, "RRTPrsDispPsi": 31.8},
			},
		},
	}

	server := createSuccessServer(t, "/"+EndpointGetVehicleStatus, responseData)
	defer server.Close()

	client := createTestClient(t, server.URL)

	result, err := client.GetVehicleStatus(context.Background(), "INTERNAL123")
	require.NoError(t, err, "GetVehicleStatus failed: %v")

	assert.Equalf(t, ResultCodeSuccess, result.ResultCode, "Expected resultCode 200S00, got %v", result.ResultCode)

	assert.Lenf(t, result.AlertInfos, 1, "Expected 1 alert info, got %d", len(result.AlertInfos))

	assert.Lenf(t, result.RemoteInfos, 1, "Expected 1 remote info, got %d", len(result.RemoteInfos))
}

// TestGetVehicleStatus_Error tests error handling.
func TestGetVehicleStatus_Error(t *testing.T) {
	t.Parallel()
	server := createErrorServer(t, "500E00", "Internal error")
	defer server.Close()

	client := createTestClient(t, server.URL)

	_, err := client.GetVehicleStatus(context.Background(), "INTERNAL123")
	require.Error(t, err, "Expected error, got nil")

	expectedError := "failed to get vehicle status: result code 500E00"
	assert.Equal(t, expectedError, err.Error())
}

// TestGetEVVehicleStatus tests getting EV vehicle status.
func TestGetEVVehicleStatus(t *testing.T) {
	t.Parallel()
	responseData := map[string]any{
		"resultCode": "200S00",
		"resultData": []map[string]any{
			{
				"OccurrenceDate": "20231201120000",
				"PlusBInformation": map[string]any{
					"VehicleInfo": map[string]any{
						"ChargeInfo": map[string]any{
							"SmaphSOC": 85, "SmaphRemDrvDistKm": 245.5,
							"ChargerConnectorFitting": 1, "ChargeStatusSub": 6,
							"MaxChargeMinuteAC": 180, "MaxChargeMinuteQBC": 45,
							"CstmzStatBatHeatAutoSW": 1, "BatteryHeaterON": 0,
						},
						"RemoteHvacInfo": map[string]any{
							"HVAC": 1, "FrontDefroster": 0, "RearDefogger": 0,
							"InteriorTemp": 22, "TargetTemp": 21,
						},
					},
				},
			},
		},
	}

	server := createSuccessServer(t, "/"+EndpointGetEVVehicleStatus, responseData)
	defer server.Close()

	client := createTestClient(t, server.URL)

	result, err := client.GetEVVehicleStatus(context.Background(), "INTERNAL123")
	require.NoError(t, err, "GetEVVehicleStatus failed: %v")

	assert.Equalf(t, ResultCodeSuccess, result.ResultCode, "Expected resultCode 200S00, got %v", result.ResultCode)

	assert.Lenf(t, result.ResultData, 1, "Expected 1 result data, got %d", len(result.ResultData))

	// Verify charge info
	chargeInfo := result.ResultData[0].PlusBInformation.VehicleInfo.ChargeInfo

	assert.InDelta(t, float64(85), chargeInfo.SmaphSOC, 0.0001)

	assert.InDelta(t, float64(1), chargeInfo.ChargerConnectorFitting, 0.0001)
}

// TestGetEVVehicleStatus_Error tests error handling.
func TestGetEVVehicleStatus_Error(t *testing.T) {
	t.Parallel()
	server := createErrorServer(t, "500E00", "Internal error")
	defer server.Close()

	client := createTestClient(t, server.URL)

	_, err := client.GetEVVehicleStatus(context.Background(), "INTERNAL123")
	require.Error(t, err, "Expected error, got nil")

	expectedError := "failed to get EV vehicle status: result code 500E00"
	assert.Equal(t, expectedError, err.Error())
}

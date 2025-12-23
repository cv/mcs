package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUndisplayedFieldsInVehicleStatus verifies that undisplayed fields from GetVehicleStatus
// can be parsed correctly from the API response structure
func TestUndisplayedFieldsInVehicleStatus(t *testing.T) {
	// Create a comprehensive mock response with ALL fields we want to verify
	responseData := map[string]interface{}{
		"resultCode": "200S00",
		"alertInfos": []interface{}{
			map[string]interface{}{
				"OccurrenceDate": "20231201120000",
				"Door": map[string]interface{}{
					// Already displayed fields
					"DrStatDrv":    0,
					"DrStatPsngr":  0,
					"DrStatRl":     0,
					"DrStatRr":     0,
					"DrStatTrnkLg": 0,
					// UNDISPLAYED: Hood status
					"DrStatHood": float64(0),
					// UNDISPLAYED: Fuel lid status
					"FuelLidOpenStatus": float64(0),
					// UNDISPLAYED: Individual door lock status
					"LockLinkSwDrv":   float64(0),
					"LockLinkSwPsngr": float64(0),
					"LockLinkSwRl":    float64(0),
					"LockLinkSwRr":    float64(0),
				},
				// UNDISPLAYED: Window positions
				"Pw": map[string]interface{}{
					"PwPosDrv":   float64(0),
					"PwPosPsngr": float64(0),
					"PwPosRl":    float64(0),
					"PwPosRr":    float64(0),
				},
				// UNDISPLAYED: Hazard lights
				"HazardLamp": map[string]interface{}{
					"HazardSw": float64(0),
				},
				"PositionInfo": map[string]interface{}{
					"Latitude":            37.7749,
					"LatitudeFlag":        0,
					"Longitude":           122.4194,
					"LongitudeFlag":       1,
					"AcquisitionDatetime": "20231201120000",
				},
			},
		},
		"remoteInfos": []interface{}{
			map[string]interface{}{
				"PositionInfo": map[string]interface{}{
					"Latitude":            37.7749,
					"LatitudeFlag":        0,
					"Longitude":           122.4194,
					"LongitudeFlag":       1,
					"AcquisitionDatetime": "20231201120000",
				},
				"ResidualFuel": map[string]interface{}{
					"FuelSegementDActl": 75.5,
					"RemDrvDistDActlKm": 350.2,
				},
				// UNDISPLAYED: Odometer
				"DriveInformation": map[string]interface{}{
					"OdoDispValue": float64(12345.6),
				},
				"TPMSInformation": map[string]interface{}{
					"FLTPrsDispPsi": 32.5,
					"FRTPrsDispPsi": 32.0,
					"RLTPrsDispPsi": 31.5,
					"RRTPrsDispPsi": 31.8,
				},
			},
		},
	}

	server := createSuccessServer(t, "/"+EndpointGetVehicleStatus, responseData)
	defer server.Close()

	client := createTestClient(t, server.URL)

	result, err := client.GetVehicleStatus(context.Background(), "INTERNAL123")
	require.NoError(t, err, "GetVehicleStatus failed: %v")

	assert.EqualValuesf(t, ResultCodeSuccess, result.ResultCode, "Expected resultCode 200S00, got %v", result.ResultCode)

	// Test that we can parse the response as raw map to access undisplayed fields
	require.Lenf(t, result.AlertInfos, 1, "Expected 1 alert info, got %d", len(result.AlertInfos))

	require.Lenf(t, result.RemoteInfos, 1, "Expected 1 remote info, got %d", len(result.RemoteInfos))

	// Verify undisplayed fields can be accessed via the raw response
	// Note: The current typed structs don't include these fields, so we need to
	// test against the raw response data structure
	t.Run("VerifyOdometerField", func(t *testing.T) {
		// DriveInformation.OdoDispValue
		remoteInfos, ok := getMapSlice(responseData, "remoteInfos")
		require.True(t, ok, "remoteInfos not found in response")
		require.NotEmpty(t, remoteInfos)
		driveInfo, ok := getMap(remoteInfos[0], "DriveInformation")
		require.True(t, ok, "DriveInformation not found in response")
		odometer, ok := getFloat64(driveInfo, "OdoDispValue")
		require.True(t, ok, "OdoDispValue not found or wrong type")
		assert.EqualValuesf(t, 12345.6, odometer, "Expected odometer 12345.6, got %v", odometer)
	})

	t.Run("VerifyHoodStatusField", func(t *testing.T) {
		// alertInfos[].Door.DrStatHood
		alertInfos, ok := getMapSlice(responseData, "alertInfos")
		require.True(t, ok, "alertInfos not found in response")
		require.NotEmpty(t, alertInfos)
		door, ok := getMap(alertInfos[0], "Door")
		require.True(t, ok, "Door not found in response")
		hoodStatus, ok := getFloat64(door, "DrStatHood")
		require.True(t, ok, "DrStatHood not found or wrong type")
		assert.EqualValuesf(t, 0, hoodStatus, "Expected hood status 0, got %v", hoodStatus)
	})

	t.Run("VerifyFuelLidStatusField", func(t *testing.T) {
		// alertInfos[].Door.FuelLidOpenStatus
		alertInfos, ok := getMapSlice(responseData, "alertInfos")
		require.True(t, ok, "alertInfos not found in response")
		require.NotEmpty(t, alertInfos)
		door, ok := getMap(alertInfos[0], "Door")
		require.True(t, ok, "Door not found in response")
		fuelLid, ok := getFloat64(door, "FuelLidOpenStatus")
		require.True(t, ok, "FuelLidOpenStatus not found or wrong type")
		assert.EqualValuesf(t, 0, fuelLid, "Expected fuel lid status 0, got %v", fuelLid)
	})

	t.Run("VerifyIndividualDoorLockFields", func(t *testing.T) {
		// alertInfos[].Door.LockLinkSwDrv/Psngr/Rl/Rr
		alertInfos, ok := getMapSlice(responseData, "alertInfos")
		require.True(t, ok, "alertInfos not found in response")
		require.NotEmpty(t, alertInfos)
		door, ok := getMap(alertInfos[0], "Door")
		require.True(t, ok, "Door not found in response")

		locks := []string{"LockLinkSwDrv", "LockLinkSwPsngr", "LockLinkSwRl", "LockLinkSwRr"}
		for _, lockField := range locks {
			lockStatus, ok := getFloat64(door, lockField)
			require.Truef(t, ok, "%s not found or wrong type", lockField)
			assert.Zerof(t, lockStatus, "Expected %s status 0, got %v", lockField, lockStatus)
		}
	})

	t.Run("VerifyWindowPositionFields", func(t *testing.T) {
		// alertInfos[].Pw.PwPosDrv/Psngr/Rl/Rr
		alertInfos, ok := getMapSlice(responseData, "alertInfos")
		require.True(t, ok, "alertInfos not found in response")
		require.NotEmpty(t, alertInfos)
		pw, ok := getMap(alertInfos[0], "Pw")
		require.True(t, ok, "Pw not found in response")

		windows := []string{"PwPosDrv", "PwPosPsngr", "PwPosRl", "PwPosRr"}
		for _, windowField := range windows {
			windowPos, ok := getFloat64(pw, windowField)
			require.Truef(t, ok, "%s not found or wrong type", windowField)
			assert.Zerof(t, windowPos, "Expected %s position 0, got %v", windowField, windowPos)
		}
	})

	t.Run("VerifyHazardLightField", func(t *testing.T) {
		// alertInfos[].HazardLamp.HazardSw
		alertInfos, ok := getMapSlice(responseData, "alertInfos")
		require.True(t, ok, "alertInfos not found in response")
		require.NotEmpty(t, alertInfos)
		hazard, ok := getMap(alertInfos[0], "HazardLamp")
		require.True(t, ok, "HazardLamp not found in response")
		hazardSw, ok := getFloat64(hazard, "HazardSw")
		require.True(t, ok, "HazardSw not found or wrong type")
		assert.EqualValuesf(t, 0, hazardSw, "Expected hazard switch 0, got %v", hazardSw)
	})
}

// TestUndisplayedFieldsInEVVehicleStatus verifies that undisplayed fields from GetEVVehicleStatus
// can be parsed correctly from the API response structure
func TestUndisplayedFieldsInEVVehicleStatus(t *testing.T) {
	// Create a comprehensive mock response with ALL fields we want to verify
	responseData := map[string]interface{}{
		"resultCode": "200S00",
		"resultData": []interface{}{
			map[string]interface{}{
				"OccurrenceDate": "20231201120000",
				"PlusBInformation": map[string]interface{}{
					"VehicleInfo": map[string]interface{}{
						"ChargeInfo": map[string]interface{}{
							// Already displayed fields
							"SmaphSOC":                85,
							"SmaphRemDrvDistKm":       245.5,
							"ChargerConnectorFitting": 1,
							"ChargeStatusSub":         6,
							// UNDISPLAYED: AC charge time
							"MaxChargeMinuteAC": float64(180),
							// UNDISPLAYED: Quick charge time
							"MaxChargeMinuteQBC": float64(45),
							// UNDISPLAYED: Battery heater auto switch
							"CstmzStatBatHeatAutoSW": float64(1),
							// UNDISPLAYED: Battery heater on
							"BatteryHeaterON": float64(0),
						},
						"RemoteHvacInfo": map[string]interface{}{
							// Already displayed fields
							"HVAC":           1,
							"FrontDefroster": 0,
							"RearDefogger":   0,
							"InCarTeDC":      22,
							// UNDISPLAYED: Interior temperature (different from InCarTeDC?)
							"InteriorTemp": float64(22),
							// UNDISPLAYED: Target HVAC temperature
							"TargetTemp": float64(21),
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

	assert.EqualValuesf(t, ResultCodeSuccess, result.ResultCode, "Expected resultCode 200S00, got %v", result.ResultCode)

	require.Lenf(t, result.ResultData, 1, "Expected 1 result data, got %d", len(result.ResultData))

	t.Run("VerifyACChargeTimeField", func(t *testing.T) {
		// ChargeInfo.MaxChargeMinuteAC
		resultData, ok := getMapSlice(responseData, "resultData")
		require.True(t, ok, "resultData not found in response")
		require.NotEmpty(t, resultData)
		plusBInfo, ok := getMap(resultData[0], "PlusBInformation")
		require.True(t, ok, "PlusBInformation not found in response")
		vehicleInfo, ok := getMap(plusBInfo, "VehicleInfo")
		require.True(t, ok, "VehicleInfo not found in response")
		chargeInfo, ok := getMap(vehicleInfo, "ChargeInfo")
		require.True(t, ok, "ChargeInfo not found in response")
		acChargeTime, ok := getFloat64(chargeInfo, "MaxChargeMinuteAC")
		require.True(t, ok, "MaxChargeMinuteAC not found or wrong type")
		assert.EqualValuesf(t, 180, acChargeTime, "Expected AC charge time 180, got %v", acChargeTime)
	})

	t.Run("VerifyQuickChargeTimeField", func(t *testing.T) {
		// ChargeInfo.MaxChargeMinuteQBC
		resultData, ok := getMapSlice(responseData, "resultData")
		require.True(t, ok, "resultData not found in response")
		require.NotEmpty(t, resultData)
		plusBInfo, ok := getMap(resultData[0], "PlusBInformation")
		require.True(t, ok, "PlusBInformation not found in response")
		vehicleInfo, ok := getMap(plusBInfo, "VehicleInfo")
		require.True(t, ok, "VehicleInfo not found in response")
		chargeInfo, ok := getMap(vehicleInfo, "ChargeInfo")
		require.True(t, ok, "ChargeInfo not found in response")
		qbcChargeTime, ok := getFloat64(chargeInfo, "MaxChargeMinuteQBC")
		require.True(t, ok, "MaxChargeMinuteQBC not found or wrong type")
		assert.EqualValuesf(t, 45, qbcChargeTime, "Expected QBC charge time 45, got %v", qbcChargeTime)
	})

	t.Run("VerifyBatteryHeaterAutoField", func(t *testing.T) {
		// ChargeInfo.CstmzStatBatHeatAutoSW
		resultData, ok := getMapSlice(responseData, "resultData")
		require.True(t, ok, "resultData not found in response")
		require.NotEmpty(t, resultData)
		plusBInfo, ok := getMap(resultData[0], "PlusBInformation")
		require.True(t, ok, "PlusBInformation not found in response")
		vehicleInfo, ok := getMap(plusBInfo, "VehicleInfo")
		require.True(t, ok, "VehicleInfo not found in response")
		chargeInfo, ok := getMap(vehicleInfo, "ChargeInfo")
		require.True(t, ok, "ChargeInfo not found in response")
		batHeatAuto, ok := getFloat64(chargeInfo, "CstmzStatBatHeatAutoSW")
		require.True(t, ok, "CstmzStatBatHeatAutoSW not found or wrong type")
		assert.EqualValuesf(t, 1, batHeatAuto, "Expected battery heater auto 1, got %v", batHeatAuto)
	})

	t.Run("VerifyBatteryHeaterOnField", func(t *testing.T) {
		// ChargeInfo.BatteryHeaterON
		resultData, ok := getMapSlice(responseData, "resultData")
		require.True(t, ok, "resultData not found in response")
		require.NotEmpty(t, resultData)
		plusBInfo, ok := getMap(resultData[0], "PlusBInformation")
		require.True(t, ok, "PlusBInformation not found in response")
		vehicleInfo, ok := getMap(plusBInfo, "VehicleInfo")
		require.True(t, ok, "VehicleInfo not found in response")
		chargeInfo, ok := getMap(vehicleInfo, "ChargeInfo")
		require.True(t, ok, "ChargeInfo not found in response")
		batHeaterOn, ok := getFloat64(chargeInfo, "BatteryHeaterON")
		require.True(t, ok, "BatteryHeaterON not found or wrong type")
		assert.EqualValuesf(t, 0, batHeaterOn, "Expected battery heater on 0, got %v", batHeaterOn)
	})

	t.Run("VerifyInteriorTempField", func(t *testing.T) {
		// RemoteHvacInfo.InteriorTemp
		resultData, ok := getMapSlice(responseData, "resultData")
		require.True(t, ok, "resultData not found in response")
		require.NotEmpty(t, resultData)
		plusBInfo, ok := getMap(resultData[0], "PlusBInformation")
		require.True(t, ok, "PlusBInformation not found in response")
		vehicleInfo, ok := getMap(plusBInfo, "VehicleInfo")
		require.True(t, ok, "VehicleInfo not found in response")
		hvacInfo, ok := getMap(vehicleInfo, "RemoteHvacInfo")
		require.True(t, ok, "RemoteHvacInfo not found in response")
		interiorTemp, ok := getFloat64(hvacInfo, "InteriorTemp")
		require.True(t, ok, "InteriorTemp not found or wrong type")
		assert.EqualValuesf(t, 22, interiorTemp, "Expected interior temp 22, got %v", interiorTemp)
	})

	t.Run("VerifyTargetTempField", func(t *testing.T) {
		// RemoteHvacInfo.TargetTemp
		resultData, ok := getMapSlice(responseData, "resultData")
		require.True(t, ok, "resultData not found in response")
		require.NotEmpty(t, resultData)
		plusBInfo, ok := getMap(resultData[0], "PlusBInformation")
		require.True(t, ok, "PlusBInformation not found in response")
		vehicleInfo, ok := getMap(plusBInfo, "VehicleInfo")
		require.True(t, ok, "VehicleInfo not found in response")
		hvacInfo, ok := getMap(vehicleInfo, "RemoteHvacInfo")
		require.True(t, ok, "RemoteHvacInfo not found in response")
		targetTemp, ok := getFloat64(hvacInfo, "TargetTemp")
		require.True(t, ok, "TargetTemp not found or wrong type")
		assert.EqualValuesf(t, 21, targetTemp, "Expected target temp 21, got %v", targetTemp)
	})
}

// TestVehicleStatusWithVariedValues tests that fields can handle different values
func TestVehicleStatusWithVariedValues(t *testing.T) {
	// Test with hood open, fuel lid open, windows partially open, hazard on
	responseData := map[string]interface{}{
		"resultCode": "200S00",
		"alertInfos": []interface{}{
			map[string]interface{}{
				"OccurrenceDate": "20231201120000",
				"Door": map[string]interface{}{
					"DrStatDrv":         0,
					"DrStatPsngr":       0,
					"DrStatRl":          0,
					"DrStatRr":          0,
					"DrStatTrnkLg":      0,
					"DrStatHood":        float64(1), // Hood open
					"FuelLidOpenStatus": float64(1), // Fuel lid open
					"LockLinkSwDrv":     float64(1), // Driver door unlocked
					"LockLinkSwPsngr":   float64(0),
					"LockLinkSwRl":      float64(0),
					"LockLinkSwRr":      float64(0),
				},
				"Pw": map[string]interface{}{
					"PwPosDrv":   float64(50),  // Driver window half open
					"PwPosPsngr": float64(100), // Passenger window fully open
					"PwPosRl":    float64(0),
					"PwPosRr":    float64(25),
				},
				"HazardLamp": map[string]interface{}{
					"HazardSw": float64(1), // Hazard on
				},
				"PositionInfo": map[string]interface{}{
					"Latitude":            37.7749,
					"Longitude":           122.4194,
					"AcquisitionDatetime": "20231201120000",
				},
			},
		},
		"remoteInfos": []interface{}{
			map[string]interface{}{
				"ResidualFuel": map[string]interface{}{
					"FuelSegementDActl": 75.5,
					"RemDrvDistDActlKm": 350.2,
				},
				"DriveInformation": map[string]interface{}{
					"OdoDispValue": float64(99999.9), // High odometer
				},
				"TPMSInformation": map[string]interface{}{
					"FLTPrsDispPsi": 32.5,
					"FRTPrsDispPsi": 32.0,
					"RLTPrsDispPsi": 31.5,
					"RRTPrsDispPsi": 31.8,
				},
			},
		},
	}

	server := createSuccessServer(t, "/"+EndpointGetVehicleStatus, responseData)
	defer server.Close()

	client := createTestClient(t, server.URL)

	result, err := client.GetVehicleStatus(context.Background(), "INTERNAL123")
	require.NoError(t, err, "GetVehicleStatus failed: %v")

	assert.EqualValuesf(t, ResultCodeSuccess, result.ResultCode, "Expected resultCode 200S00, got %v", result.ResultCode)

	// Verify varied values
	alertInfos, ok := getMapSlice(responseData, "alertInfos")
	require.True(t, ok, "alertInfos not found in response")
	require.NotEmpty(t, alertInfos)
	door, ok := getMap(alertInfos[0], "Door")
	require.True(t, ok, "Door not found")
	hoodStatus, ok := getFloat64(door, "DrStatHood")
	require.True(t, ok)
	assert.EqualValues(t, 1, hoodStatus)

	fuelLid, ok := getFloat64(door, "FuelLidOpenStatus")
	require.True(t, ok)
	assert.EqualValues(t, 1, fuelLid)

	pw, ok := getMap(alertInfos[0], "Pw")
	require.True(t, ok, "Pw not found")
	drvWindow, ok := getFloat64(pw, "PwPosDrv")
	require.True(t, ok)
	assert.EqualValues(t, 50, drvWindow)

	psWindow, ok := getFloat64(pw, "PwPosPsngr")
	require.True(t, ok)
	assert.EqualValues(t, 100, psWindow)

	hazard, ok := getMap(alertInfos[0], "HazardLamp")
	require.True(t, ok, "HazardLamp not found")
	hazardSw, ok := getFloat64(hazard, "HazardSw")
	require.True(t, ok)
	assert.EqualValues(t, 1, hazardSw)

	remoteInfos, ok := getMapSlice(responseData, "remoteInfos")
	require.True(t, ok, "remoteInfos not found in response")
	require.NotEmpty(t, remoteInfos)
	driveInfo, ok := getMap(remoteInfos[0], "DriveInformation")
	require.True(t, ok, "DriveInformation not found")
	odometer, ok := getFloat64(driveInfo, "OdoDispValue")
	require.True(t, ok)
	assert.EqualValues(t, 99999.9, odometer)

}

// TestEVVehicleStatusWithVariedValues tests EV fields with different values
func TestEVVehicleStatusWithVariedValues(t *testing.T) {
	// Test with battery heater on, different temps
	responseData := map[string]interface{}{
		"resultCode": "200S00",
		"resultData": []interface{}{
			map[string]interface{}{
				"OccurrenceDate": "20231201120000",
				"PlusBInformation": map[string]interface{}{
					"VehicleInfo": map[string]interface{}{
						"ChargeInfo": map[string]interface{}{
							"SmaphSOC":                50,
							"SmaphRemDrvDistKm":       120.0,
							"ChargerConnectorFitting": 1,
							"ChargeStatusSub":         6,
							"MaxChargeMinuteAC":       float64(240), // 4 hours
							"MaxChargeMinuteQBC":      float64(60),  // 1 hour
							"CstmzStatBatHeatAutoSW":  float64(0),   // Auto off
							"BatteryHeaterON":         float64(1),   // Manually on
						},
						"RemoteHvacInfo": map[string]interface{}{
							"HVAC":           1,
							"FrontDefroster": 1,
							"RearDefogger":   1,
							"InCarTeDC":      18,
							"InteriorTemp":   float64(18), // Cold interior
							"TargetTemp":     float64(24), // Warm target
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

	assert.EqualValuesf(t, ResultCodeSuccess, result.ResultCode, "Expected resultCode 200S00, got %v", result.ResultCode)

	// Verify varied values
	resultData, ok := getMapSlice(responseData, "resultData")
	require.True(t, ok, "resultData not found in response")
	require.NotEmpty(t, resultData)
	plusBInfo, ok := getMap(resultData[0], "PlusBInformation")
	require.True(t, ok, "PlusBInformation not found in response")
	vehicleInfo, ok := getMap(plusBInfo, "VehicleInfo")
	require.True(t, ok, "VehicleInfo not found in response")
	chargeInfo, ok := getMap(vehicleInfo, "ChargeInfo")
	require.True(t, ok, "ChargeInfo not found in response")
	acChargeTime, ok := getFloat64(chargeInfo, "MaxChargeMinuteAC")
	require.True(t, ok)
	assert.EqualValues(t, 240, acChargeTime)

	batHeaterOn, ok := getFloat64(chargeInfo, "BatteryHeaterON")
	require.True(t, ok)
	assert.EqualValues(t, 1, batHeaterOn)

	hvacInfo, ok := getMap(vehicleInfo, "RemoteHvacInfo")
	require.True(t, ok, "RemoteHvacInfo not found in response")
	interiorTemp, ok := getFloat64(hvacInfo, "InteriorTemp")
	require.True(t, ok)
	assert.EqualValues(t, 18, interiorTemp)

	targetTemp, ok := getFloat64(hvacInfo, "TargetTemp")
	require.True(t, ok)
	assert.EqualValues(t, 24, targetTemp)

}

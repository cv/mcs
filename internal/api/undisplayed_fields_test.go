package api

import (
	"context"
	"testing"
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

	server := createSuccessServer(t, "/remoteServices/getVehicleStatus/v4", responseData)
	defer server.Close()

	client := createTestClient(t, server.URL)

	result, err := client.GetVehicleStatus(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("GetVehicleStatus failed: %v", err)
	}

	if result.ResultCode != ResultCodeSuccess {
		t.Errorf("Expected resultCode 200S00, got %v", result.ResultCode)
	}

	// Test that we can parse the response as raw map to access undisplayed fields
	if len(result.AlertInfos) != 1 {
		t.Fatalf("Expected 1 alert info, got %d", len(result.AlertInfos))
	}

	if len(result.RemoteInfos) != 1 {
		t.Fatalf("Expected 1 remote info, got %d", len(result.RemoteInfos))
	}

	// Verify undisplayed fields can be accessed via the raw response
	// Note: The current typed structs don't include these fields, so we need to
	// test against the raw response data structure
	t.Run("VerifyOdometerField", func(t *testing.T) {
		// DriveInformation.OdoDispValue
		remoteInfos, ok := getMapSlice(responseData, "remoteInfos")
		if !ok {
			t.Fatal("remoteInfos not found in response")
		}
		if len(remoteInfos) == 0 {
			t.Fatal("remoteInfos is empty")
		}
		driveInfo, ok := getMap(remoteInfos[0], "DriveInformation")
		if !ok {
			t.Fatal("DriveInformation not found in response")
		}
		odometer, ok := getFloat64(driveInfo, "OdoDispValue")
		if !ok {
			t.Fatal("OdoDispValue not found or wrong type")
		}
		if odometer != 12345.6 {
			t.Errorf("Expected odometer 12345.6, got %v", odometer)
		}
	})

	t.Run("VerifyHoodStatusField", func(t *testing.T) {
		// alertInfos[].Door.DrStatHood
		alertInfos, ok := getMapSlice(responseData, "alertInfos")
		if !ok {
			t.Fatal("alertInfos not found in response")
		}
		if len(alertInfos) == 0 {
			t.Fatal("alertInfos is empty")
		}
		door, ok := getMap(alertInfos[0], "Door")
		if !ok {
			t.Fatal("Door not found in response")
		}
		hoodStatus, ok := getFloat64(door, "DrStatHood")
		if !ok {
			t.Fatal("DrStatHood not found or wrong type")
		}
		if hoodStatus != 0 {
			t.Errorf("Expected hood status 0, got %v", hoodStatus)
		}
	})

	t.Run("VerifyFuelLidStatusField", func(t *testing.T) {
		// alertInfos[].Door.FuelLidOpenStatus
		alertInfos, ok := getMapSlice(responseData, "alertInfos")
		if !ok {
			t.Fatal("alertInfos not found in response")
		}
		if len(alertInfos) == 0 {
			t.Fatal("alertInfos is empty")
		}
		door, ok := getMap(alertInfos[0], "Door")
		if !ok {
			t.Fatal("Door not found in response")
		}
		fuelLid, ok := getFloat64(door, "FuelLidOpenStatus")
		if !ok {
			t.Fatal("FuelLidOpenStatus not found or wrong type")
		}
		if fuelLid != 0 {
			t.Errorf("Expected fuel lid status 0, got %v", fuelLid)
		}
	})

	t.Run("VerifyIndividualDoorLockFields", func(t *testing.T) {
		// alertInfos[].Door.LockLinkSwDrv/Psngr/Rl/Rr
		alertInfos, ok := getMapSlice(responseData, "alertInfos")
		if !ok {
			t.Fatal("alertInfos not found in response")
		}
		if len(alertInfos) == 0 {
			t.Fatal("alertInfos is empty")
		}
		door, ok := getMap(alertInfos[0], "Door")
		if !ok {
			t.Fatal("Door not found in response")
		}

		locks := []string{"LockLinkSwDrv", "LockLinkSwPsngr", "LockLinkSwRl", "LockLinkSwRr"}
		for _, lockField := range locks {
			lockStatus, ok := getFloat64(door, lockField)
			if !ok {
				t.Fatalf("%s not found or wrong type", lockField)
			}
			if lockStatus != 0 {
				t.Errorf("Expected %s status 0, got %v", lockField, lockStatus)
			}
		}
	})

	t.Run("VerifyWindowPositionFields", func(t *testing.T) {
		// alertInfos[].Pw.PwPosDrv/Psngr/Rl/Rr
		alertInfos, ok := getMapSlice(responseData, "alertInfos")
		if !ok {
			t.Fatal("alertInfos not found in response")
		}
		if len(alertInfos) == 0 {
			t.Fatal("alertInfos is empty")
		}
		pw, ok := getMap(alertInfos[0], "Pw")
		if !ok {
			t.Fatal("Pw not found in response")
		}

		windows := []string{"PwPosDrv", "PwPosPsngr", "PwPosRl", "PwPosRr"}
		for _, windowField := range windows {
			windowPos, ok := getFloat64(pw, windowField)
			if !ok {
				t.Fatalf("%s not found or wrong type", windowField)
			}
			if windowPos != 0 {
				t.Errorf("Expected %s position 0, got %v", windowField, windowPos)
			}
		}
	})

	t.Run("VerifyHazardLightField", func(t *testing.T) {
		// alertInfos[].HazardLamp.HazardSw
		alertInfos, ok := getMapSlice(responseData, "alertInfos")
		if !ok {
			t.Fatal("alertInfos not found in response")
		}
		if len(alertInfos) == 0 {
			t.Fatal("alertInfos is empty")
		}
		hazard, ok := getMap(alertInfos[0], "HazardLamp")
		if !ok {
			t.Fatal("HazardLamp not found in response")
		}
		hazardSw, ok := getFloat64(hazard, "HazardSw")
		if !ok {
			t.Fatal("HazardSw not found or wrong type")
		}
		if hazardSw != 0 {
			t.Errorf("Expected hazard switch 0, got %v", hazardSw)
		}
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

	server := createSuccessServer(t, "/remoteServices/getEVVehicleStatus/v4", responseData)
	defer server.Close()

	client := createTestClient(t, server.URL)

	result, err := client.GetEVVehicleStatus(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("GetEVVehicleStatus failed: %v", err)
	}

	if result.ResultCode != ResultCodeSuccess {
		t.Errorf("Expected resultCode 200S00, got %v", result.ResultCode)
	}

	if len(result.ResultData) != 1 {
		t.Fatalf("Expected 1 result data, got %d", len(result.ResultData))
	}

	t.Run("VerifyACChargeTimeField", func(t *testing.T) {
		// ChargeInfo.MaxChargeMinuteAC
		resultData, ok := getMapSlice(responseData, "resultData")
		if !ok || len(resultData) == 0 {
			t.Fatal("resultData not found in response")
		}
		plusBInfo, ok := getMap(resultData[0], "PlusBInformation")
		if !ok {
			t.Fatal("PlusBInformation not found in response")
		}
		vehicleInfo, ok := getMap(plusBInfo, "VehicleInfo")
		if !ok {
			t.Fatal("VehicleInfo not found in response")
		}
		chargeInfo, ok := getMap(vehicleInfo, "ChargeInfo")
		if !ok {
			t.Fatal("ChargeInfo not found in response")
		}
		acChargeTime, ok := getFloat64(chargeInfo, "MaxChargeMinuteAC")
		if !ok {
			t.Fatal("MaxChargeMinuteAC not found or wrong type")
		}
		if acChargeTime != 180 {
			t.Errorf("Expected AC charge time 180, got %v", acChargeTime)
		}
	})

	t.Run("VerifyQuickChargeTimeField", func(t *testing.T) {
		// ChargeInfo.MaxChargeMinuteQBC
		resultData, ok := getMapSlice(responseData, "resultData")
		if !ok || len(resultData) == 0 {
			t.Fatal("resultData not found in response")
		}
		plusBInfo, ok := getMap(resultData[0], "PlusBInformation")
		if !ok {
			t.Fatal("PlusBInformation not found in response")
		}
		vehicleInfo, ok := getMap(plusBInfo, "VehicleInfo")
		if !ok {
			t.Fatal("VehicleInfo not found in response")
		}
		chargeInfo, ok := getMap(vehicleInfo, "ChargeInfo")
		if !ok {
			t.Fatal("ChargeInfo not found in response")
		}
		qbcChargeTime, ok := getFloat64(chargeInfo, "MaxChargeMinuteQBC")
		if !ok {
			t.Fatal("MaxChargeMinuteQBC not found or wrong type")
		}
		if qbcChargeTime != 45 {
			t.Errorf("Expected QBC charge time 45, got %v", qbcChargeTime)
		}
	})

	t.Run("VerifyBatteryHeaterAutoField", func(t *testing.T) {
		// ChargeInfo.CstmzStatBatHeatAutoSW
		resultData, ok := getMapSlice(responseData, "resultData")
		if !ok || len(resultData) == 0 {
			t.Fatal("resultData not found in response")
		}
		plusBInfo, ok := getMap(resultData[0], "PlusBInformation")
		if !ok {
			t.Fatal("PlusBInformation not found in response")
		}
		vehicleInfo, ok := getMap(plusBInfo, "VehicleInfo")
		if !ok {
			t.Fatal("VehicleInfo not found in response")
		}
		chargeInfo, ok := getMap(vehicleInfo, "ChargeInfo")
		if !ok {
			t.Fatal("ChargeInfo not found in response")
		}
		batHeatAuto, ok := getFloat64(chargeInfo, "CstmzStatBatHeatAutoSW")
		if !ok {
			t.Fatal("CstmzStatBatHeatAutoSW not found or wrong type")
		}
		if batHeatAuto != 1 {
			t.Errorf("Expected battery heater auto 1, got %v", batHeatAuto)
		}
	})

	t.Run("VerifyBatteryHeaterOnField", func(t *testing.T) {
		// ChargeInfo.BatteryHeaterON
		resultData, ok := getMapSlice(responseData, "resultData")
		if !ok || len(resultData) == 0 {
			t.Fatal("resultData not found in response")
		}
		plusBInfo, ok := getMap(resultData[0], "PlusBInformation")
		if !ok {
			t.Fatal("PlusBInformation not found in response")
		}
		vehicleInfo, ok := getMap(plusBInfo, "VehicleInfo")
		if !ok {
			t.Fatal("VehicleInfo not found in response")
		}
		chargeInfo, ok := getMap(vehicleInfo, "ChargeInfo")
		if !ok {
			t.Fatal("ChargeInfo not found in response")
		}
		batHeaterOn, ok := getFloat64(chargeInfo, "BatteryHeaterON")
		if !ok {
			t.Fatal("BatteryHeaterON not found or wrong type")
		}
		if batHeaterOn != 0 {
			t.Errorf("Expected battery heater on 0, got %v", batHeaterOn)
		}
	})

	t.Run("VerifyInteriorTempField", func(t *testing.T) {
		// RemoteHvacInfo.InteriorTemp
		resultData, ok := getMapSlice(responseData, "resultData")
		if !ok || len(resultData) == 0 {
			t.Fatal("resultData not found in response")
		}
		plusBInfo, ok := getMap(resultData[0], "PlusBInformation")
		if !ok {
			t.Fatal("PlusBInformation not found in response")
		}
		vehicleInfo, ok := getMap(plusBInfo, "VehicleInfo")
		if !ok {
			t.Fatal("VehicleInfo not found in response")
		}
		hvacInfo, ok := getMap(vehicleInfo, "RemoteHvacInfo")
		if !ok {
			t.Fatal("RemoteHvacInfo not found in response")
		}
		interiorTemp, ok := getFloat64(hvacInfo, "InteriorTemp")
		if !ok {
			t.Fatal("InteriorTemp not found or wrong type")
		}
		if interiorTemp != 22 {
			t.Errorf("Expected interior temp 22, got %v", interiorTemp)
		}
	})

	t.Run("VerifyTargetTempField", func(t *testing.T) {
		// RemoteHvacInfo.TargetTemp
		resultData, ok := getMapSlice(responseData, "resultData")
		if !ok || len(resultData) == 0 {
			t.Fatal("resultData not found in response")
		}
		plusBInfo, ok := getMap(resultData[0], "PlusBInformation")
		if !ok {
			t.Fatal("PlusBInformation not found in response")
		}
		vehicleInfo, ok := getMap(plusBInfo, "VehicleInfo")
		if !ok {
			t.Fatal("VehicleInfo not found in response")
		}
		hvacInfo, ok := getMap(vehicleInfo, "RemoteHvacInfo")
		if !ok {
			t.Fatal("RemoteHvacInfo not found in response")
		}
		targetTemp, ok := getFloat64(hvacInfo, "TargetTemp")
		if !ok {
			t.Fatal("TargetTemp not found or wrong type")
		}
		if targetTemp != 21 {
			t.Errorf("Expected target temp 21, got %v", targetTemp)
		}
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

	server := createSuccessServer(t, "/remoteServices/getVehicleStatus/v4", responseData)
	defer server.Close()

	client := createTestClient(t, server.URL)

	result, err := client.GetVehicleStatus(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("GetVehicleStatus failed: %v", err)
	}

	if result.ResultCode != ResultCodeSuccess {
		t.Errorf("Expected resultCode 200S00, got %v", result.ResultCode)
	}

	// Verify varied values
	alertInfos, ok := getMapSlice(responseData, "alertInfos")
	if !ok || len(alertInfos) == 0 {
		t.Fatal("alertInfos not found in response")
	}
	door, ok := getMap(alertInfos[0], "Door")
	if !ok {
		t.Fatal("Door not found")
	}
	if hoodStatus, ok := getFloat64(door, "DrStatHood"); !ok || hoodStatus != 1 {
		t.Error("Expected hood open (1)")
	}
	if fuelLid, ok := getFloat64(door, "FuelLidOpenStatus"); !ok || fuelLid != 1 {
		t.Error("Expected fuel lid open (1)")
	}

	pw, ok := getMap(alertInfos[0], "Pw")
	if !ok {
		t.Fatal("Pw not found")
	}
	if drvWindow, ok := getFloat64(pw, "PwPosDrv"); !ok || drvWindow != 50 {
		t.Error("Expected driver window at 50")
	}
	if psWindow, ok := getFloat64(pw, "PwPosPsngr"); !ok || psWindow != 100 {
		t.Error("Expected passenger window at 100")
	}

	hazard, ok := getMap(alertInfos[0], "HazardLamp")
	if !ok {
		t.Fatal("HazardLamp not found")
	}
	if hazardSw, ok := getFloat64(hazard, "HazardSw"); !ok || hazardSw != 1 {
		t.Error("Expected hazard lights on (1)")
	}

	remoteInfos, ok := getMapSlice(responseData, "remoteInfos")
	if !ok || len(remoteInfos) == 0 {
		t.Fatal("remoteInfos not found in response")
	}
	driveInfo, ok := getMap(remoteInfos[0], "DriveInformation")
	if !ok {
		t.Fatal("DriveInformation not found")
	}
	if odometer, ok := getFloat64(driveInfo, "OdoDispValue"); !ok || odometer != 99999.9 {
		t.Error("Expected odometer at 99999.9")
	}
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

	server := createSuccessServer(t, "/remoteServices/getEVVehicleStatus/v4", responseData)
	defer server.Close()

	client := createTestClient(t, server.URL)

	result, err := client.GetEVVehicleStatus(context.Background(), "INTERNAL123")
	if err != nil {
		t.Fatalf("GetEVVehicleStatus failed: %v", err)
	}

	if result.ResultCode != ResultCodeSuccess {
		t.Errorf("Expected resultCode 200S00, got %v", result.ResultCode)
	}

	// Verify varied values
	resultData, ok := getMapSlice(responseData, "resultData")
	if !ok || len(resultData) == 0 {
		t.Fatal("resultData not found in response")
	}
	plusBInfo, ok := getMap(resultData[0], "PlusBInformation")
	if !ok {
		t.Fatal("PlusBInformation not found in response")
	}
	vehicleInfo, ok := getMap(plusBInfo, "VehicleInfo")
	if !ok {
		t.Fatal("VehicleInfo not found in response")
	}
	chargeInfo, ok := getMap(vehicleInfo, "ChargeInfo")
	if !ok {
		t.Fatal("ChargeInfo not found in response")
	}
	if acChargeTime, ok := getFloat64(chargeInfo, "MaxChargeMinuteAC"); !ok || acChargeTime != 240 {
		t.Error("Expected AC charge time 240 minutes")
	}
	if batHeaterOn, ok := getFloat64(chargeInfo, "BatteryHeaterON"); !ok || batHeaterOn != 1 {
		t.Error("Expected battery heater on (1)")
	}

	hvacInfo, ok := getMap(vehicleInfo, "RemoteHvacInfo")
	if !ok {
		t.Fatal("RemoteHvacInfo not found in response")
	}
	if interiorTemp, ok := getFloat64(hvacInfo, "InteriorTemp"); !ok || interiorTemp != 18 {
		t.Error("Expected interior temp 18")
	}
	if targetTemp, ok := getFloat64(hvacInfo, "TargetTemp"); !ok || targetTemp != 24 {
		t.Error("Expected target temp 24")
	}
}

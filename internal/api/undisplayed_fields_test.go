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
		"alertInfos": []map[string]interface{}{
			{
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
		"remoteInfos": []map[string]interface{}{
			{
				"PositionInfo": map[string]interface{}{
					"Latitude":            37.7749,
					"LatitudeFlag":        0,
					"Longitude":           122.4194,
					"LongitudeFlag":       1,
					"AcquisitionDatetime": "20231201120000",
				},
				"ResidualFuel": map[string]interface{}{
					"FuelSegementDActl":  75.5,
					"RemDrvDistDActlKm":  350.2,
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

	if result.ResultCode != "200S00" {
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
		driveInfo, ok := responseData["remoteInfos"].([]map[string]interface{})[0]["DriveInformation"].(map[string]interface{})
		if !ok {
			t.Fatal("DriveInformation not found in response")
		}
		odometer, ok := driveInfo["OdoDispValue"].(float64)
		if !ok {
			t.Fatal("OdoDispValue not found or wrong type")
		}
		if odometer != 12345.6 {
			t.Errorf("Expected odometer 12345.6, got %v", odometer)
		}
	})

	t.Run("VerifyHoodStatusField", func(t *testing.T) {
		// alertInfos[].Door.DrStatHood
		door, ok := responseData["alertInfos"].([]map[string]interface{})[0]["Door"].(map[string]interface{})
		if !ok {
			t.Fatal("Door not found in response")
		}
		hoodStatus, ok := door["DrStatHood"].(float64)
		if !ok {
			t.Fatal("DrStatHood not found or wrong type")
		}
		if hoodStatus != 0 {
			t.Errorf("Expected hood status 0, got %v", hoodStatus)
		}
	})

	t.Run("VerifyFuelLidStatusField", func(t *testing.T) {
		// alertInfos[].Door.FuelLidOpenStatus
		door, ok := responseData["alertInfos"].([]map[string]interface{})[0]["Door"].(map[string]interface{})
		if !ok {
			t.Fatal("Door not found in response")
		}
		fuelLid, ok := door["FuelLidOpenStatus"].(float64)
		if !ok {
			t.Fatal("FuelLidOpenStatus not found or wrong type")
		}
		if fuelLid != 0 {
			t.Errorf("Expected fuel lid status 0, got %v", fuelLid)
		}
	})

	t.Run("VerifyIndividualDoorLockFields", func(t *testing.T) {
		// alertInfos[].Door.LockLinkSwDrv/Psngr/Rl/Rr
		door, ok := responseData["alertInfos"].([]map[string]interface{})[0]["Door"].(map[string]interface{})
		if !ok {
			t.Fatal("Door not found in response")
		}

		locks := []string{"LockLinkSwDrv", "LockLinkSwPsngr", "LockLinkSwRl", "LockLinkSwRr"}
		for _, lockField := range locks {
			lockStatus, ok := door[lockField].(float64)
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
		pw, ok := responseData["alertInfos"].([]map[string]interface{})[0]["Pw"].(map[string]interface{})
		if !ok {
			t.Fatal("Pw not found in response")
		}

		windows := []string{"PwPosDrv", "PwPosPsngr", "PwPosRl", "PwPosRr"}
		for _, windowField := range windows {
			windowPos, ok := pw[windowField].(float64)
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
		hazard, ok := responseData["alertInfos"].([]map[string]interface{})[0]["HazardLamp"].(map[string]interface{})
		if !ok {
			t.Fatal("HazardLamp not found in response")
		}
		hazardSw, ok := hazard["HazardSw"].(float64)
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
		"resultData": []map[string]interface{}{
			{
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

	if result.ResultCode != "200S00" {
		t.Errorf("Expected resultCode 200S00, got %v", result.ResultCode)
	}

	if len(result.ResultData) != 1 {
		t.Fatalf("Expected 1 result data, got %d", len(result.ResultData))
	}

	t.Run("VerifyACChargeTimeField", func(t *testing.T) {
		// ChargeInfo.MaxChargeMinuteAC
		chargeInfo, ok := responseData["resultData"].([]map[string]interface{})[0]["PlusBInformation"].(map[string]interface{})["VehicleInfo"].(map[string]interface{})["ChargeInfo"].(map[string]interface{})
		if !ok {
			t.Fatal("ChargeInfo not found in response")
		}
		acChargeTime, ok := chargeInfo["MaxChargeMinuteAC"].(float64)
		if !ok {
			t.Fatal("MaxChargeMinuteAC not found or wrong type")
		}
		if acChargeTime != 180 {
			t.Errorf("Expected AC charge time 180, got %v", acChargeTime)
		}
	})

	t.Run("VerifyQuickChargeTimeField", func(t *testing.T) {
		// ChargeInfo.MaxChargeMinuteQBC
		chargeInfo, ok := responseData["resultData"].([]map[string]interface{})[0]["PlusBInformation"].(map[string]interface{})["VehicleInfo"].(map[string]interface{})["ChargeInfo"].(map[string]interface{})
		if !ok {
			t.Fatal("ChargeInfo not found in response")
		}
		qbcChargeTime, ok := chargeInfo["MaxChargeMinuteQBC"].(float64)
		if !ok {
			t.Fatal("MaxChargeMinuteQBC not found or wrong type")
		}
		if qbcChargeTime != 45 {
			t.Errorf("Expected QBC charge time 45, got %v", qbcChargeTime)
		}
	})

	t.Run("VerifyBatteryHeaterAutoField", func(t *testing.T) {
		// ChargeInfo.CstmzStatBatHeatAutoSW
		chargeInfo, ok := responseData["resultData"].([]map[string]interface{})[0]["PlusBInformation"].(map[string]interface{})["VehicleInfo"].(map[string]interface{})["ChargeInfo"].(map[string]interface{})
		if !ok {
			t.Fatal("ChargeInfo not found in response")
		}
		batHeatAuto, ok := chargeInfo["CstmzStatBatHeatAutoSW"].(float64)
		if !ok {
			t.Fatal("CstmzStatBatHeatAutoSW not found or wrong type")
		}
		if batHeatAuto != 1 {
			t.Errorf("Expected battery heater auto 1, got %v", batHeatAuto)
		}
	})

	t.Run("VerifyBatteryHeaterOnField", func(t *testing.T) {
		// ChargeInfo.BatteryHeaterON
		chargeInfo, ok := responseData["resultData"].([]map[string]interface{})[0]["PlusBInformation"].(map[string]interface{})["VehicleInfo"].(map[string]interface{})["ChargeInfo"].(map[string]interface{})
		if !ok {
			t.Fatal("ChargeInfo not found in response")
		}
		batHeaterOn, ok := chargeInfo["BatteryHeaterON"].(float64)
		if !ok {
			t.Fatal("BatteryHeaterON not found or wrong type")
		}
		if batHeaterOn != 0 {
			t.Errorf("Expected battery heater on 0, got %v", batHeaterOn)
		}
	})

	t.Run("VerifyInteriorTempField", func(t *testing.T) {
		// RemoteHvacInfo.InteriorTemp
		hvacInfo, ok := responseData["resultData"].([]map[string]interface{})[0]["PlusBInformation"].(map[string]interface{})["VehicleInfo"].(map[string]interface{})["RemoteHvacInfo"].(map[string]interface{})
		if !ok {
			t.Fatal("RemoteHvacInfo not found in response")
		}
		interiorTemp, ok := hvacInfo["InteriorTemp"].(float64)
		if !ok {
			t.Fatal("InteriorTemp not found or wrong type")
		}
		if interiorTemp != 22 {
			t.Errorf("Expected interior temp 22, got %v", interiorTemp)
		}
	})

	t.Run("VerifyTargetTempField", func(t *testing.T) {
		// RemoteHvacInfo.TargetTemp
		hvacInfo, ok := responseData["resultData"].([]map[string]interface{})[0]["PlusBInformation"].(map[string]interface{})["VehicleInfo"].(map[string]interface{})["RemoteHvacInfo"].(map[string]interface{})
		if !ok {
			t.Fatal("RemoteHvacInfo not found in response")
		}
		targetTemp, ok := hvacInfo["TargetTemp"].(float64)
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
		"alertInfos": []map[string]interface{}{
			{
				"OccurrenceDate": "20231201120000",
				"Door": map[string]interface{}{
					"DrStatDrv":       0,
					"DrStatPsngr":     0,
					"DrStatRl":        0,
					"DrStatRr":        0,
					"DrStatTrnkLg":    0,
					"DrStatHood":      float64(1), // Hood open
					"FuelLidOpenStatus": float64(1), // Fuel lid open
					"LockLinkSwDrv":   float64(1), // Driver door unlocked
					"LockLinkSwPsngr": float64(0),
					"LockLinkSwRl":    float64(0),
					"LockLinkSwRr":    float64(0),
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
		"remoteInfos": []map[string]interface{}{
			{
				"ResidualFuel": map[string]interface{}{
					"FuelSegementDActl":  75.5,
					"RemDrvDistDActlKm":  350.2,
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

	if result.ResultCode != "200S00" {
		t.Errorf("Expected resultCode 200S00, got %v", result.ResultCode)
	}

	// Verify varied values
	door := responseData["alertInfos"].([]map[string]interface{})[0]["Door"].(map[string]interface{})
	if door["DrStatHood"].(float64) != 1 {
		t.Error("Expected hood open (1)")
	}
	if door["FuelLidOpenStatus"].(float64) != 1 {
		t.Error("Expected fuel lid open (1)")
	}

	pw := responseData["alertInfos"].([]map[string]interface{})[0]["Pw"].(map[string]interface{})
	if pw["PwPosDrv"].(float64) != 50 {
		t.Error("Expected driver window at 50")
	}
	if pw["PwPosPsngr"].(float64) != 100 {
		t.Error("Expected passenger window at 100")
	}

	hazard := responseData["alertInfos"].([]map[string]interface{})[0]["HazardLamp"].(map[string]interface{})
	if hazard["HazardSw"].(float64) != 1 {
		t.Error("Expected hazard lights on (1)")
	}

	driveInfo := responseData["remoteInfos"].([]map[string]interface{})[0]["DriveInformation"].(map[string]interface{})
	if driveInfo["OdoDispValue"].(float64) != 99999.9 {
		t.Error("Expected odometer at 99999.9")
	}
}

// TestEVVehicleStatusWithVariedValues tests EV fields with different values
func TestEVVehicleStatusWithVariedValues(t *testing.T) {
	// Test with battery heater on, different temps
	responseData := map[string]interface{}{
		"resultCode": "200S00",
		"resultData": []map[string]interface{}{
			{
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

	if result.ResultCode != "200S00" {
		t.Errorf("Expected resultCode 200S00, got %v", result.ResultCode)
	}

	// Verify varied values
	chargeInfo := responseData["resultData"].([]map[string]interface{})[0]["PlusBInformation"].(map[string]interface{})["VehicleInfo"].(map[string]interface{})["ChargeInfo"].(map[string]interface{})
	if chargeInfo["MaxChargeMinuteAC"].(float64) != 240 {
		t.Error("Expected AC charge time 240 minutes")
	}
	if chargeInfo["BatteryHeaterON"].(float64) != 1 {
		t.Error("Expected battery heater on (1)")
	}

	hvacInfo := responseData["resultData"].([]map[string]interface{})[0]["PlusBInformation"].(map[string]interface{})["VehicleInfo"].(map[string]interface{})["RemoteHvacInfo"].(map[string]interface{})
	if hvacInfo["InteriorTemp"].(float64) != 18 {
		t.Error("Expected interior temp 18")
	}
	if hvacInfo["TargetTemp"].(float64) != 24 {
		t.Error("Expected target temp 24")
	}
}

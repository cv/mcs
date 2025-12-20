package cli

import (
	"testing"

	"github.com/cv/mcs/internal/api"
)

// TestBatteryInfoToMap tests batteryInfoToMap conversion
func TestBatteryInfoToMap(t *testing.T) {
	tests := []struct {
		name        string
		batteryInfo api.BatteryInfo
		wantFields  map[string]interface{}
	}{
		{
			name: "charging with time estimates",
			batteryInfo: api.BatteryInfo{
				BatteryLevel:     66,
				RangeKm:          245.5,
				ChargeTimeACMin:  180,
				ChargeTimeQBCMin: 45,
				PluggedIn:        true,
				Charging:         true,
				HeaterOn:         false,
				HeaterAuto:       false,
			},
			wantFields: map[string]interface{}{
				"battery_level":           float64(66),
				"range_km":                245.5,
				"plugged_in":              true,
				"charging":                true,
				"heater_on":               false,
				"heater_auto":             false,
				"charge_time_ac_minutes":  float64(180),
				"charge_time_qbc_minutes": float64(45),
			},
		},
		{
			name: "not charging",
			batteryInfo: api.BatteryInfo{
				BatteryLevel:     50,
				RangeKm:          150.0,
				ChargeTimeACMin:  0,
				ChargeTimeQBCMin: 0,
				PluggedIn:        false,
				Charging:         false,
				HeaterOn:         true,
				HeaterAuto:       true,
			},
			wantFields: map[string]interface{}{
				"battery_level": float64(50),
				"range_km":      150.0,
				"plugged_in":    false,
				"charging":      false,
				"heater_on":     true,
				"heater_auto":   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := batteryInfoToMap(tt.batteryInfo)

			for key, expected := range tt.wantFields {
				actual, ok := data[key]
				if !ok {
					t.Errorf("Expected key %q to exist in map", key)
					continue
				}
				if actual != expected {
					t.Errorf("Expected %s to be %v, got %v", key, expected, actual)
				}
			}

			// Verify charge time fields are only present when charging
			if !tt.batteryInfo.Charging {
				if _, ok := data["charge_time_ac_minutes"]; ok {
					t.Error("Expected charge_time_ac_minutes to not be present when not charging")
				}
				if _, ok := data["charge_time_qbc_minutes"]; ok {
					t.Error("Expected charge_time_qbc_minutes to not be present when not charging")
				}
			}
		})
	}
}

// TestFuelInfoToMap tests fuelInfoToMap conversion
func TestFuelInfoToMap(t *testing.T) {
	fuelInfo := api.FuelInfo{
		FuelLevel: 92,
		RangeKm:   630.0,
	}

	data := fuelInfoToMap(fuelInfo)

	assertMapValue(t, data, "fuel_level", float64(92))
	assertMapValue(t, data, "range_km", 630.0)
}

// TestLocationInfoToMap tests locationInfoToMap conversion
func TestLocationInfoToMap(t *testing.T) {
	locationInfo := api.LocationInfo{
		Latitude:  37.7749,
		Longitude: -122.4194,
		Timestamp: "20231201120000",
	}

	data := locationInfoToMap(locationInfo)

	assertMapValue(t, data, "latitude", 37.7749)
	assertMapValue(t, data, "longitude", -122.4194)
	assertMapValue(t, data, "timestamp", "20231201120000")

	mapsURL, ok := data["maps_url"].(string)
	if !ok {
		t.Fatal("Expected maps_url to be a string")
	}
	expectedURL := "https://maps.google.com/?q=37.774900,-122.419400"
	if mapsURL != expectedURL {
		t.Errorf("Expected maps_url %q, got %q", expectedURL, mapsURL)
	}
}

// TestTireInfoToMap tests tireInfoToMap conversion
func TestTireInfoToMap(t *testing.T) {
	tireInfo := api.TireInfo{
		FrontLeftPsi:  32.5,
		FrontRightPsi: 32.0,
		RearLeftPsi:   31.5,
		RearRightPsi:  31.8,
	}

	data := tireInfoToMap(tireInfo)

	assertMapValue(t, data, "front_left_psi", 32.5)
	assertMapValue(t, data, "front_right_psi", 32.0)
	assertMapValue(t, data, "rear_left_psi", 31.5)
	assertMapValue(t, data, "rear_right_psi", 31.8)
}

// TestDoorStatusToMap tests doorStatusToMap conversion
func TestDoorStatusToMap(t *testing.T) {
	doorStatus := api.DoorStatus{
		AllLocked:       true,
		DriverOpen:      false,
		PassengerOpen:   false,
		RearLeftOpen:    false,
		RearRightOpen:   false,
		TrunkOpen:       false,
		HoodOpen:        false,
		FuelLidOpen:     false,
		DriverLocked:    true,
		PassengerLocked: true,
		RearLeftLocked:  true,
		RearRightLocked: true,
	}

	data := doorStatusToMap(doorStatus)

	assertMapValue(t, data, "all_locked", true)
	assertMapValue(t, data, "driver_open", false)
	assertMapValue(t, data, "driver_locked", true)
}

// TestOdometerInfoToMap tests odometerInfoToMap conversion
func TestOdometerInfoToMap(t *testing.T) {
	odometerInfo := api.OdometerInfo{OdometerKm: 12345.6}

	data := odometerInfoToMap(odometerInfo)

	assertMapValue(t, data, "odometer_km", 12345.6)
}

// TestHvacInfoToMap tests hvacInfoToMap conversion
func TestHvacInfoToMap(t *testing.T) {
	hvacInfo := api.HVACInfo{
		HVACOn:         true,
		FrontDefroster: true,
		RearDefroster:  false,
		InteriorTempC:  21,
		TargetTempC:    22,
	}

	data := hvacInfoToMap(hvacInfo)

	assertMapValue(t, data, "hvac_on", true)
	assertMapValue(t, data, "front_defroster", true)
	assertMapValue(t, data, "rear_defroster", false)
	assertMapValue(t, data, "interior_temperature_c", float64(21))
	assertMapValue(t, data, "target_temperature_c", float64(22))
}

// TestWindowStatusToMap tests windowStatusToMap conversion
func TestWindowStatusToMap(t *testing.T) {
	windowStatus := api.WindowStatus{
		DriverPosition:    25,
		PassengerPosition: 50,
		RearLeftPosition:  75,
		RearRightPosition: 100,
	}

	data := windowStatusToMap(windowStatus)

	assertMapValue(t, data, "driver_position", float64(25))
	assertMapValue(t, data, "passenger_position", float64(50))
	assertMapValue(t, data, "rear_left_position", float64(75))
	assertMapValue(t, data, "rear_right_position", float64(100))
}

// TestExtractVehicleInfoDataHelper tests vehicle info extraction
func TestExtractVehicleInfoDataHelper(t *testing.T) {
	info := VehicleInfo{
		VIN:       "JM3KKEHC1R0123456",
		Nickname:  "My CX-90",
		ModelName: "CX-90 PHEV",
		ModelYear: "2024",
	}

	data := extractVehicleInfoData(info)

	assertMapValue(t, data, "vin", "JM3KKEHC1R0123456")
	assertMapValue(t, data, "nickname", "My CX-90")
	assertMapValue(t, data, "model_name", "CX-90 PHEV")
	assertMapValue(t, data, "model_year", "2024")
}

// TestExtractWithGetter tests the generic extraction helper
func TestExtractWithGetter(t *testing.T) {
	// Test with fuel info
	vehicleStatus := &api.VehicleStatusResponse{
		ResultCode: api.ResultCodeSuccess,
		RemoteInfos: []api.RemoteInfo{
			{
				ResidualFuel: api.ResidualFuel{
					FuelSegmentDActl:  92,
					RemDrvDistDActlKm: 630.0,
				},
			},
		},
	}

	data := extractWithGetter(
		vehicleStatus.GetFuelInfo,
		fuelInfoToMap,
	)

	assertMapValue(t, data, "fuel_level", float64(92))
	assertMapValue(t, data, "range_km", 630.0)
}

// TestExtractWithGetterError tests the generic extraction helper with an error case
func TestExtractWithGetterError(t *testing.T) {
	// Test with empty response that will cause an error
	vehicleStatus := &api.VehicleStatusResponse{
		ResultCode:  api.ResultCodeSuccess,
		RemoteInfos: []api.RemoteInfo{},
	}

	// Should return empty map when getter fails
	data := extractWithGetter(
		vehicleStatus.GetFuelInfo,
		fuelInfoToMap,
	)

	if len(data) != 0 {
		t.Errorf("Expected empty map when getter fails, got %v", data)
	}
}

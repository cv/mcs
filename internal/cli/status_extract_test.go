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

	if data["fuel_level"] != float64(92) {
		t.Errorf("Expected fuel_level 92, got %v", data["fuel_level"])
	}
	if data["range_km"] != 630.0 {
		t.Errorf("Expected range_km 630.0, got %v", data["range_km"])
	}
}

// TestLocationInfoToMap tests locationInfoToMap conversion
func TestLocationInfoToMap(t *testing.T) {
	locationInfo := api.LocationInfo{
		Latitude:  37.7749,
		Longitude: -122.4194,
		Timestamp: "20231201120000",
	}

	data := locationInfoToMap(locationInfo)

	if data["latitude"] != 37.7749 {
		t.Errorf("Expected latitude 37.7749, got %v", data["latitude"])
	}
	if data["longitude"] != -122.4194 {
		t.Errorf("Expected longitude -122.4194, got %v", data["longitude"])
	}
	if data["timestamp"] != "20231201120000" {
		t.Errorf("Expected timestamp 20231201120000, got %v", data["timestamp"])
	}

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

	if data["front_left_psi"] != 32.5 {
		t.Errorf("Expected front_left_psi 32.5, got %v", data["front_left_psi"])
	}
	if data["front_right_psi"] != 32.0 {
		t.Errorf("Expected front_right_psi 32.0, got %v", data["front_right_psi"])
	}
	if data["rear_left_psi"] != 31.5 {
		t.Errorf("Expected rear_left_psi 31.5, got %v", data["rear_left_psi"])
	}
	if data["rear_right_psi"] != 31.8 {
		t.Errorf("Expected rear_right_psi 31.8, got %v", data["rear_right_psi"])
	}
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

	if data["all_locked"] != true {
		t.Errorf("Expected all_locked true, got %v", data["all_locked"])
	}
	if data["driver_open"] != false {
		t.Errorf("Expected driver_open false, got %v", data["driver_open"])
	}
	if data["driver_locked"] != true {
		t.Errorf("Expected driver_locked true, got %v", data["driver_locked"])
	}
}

// TestOdometerInfoToMap tests odometerInfoToMap conversion
func TestOdometerInfoToMap(t *testing.T) {
	odometerInfo := api.OdometerInfo{OdometerKm: 12345.6}

	data := odometerInfoToMap(odometerInfo)

	if data["odometer_km"] != 12345.6 {
		t.Errorf("Expected odometer_km 12345.6, got %v", data["odometer_km"])
	}
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

	if data["hvac_on"] != true {
		t.Errorf("Expected hvac_on true, got %v", data["hvac_on"])
	}
	if data["front_defroster"] != true {
		t.Errorf("Expected front_defroster true, got %v", data["front_defroster"])
	}
	if data["rear_defroster"] != false {
		t.Errorf("Expected rear_defroster false, got %v", data["rear_defroster"])
	}
	if data["interior_temperature_c"] != float64(21) {
		t.Errorf("Expected interior_temperature_c 21, got %v", data["interior_temperature_c"])
	}
	if data["target_temperature_c"] != float64(22) {
		t.Errorf("Expected target_temperature_c 22, got %v", data["target_temperature_c"])
	}
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

	if data["driver_position"] != float64(25) {
		t.Errorf("Expected driver_position 25, got %v", data["driver_position"])
	}
	if data["passenger_position"] != float64(50) {
		t.Errorf("Expected passenger_position 50, got %v", data["passenger_position"])
	}
	if data["rear_left_position"] != float64(75) {
		t.Errorf("Expected rear_left_position 75, got %v", data["rear_left_position"])
	}
	if data["rear_right_position"] != float64(100) {
		t.Errorf("Expected rear_right_position 100, got %v", data["rear_right_position"])
	}
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

	if data["vin"] != "JM3KKEHC1R0123456" {
		t.Errorf("Expected vin JM3KKEHC1R0123456, got %v", data["vin"])
	}
	if data["nickname"] != "My CX-90" {
		t.Errorf("Expected nickname My CX-90, got %v", data["nickname"])
	}
	if data["model_name"] != "CX-90 PHEV" {
		t.Errorf("Expected model_name CX-90 PHEV, got %v", data["model_name"])
	}
	if data["model_year"] != "2024" {
		t.Errorf("Expected model_year 2024, got %v", data["model_year"])
	}
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

	if data["fuel_level"] != float64(92) {
		t.Errorf("Expected fuel_level 92, got %v", data["fuel_level"])
	}
	if data["range_km"] != 630.0 {
		t.Errorf("Expected range_km 630.0, got %v", data["range_km"])
	}
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

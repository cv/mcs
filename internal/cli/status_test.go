package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cv/mcs/internal/api"
)

// parseJSONToMap is a test helper that parses a JSON string into a map
func parseJSONToMap(t *testing.T, jsonStr string) map[string]interface{} {
	t.Helper()
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}
	return data
}

// assertMapValue is a test helper that asserts a map value equals expected
func assertMapValue(t *testing.T, data map[string]interface{}, key string, expected interface{}) {
	t.Helper()
	actual, ok := data[key]
	if !ok {
		t.Errorf("Expected key %q to exist in map", key)
		return
	}
	if actual != expected {
		t.Errorf("Expected %s to be %v, got %v", key, expected, actual)
	}
}

// TestStatusCommand tests the status command
func TestStatusCommand(t *testing.T) {
	cmd := NewStatusCmd()

	if cmd.Use != "status" {
		t.Errorf("Expected Use to be 'status', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}

// TestStatusCommand_NoSubcommand tests status command without subcommand
func TestStatusCommand_NoSubcommand(t *testing.T) {
	// This should show all status information
	cmd := NewStatusCmd()
	cmd.SetArgs([]string{})

	// We need to inject a mock client - this will be handled in the actual implementation
	// For now, we test that the command structure is correct
	if err := cmd.ValidateArgs([]string{}); err != nil {
		t.Errorf("Status command should accept no arguments: %v", err)
	}
}

// TestStatusCommand_Subcommands tests all status subcommands using table-driven pattern
func TestStatusCommand_Subcommands(t *testing.T) {
	subcommands := []string{"battery", "fuel", "location", "tires", "doors"}

	for _, name := range subcommands {
		t.Run(name, func(t *testing.T) {
			cmd := NewStatusCmd()
			subCmd := findSubcommand(cmd, name)

			if subCmd == nil {
				t.Fatalf("Expected %s subcommand to exist", name)
			}

			if subCmd.Short == "" {
				t.Errorf("Expected %s subcommand to have a description", name)
			}
		})
	}
}

// TestStatusCommand_JSONFlag tests the JSON output flag
func TestStatusCommand_JSONFlag(t *testing.T) {
	cmd := NewStatusCmd()

	// Check if json flag exists
	jsonFlag := cmd.PersistentFlags().Lookup("json")
	if jsonFlag == nil {
		t.Fatal("Expected --json flag to exist")
	}

	if jsonFlag.Value.Type() != "bool" {
		t.Errorf("Expected --json flag to be bool, got %s", jsonFlag.Value.Type())
	}
}

// TestFormatBatteryStatus tests battery status formatting
func TestFormatBatteryStatus(t *testing.T) {
	tests := []struct {
		name             string
		batteryLevel     float64
		range_           float64
		chargeTimeACMin  float64
		chargeTimeQBCMin float64
		pluggedIn        bool
		charging         bool
		expectedOutput   string
	}{
		{
			name:             "charging with time estimates",
			batteryLevel:     66,
			range_:           245.5,
			chargeTimeACMin:  180,
			chargeTimeQBCMin: 45,
			pluggedIn:        true,
			charging:         true,
			expectedOutput:   "BATTERY: 66% (245.5 km range) [charging, ~45m quick / ~3h AC]",
		},
		{
			name:             "charging with only AC time",
			batteryLevel:     50,
			range_:           150.0,
			chargeTimeACMin:  150,
			chargeTimeQBCMin: 0,
			pluggedIn:        true,
			charging:         true,
			expectedOutput:   "BATTERY: 50% (150.0 km range) [charging, ~2h 30m to full]",
		},
		{
			name:             "charging with no time estimates",
			batteryLevel:     45,
			range_:           120.0,
			chargeTimeACMin:  0,
			chargeTimeQBCMin: 0,
			pluggedIn:        true,
			charging:         true,
			expectedOutput:   "BATTERY: 45% (120.0 km range) [charging]",
		},
		{
			name:             "plugged not charging",
			batteryLevel:     100,
			range_:           300.0,
			chargeTimeACMin:  0,
			chargeTimeQBCMin: 0,
			pluggedIn:        true,
			charging:         false,
			expectedOutput:   "BATTERY: 100% (300.0 km range) [plugged in, not charging]",
		},
		{
			name:             "unplugged",
			batteryLevel:     50,
			range_:           150.0,
			chargeTimeACMin:  0,
			chargeTimeQBCMin: 0,
			pluggedIn:        false,
			charging:         false,
			expectedOutput:   "BATTERY: 50% (150.0 km range)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batteryInfo := api.BatteryInfo{
				BatteryLevel:     tt.batteryLevel,
				RangeKm:          tt.range_,
				ChargeTimeACMin:  tt.chargeTimeACMin,
				ChargeTimeQBCMin: tt.chargeTimeQBCMin,
				PluggedIn:        tt.pluggedIn,
				Charging:         tt.charging,
				HeaterOn:         false,
				HeaterAuto:       false,
			}
			result, err := formatBatteryStatus(batteryInfo, false)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result != tt.expectedOutput {
				t.Errorf("Expected '%s', got '%s'", tt.expectedOutput, result)
			}
		})
	}
}

// TestFormatBatteryStatus_JSON tests battery status JSON formatting
func TestFormatBatteryStatus_JSON(t *testing.T) {
	batteryInfo := api.BatteryInfo{
		BatteryLevel:     66,
		RangeKm:          245.5,
		ChargeTimeACMin:  180,
		ChargeTimeQBCMin: 45,
		PluggedIn:        true,
		Charging:         true,
		HeaterOn:         false,
		HeaterAuto:       false,
	}
	result, err := formatBatteryStatus(batteryInfo, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	data := parseJSONToMap(t, result)
	assertMapValue(t, data, "battery_level", float64(66))
	assertMapValue(t, data, "range_km", 245.5)
	assertMapValue(t, data, "plugged_in", true)
	assertMapValue(t, data, "charging", true)
	assertMapValue(t, data, "charge_time_ac_minutes", float64(180))
	assertMapValue(t, data, "charge_time_qbc_minutes", float64(45))
}

// TestFormatBatteryStatus_WithHeater tests battery heater display
func TestFormatBatteryStatus_WithHeater(t *testing.T) {
	tests := []struct {
		name       string
		heaterOn   bool
		heaterAuto bool
		expected   string
	}{
		{
			name:       "heater on with auto",
			heaterOn:   true,
			heaterAuto: true,
			expected:   "BATTERY: 66% (245.5 km range) [battery heater on, auto enabled]",
		},
		{
			name:       "heater on without auto",
			heaterOn:   true,
			heaterAuto: false,
			expected:   "BATTERY: 66% (245.5 km range) [battery heater on]",
		},
		{
			name:       "heater off with auto enabled",
			heaterOn:   false,
			heaterAuto: true,
			expected:   "BATTERY: 66% (245.5 km range) [battery heater auto enabled]",
		},
		{
			name:       "heater off without auto",
			heaterOn:   false,
			heaterAuto: false,
			expected:   "BATTERY: 66% (245.5 km range)",
		},
		{
			name:       "charging with heater on",
			heaterOn:   true,
			heaterAuto: true,
			expected:   "BATTERY: 66% (245.5 km range) [charging, ~45m quick / ~3h AC, battery heater on, auto enabled]",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var batteryInfo api.BatteryInfo
			if i == 4 {
				// Last test case includes charging
				batteryInfo = api.BatteryInfo{
					BatteryLevel:     66,
					RangeKm:          245.5,
					ChargeTimeACMin:  180,
					ChargeTimeQBCMin: 45,
					PluggedIn:        true,
					Charging:         true,
					HeaterOn:         tt.heaterOn,
					HeaterAuto:       tt.heaterAuto,
				}
			} else {
				batteryInfo = api.BatteryInfo{
					BatteryLevel:     66,
					RangeKm:          245.5,
					ChargeTimeACMin:  0,
					ChargeTimeQBCMin: 0,
					PluggedIn:        false,
					Charging:         false,
					HeaterOn:         tt.heaterOn,
					HeaterAuto:       tt.heaterAuto,
				}
			}
			result, err := formatBatteryStatus(batteryInfo, false)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestFormatFuelStatus tests fuel status formatting
func TestFormatFuelStatus(t *testing.T) {
	fuelInfo := api.FuelInfo{
		FuelLevel: 92,
		RangeKm:   630.0,
	}
	result, err := formatFuelStatus(fuelInfo, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected := "FUEL: 92% (630.0 km range)"

	if !strings.Contains(result, expected) {
		t.Errorf("Expected output to contain '%s', got '%s'", expected, result)
	}
}

// TestFormatFuelStatus_JSON tests fuel status JSON formatting
func TestFormatFuelStatus_JSON(t *testing.T) {
	fuelInfo := api.FuelInfo{
		FuelLevel: 92,
		RangeKm:   630.0,
	}
	result, err := formatFuelStatus(fuelInfo, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	data := parseJSONToMap(t, result)
	assertMapValue(t, data, "fuel_level", float64(92))
	assertMapValue(t, data, "range_km", 630.0)
}

// TestFormatDoorsStatus tests doors status formatting
func TestFormatDoorsStatus(t *testing.T) {
	tests := []struct {
		name           string
		doorStatus     api.DoorStatus
		expectedOutput string
	}{
		{
			name: "all locked and closed",
			doorStatus: api.DoorStatus{
				DriverOpen:      false,
				PassengerOpen:   false,
				RearLeftOpen:    false,
				RearRightOpen:   false,
				TrunkOpen:       false,
				HoodOpen:        false,
				DriverLocked:    true,
				PassengerLocked: true,
				RearLeftLocked:  true,
				RearRightLocked: true,
				AllLocked:       true,
			},
			expectedOutput: "DOORS: All locked",
		},
		{
			name: "driver door unlocked",
			doorStatus: api.DoorStatus{
				DriverOpen:      false,
				PassengerOpen:   false,
				RearLeftOpen:    false,
				RearRightOpen:   false,
				TrunkOpen:       false,
				HoodOpen:        false,
				DriverLocked:    false,
				PassengerLocked: true,
				RearLeftLocked:  true,
				RearRightLocked: true,
				AllLocked:       false,
			},
			expectedOutput: "DOORS: Driver unlocked",
		},
		{
			name: "trunk and hood open",
			doorStatus: api.DoorStatus{
				DriverOpen:      false,
				PassengerOpen:   false,
				RearLeftOpen:    false,
				RearRightOpen:   false,
				TrunkOpen:       true,
				HoodOpen:        true,
				DriverLocked:    true,
				PassengerLocked: true,
				RearLeftLocked:  true,
				RearRightLocked: true,
				AllLocked:       false,
			},
			expectedOutput: "DOORS: Trunk open, Hood open",
		},
		{
			name: "multiple issues",
			doorStatus: api.DoorStatus{
				DriverOpen:      false,
				PassengerOpen:   true,
				RearLeftOpen:    false,
				RearRightOpen:   false,
				TrunkOpen:       true,
				HoodOpen:        false,
				DriverLocked:    false,
				PassengerLocked: true,
				RearLeftLocked:  true,
				RearRightLocked: true,
				AllLocked:       false,
			},
			expectedOutput: "DOORS: Driver unlocked, Passenger open, Trunk open",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := formatDoorsStatus(tt.doorStatus, false)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result != tt.expectedOutput {
				t.Errorf("Expected '%s', got '%s'", tt.expectedOutput, result)
			}
		})
	}
}

// TestFormatTiresStatus tests tire status formatting
func TestFormatTiresStatus(t *testing.T) {
	tireInfo := api.TireInfo{
		FrontLeftPsi:  32.5,
		FrontRightPsi: 32.0,
		RearLeftPsi:   31.5,
		RearRightPsi:  31.8,
	}
	result, err := formatTiresStatus(tireInfo, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	expected := "TIRES: FL:32.5 FR:32.0 RL:31.5 RR:31.8 PSI"

	if !strings.Contains(result, expected) {
		t.Errorf("Expected output to contain '%s', got '%s'", expected, result)
	}
}

// TestFormatLocationStatus tests location status formatting
func TestFormatLocationStatus(t *testing.T) {
	locationInfo := api.LocationInfo{
		Latitude:  37.7749,
		Longitude: 122.4194,
		Timestamp: "20231201120000",
	}
	result, err := formatLocationStatus(locationInfo, false)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(result, "LOCATION:") {
		t.Error("Expected output to contain 'LOCATION:'")
	}

	if !strings.Contains(result, "37.7749") {
		t.Error("Expected output to contain latitude")
	}

	if !strings.Contains(result, "122.4194") {
		t.Error("Expected output to contain longitude")
	}
}

// TestGetInternalVIN tests getting internal VIN from vehicle base info
func TestGetInternalVIN(t *testing.T) {
	vecBaseInfos := &api.VecBaseInfosResponse{
		ResultCode: api.ResultCodeSuccess,
		VecBaseInfos: []api.VecBaseInfo{
			{
				Vehicle: api.Vehicle{
					CvInformation: api.CvInformation{
						InternalVIN: "INTERNAL123",
					},
				},
			},
		},
	}

	vin, err := vecBaseInfos.GetInternalVIN()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if vin != "INTERNAL123" {
		t.Errorf("Expected VIN 'INTERNAL123', got '%s'", vin)
	}
}

// TestGetInternalVIN_NoVehicles tests error when no vehicles found
func TestGetInternalVIN_NoVehicles(t *testing.T) {
	vecBaseInfos := &api.VecBaseInfosResponse{
		ResultCode:   "ResultCodeSuccess",
		VecBaseInfos: []api.VecBaseInfo{},
	}

	_, err := vecBaseInfos.GetInternalVIN()
	if err == nil {
		t.Fatal("Expected error for no vehicles, got nil")
	}
}

// TestRunStatus_Integration tests the full status command integration
func TestRunStatus_Integration(t *testing.T) {
	// This would require mocking the API client
	// For now, we just test that the function signature is correct
	cmd := NewStatusCmd()

	// Test that we can execute the command (it will fail due to missing config, but that's expected)
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// We don't execute it here because it would need a real API client
	// The actual execution tests would be in integration tests
}

// TestFormatHvacStatus tests HVAC status formatting
func TestFormatHvacStatus(t *testing.T) {
	tests := []struct {
		name           string
		hvacOn         bool
		frontDefroster bool
		rearDefroster  bool
		interiorTempC  float64
		targetTempC    float64
		expectedOutput string
	}{
		{
			name:           "hvac on with both defrosters",
			hvacOn:         true,
			frontDefroster: true,
			rearDefroster:  true,
			interiorTempC:  21,
			targetTempC:    21,
			expectedOutput: "CLIMATE: On, 21°C (front and rear defrosters on)",
		},
		{
			name:           "hvac on with front defroster only",
			hvacOn:         true,
			frontDefroster: true,
			rearDefroster:  false,
			interiorTempC:  19,
			targetTempC:    19,
			expectedOutput: "CLIMATE: On, 19°C (front defroster on)",
		},
		{
			name:           "hvac on with rear defroster only",
			hvacOn:         true,
			frontDefroster: false,
			rearDefroster:  true,
			interiorTempC:  22,
			targetTempC:    22,
			expectedOutput: "CLIMATE: On, 22°C (rear defroster on)",
		},
		{
			name:           "hvac on no defrosters",
			hvacOn:         true,
			frontDefroster: false,
			rearDefroster:  false,
			interiorTempC:  20,
			targetTempC:    20,
			expectedOutput: "CLIMATE: On, 20°C",
		},
		{
			name:           "hvac on with target temp different from interior",
			hvacOn:         true,
			frontDefroster: false,
			rearDefroster:  false,
			interiorTempC:  18,
			targetTempC:    22,
			expectedOutput: "CLIMATE: On, 18°C → 22°C",
		},
		{
			name:           "hvac off",
			hvacOn:         false,
			frontDefroster: false,
			rearDefroster:  false,
			interiorTempC:  15,
			targetTempC:    20,
			expectedOutput: "CLIMATE: Off, 15°C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hvacInfo := api.HVACInfo{
				HVACOn:         tt.hvacOn,
				FrontDefroster: tt.frontDefroster,
				RearDefroster:  tt.rearDefroster,
				InteriorTempC:  tt.interiorTempC,
				TargetTempC:    tt.targetTempC,
			}
			result, err := formatHvacStatus(hvacInfo, false)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result != tt.expectedOutput {
				t.Errorf("Expected '%s', got '%s'", tt.expectedOutput, result)
			}
		})
	}
}

// TestFormatHvacStatus_JSON tests HVAC status JSON formatting
func TestFormatHvacStatus_JSON(t *testing.T) {
	hvacInfo := api.HVACInfo{
		HVACOn:         true,
		FrontDefroster: true,
		RearDefroster:  false,
		InteriorTempC:  21,
		TargetTempC:    22,
	}
	result, err := formatHvacStatus(hvacInfo, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	data := parseJSONToMap(t, result)
	assertMapValue(t, data, "hvac_on", true)
	assertMapValue(t, data, "front_defroster", true)
	assertMapValue(t, data, "rear_defroster", false)
	assertMapValue(t, data, "interior_temperature_c", float64(21))
	assertMapValue(t, data, "target_temperature_c", float64(22))
}

// TestGetHvacInfo tests extracting HVAC info from EV status
func TestGetHvacInfo(t *testing.T) {
	evStatus := &api.EVVehicleStatusResponse{
		ResultCode: api.ResultCodeSuccess,
		ResultData: []api.EVResultData{
			{
				OccurrenceDate: "20231201120000",
				PlusBInformation: api.PlusBInformation{
					VehicleInfo: api.EVVehicleInfo{
						ChargeInfo: api.ChargeInfo{},
						RemoteHvacInfo: &api.RemoteHvacInfo{
							HVAC:           1,
							FrontDefroster: 1,
							RearDefogger:   0,
							InCarTeDC:      21.5,
							TargetTemp:     22.0,
						},
					},
				},
			},
		},
	}

	hvacInfo, err := evStatus.GetHvacInfo()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !hvacInfo.HVACOn {
		t.Error("Expected HVACOn to be true")
	}
	if !hvacInfo.FrontDefroster {
		t.Error("Expected FrontDefroster to be true")
	}
	if hvacInfo.RearDefroster {
		t.Error("Expected RearDefroster to be false")
	}
	if hvacInfo.InteriorTempC != 21.5 {
		t.Errorf("Expected InteriorTempC 21.5, got %v", hvacInfo.InteriorTempC)
	}
	if hvacInfo.TargetTempC != 22.0 {
		t.Errorf("Expected TargetTempC 22.0, got %v", hvacInfo.TargetTempC)
	}
}

// TestGetHvacInfo_MissingData tests extracting HVAC info when data is missing
func TestGetHvacInfo_MissingData(t *testing.T) {
	evStatus := &api.EVVehicleStatusResponse{
		ResultCode: api.ResultCodeSuccess,
		ResultData: []api.EVResultData{
			{
				OccurrenceDate: "20231201120000",
				PlusBInformation: api.PlusBInformation{
					VehicleInfo: api.EVVehicleInfo{
						ChargeInfo:     api.ChargeInfo{},
						RemoteHvacInfo: nil, // No HVAC info
					},
				},
			},
		},
	}

	_, err := evStatus.GetHvacInfo()
	if err == nil {
		t.Fatal("Expected error when HVAC info is missing, got nil")
	}
}

// TestExtractHvacData tests extracting HVAC data for JSON output
func TestExtractHvacData(t *testing.T) {
	evStatus := &api.EVVehicleStatusResponse{
		ResultCode: api.ResultCodeSuccess,
		ResultData: []api.EVResultData{
			{
				OccurrenceDate: "20231201120000",
				PlusBInformation: api.PlusBInformation{
					VehicleInfo: api.EVVehicleInfo{
						ChargeInfo: api.ChargeInfo{},
						RemoteHvacInfo: &api.RemoteHvacInfo{
							HVAC:           1,
							FrontDefroster: 0,
							RearDefogger:   1,
							InCarTeDC:      18,
							TargetTemp:     22,
						},
					},
				},
			},
		},
	}

	data := extractHvacData(evStatus)

	assertMapValue(t, data, "hvac_on", true)
	assertMapValue(t, data, "front_defroster", false)
	assertMapValue(t, data, "rear_defroster", true)
	assertMapValue(t, data, "interior_temperature_c", float64(18))
	assertMapValue(t, data, "target_temperature_c", float64(22))
}

// TestFormatOdometerStatus tests odometer status formatting
func TestFormatOdometerStatus(t *testing.T) {
	tests := []struct {
		name           string
		odometerKm     float64
		expectedOutput string
	}{
		{
			name:           "typical odometer",
			odometerKm:     12345.6,
			expectedOutput: "ODOMETER: 12,345.6 km",
		},
		{
			name:           "high odometer",
			odometerKm:     99999.9,
			expectedOutput: "ODOMETER: 99,999.9 km",
		},
		{
			name:           "low odometer",
			odometerKm:     123.4,
			expectedOutput: "ODOMETER: 123.4 km",
		},
		{
			name:           "zero odometer",
			odometerKm:     0,
			expectedOutput: "ODOMETER: 0.0 km",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			odometerInfo := api.OdometerInfo{OdometerKm: tt.odometerKm}
			result, err := formatOdometerStatus(odometerInfo, false)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result != tt.expectedOutput {
				t.Errorf("Expected '%s', got '%s'", tt.expectedOutput, result)
			}
		})
	}
}

// TestFormatOdometerStatus_JSON tests odometer status JSON formatting
func TestFormatOdometerStatus_JSON(t *testing.T) {
	odometerInfo := api.OdometerInfo{OdometerKm: 12345.6}
	result, err := formatOdometerStatus(odometerInfo, true)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	data := parseJSONToMap(t, result)
	assertMapValue(t, data, "odometer_km", 12345.6)
}

// TestGetOdometerInfo tests extracting odometer info from vehicle status
func TestGetOdometerInfo(t *testing.T) {
	vehicleStatus := &api.VehicleStatusResponse{
		ResultCode: api.ResultCodeSuccess,
		RemoteInfos: []api.RemoteInfo{
			{
				DriveInformation: api.DriveInformation{
					OdoDispValue: 12345.6,
				},
			},
		},
	}

	odometerInfo, err := vehicleStatus.GetOdometerInfo()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if odometerInfo.OdometerKm != 12345.6 {
		t.Errorf("Expected odometer 12345.6, got %v", odometerInfo.OdometerKm)
	}
}

// TestExtractOdometerData tests extracting odometer data for JSON output
func TestExtractOdometerData(t *testing.T) {
	vehicleStatus := &api.VehicleStatusResponse{
		ResultCode: api.ResultCodeSuccess,
		RemoteInfos: []api.RemoteInfo{
			{
				DriveInformation: api.DriveInformation{
					OdoDispValue: 12345.6,
				},
			},
		},
	}

	data := extractOdometerData(vehicleStatus)

	assertMapValue(t, data, "odometer_km", 12345.6)
}

// TestFormatChargeTime tests charge time formatting
func TestFormatChargeTime(t *testing.T) {
	tests := []struct {
		name       string
		acMinutes  float64
		qbcMinutes float64
		expected   string
	}{
		{
			name:       "both available and different",
			acMinutes:  180,
			qbcMinutes: 45,
			expected:   "~45m quick / ~3h AC",
		},
		{
			name:       "only AC available",
			acMinutes:  150,
			qbcMinutes: 0,
			expected:   "~2h 30m to full",
		},
		{
			name:       "only QBC available",
			acMinutes:  0,
			qbcMinutes: 60,
			expected:   "~1h to full",
		},
		{
			name:       "both zero",
			acMinutes:  0,
			qbcMinutes: 0,
			expected:   "",
		},
		{
			name:       "short time",
			acMinutes:  30,
			qbcMinutes: 0,
			expected:   "~30m to full",
		},
		{
			name:       "exact hour",
			acMinutes:  120,
			qbcMinutes: 0,
			expected:   "~2h to full",
		},
		{
			name:       "same time for both",
			acMinutes:  60,
			qbcMinutes: 60,
			expected:   "~1h to full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatChargeTime(tt.acMinutes, tt.qbcMinutes)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestFormatVehicleHeader tests vehicle header formatting
func TestFormatVehicleHeader(t *testing.T) {
	tests := []struct {
		name     string
		info     VehicleInfo
		expected string
	}{
		{
			name: "full info with nickname",
			info: VehicleInfo{
				VIN:       "JM3KKEHC1R0123456",
				Nickname:  "My CX-90",
				ModelName: "CX-90 PHEV",
				ModelYear: "2024",
			},
			expected: "CX-90 PHEV (2024) \"My CX-90\"\nVIN: JM3KKEHC1R0123456\n",
		},
		{
			name: "model without nickname",
			info: VehicleInfo{
				VIN:       "JM3KKEHC1R0123456",
				ModelName: "CX-90 PHEV",
				ModelYear: "2024",
			},
			expected: "CX-90 PHEV (2024)\nVIN: JM3KKEHC1R0123456\n",
		},
		{
			name: "model without year",
			info: VehicleInfo{
				VIN:       "JM3KKEHC1R0123456",
				ModelName: "CX-90 PHEV",
			},
			expected: "CX-90 PHEV\nVIN: JM3KKEHC1R0123456\n",
		},
		{
			name: "only nickname",
			info: VehicleInfo{
				VIN:      "JM3KKEHC1R0123456",
				Nickname: "My Car",
			},
			expected: "\"My Car\"\nVIN: JM3KKEHC1R0123456\n",
		},
		{
			name: "only VIN",
			info: VehicleInfo{
				VIN: "JM3KKEHC1R0123456",
			},
			expected: "VIN: JM3KKEHC1R0123456\n",
		},
		{
			name:     "empty info",
			info:     VehicleInfo{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatVehicleHeader(tt.info)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestExtractVehicleInfoData tests vehicle info extraction for JSON
func TestExtractVehicleInfoData(t *testing.T) {
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

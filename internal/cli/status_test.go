package cli

import (
	"bytes"
	"testing"
	"time"

	"github.com/cv/mcs/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStatusCommand tests the status command
func TestStatusCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		expectedUse   string
		checkShortSet bool
	}{
		{
			name:          "command has correct use",
			expectedUse:   "status",
			checkShortSet: false,
		},
		{
			name:          "command has short description",
			expectedUse:   "status",
			checkShortSet: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewStatusCmd()

			assert.Equal(t, tt.expectedUse, cmd.Use)

			assert.False(t, tt.checkShortSet && cmd.Short == "", "Expected Short description to be set")
		})
	}
}

// TestStatusCommand_NoSubcommand tests status command without subcommand
func TestStatusCommand_NoSubcommand(t *testing.T) {
	t.Parallel()
	// This should show all status information
	cmd := NewStatusCmd()
	cmd.SetArgs([]string{})

	// We need to inject a mock client - this will be handled in the actual implementation
	// For now, we test that the command structure is correct
	err := cmd.ValidateArgs([]string{})
	require.NoErrorf(t, err, "Status command should accept no arguments: %v", err)

}

// TestStatusCommand_Subcommands tests all status subcommands using table-driven pattern
func TestStatusCommand_Subcommands(t *testing.T) {
	t.Parallel()
	subcommands := []string{"battery", "fuel", "location", "tires", "doors"}

	for _, name := range subcommands {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cmd := NewStatusCmd()
			subCmd := findSubcommand(cmd, name)

			require.NotNilf(t, subCmd, "Expected %s subcommand to exist", name)

			assert.NotEmptyf(t, subCmd.Short, "Expected %s subcommand to have a description", name)
		})
	}
}

// TestStatusCommand_JSONFlag tests the JSON output flag
func TestStatusCommand_JSONFlag(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		flagName     string
		expectedType string
		shouldExist  bool
	}{
		{
			name:         "json flag exists",
			flagName:     "json",
			expectedType: "bool",
			shouldExist:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd := NewStatusCmd()
			flag := cmd.PersistentFlags().Lookup(tt.flagName)

			if tt.shouldExist {
				require.NotNil(t, flag)
			}

			if flag != nil {
				assert.Equal(t, tt.expectedType, flag.Value.Type())
			}

		})
	}
}

// TestFormatBatteryStatus tests battery status formatting
func TestFormatBatteryStatus(t *testing.T) {
	t.Parallel()
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
			expectedOutput:   "BATTERY: [██████░░░░] 66% (245.5 km range) [charging, ~45m quick / ~3h AC]",
		},
		{
			name:             "charging with only AC time",
			batteryLevel:     50,
			range_:           150.0,
			chargeTimeACMin:  150,
			chargeTimeQBCMin: 0,
			pluggedIn:        true,
			charging:         true,
			expectedOutput:   "BATTERY: [█████░░░░░] 50% (150.0 km range) [charging, ~2h 30m to full]",
		},
		{
			name:             "charging with no time estimates",
			batteryLevel:     45,
			range_:           120.0,
			chargeTimeACMin:  0,
			chargeTimeQBCMin: 0,
			pluggedIn:        true,
			charging:         true,
			expectedOutput:   "BATTERY: [████░░░░░░] 45% (120.0 km range) [charging]",
		},
		{
			name:             "plugged not charging",
			batteryLevel:     100,
			range_:           300.0,
			chargeTimeACMin:  0,
			chargeTimeQBCMin: 0,
			pluggedIn:        true,
			charging:         false,
			expectedOutput:   "BATTERY: [██████████] 100% (300.0 km range) [plugged in, not charging]",
		},
		{
			name:             "unplugged",
			batteryLevel:     50,
			range_:           150.0,
			chargeTimeACMin:  0,
			chargeTimeQBCMin: 0,
			pluggedIn:        false,
			charging:         false,
			expectedOutput:   "BATTERY: [█████░░░░░] 50% (150.0 km range)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
			require.NoError(t, err, "Unexpected error: %v")
			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}

// TestFormatBatteryStatus_JSON tests battery status JSON formatting
func TestFormatBatteryStatus_JSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		batteryInfo  api.BatteryInfo
		expectedJSON map[string]any
	}{
		{
			name: "battery status JSON format",
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
			expectedJSON: map[string]any{
				"battery_level":           float64(66),
				"range_km":                245.5,
				"plugged_in":              true,
				"charging":                true,
				"charge_time_ac_minutes":  float64(180),
				"charge_time_qbc_minutes": float64(45),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := formatBatteryStatus(tt.batteryInfo, true)
			require.NoError(t, err, "Unexpected error: %v")

			data := parseJSONToMap(t, result)
			for key, expected := range tt.expectedJSON {
				assertMapValue(t, data, key, expected)
			}
		})
	}
}

// TestFormatBatteryStatus_WithHeater tests battery heater display
func TestFormatBatteryStatus_WithHeater(t *testing.T) {
	t.Parallel()
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
			expected:   "BATTERY: [██████░░░░] 66% (245.5 km range) [battery heater on, auto enabled]",
		},
		{
			name:       "heater on without auto",
			heaterOn:   true,
			heaterAuto: false,
			expected:   "BATTERY: [██████░░░░] 66% (245.5 km range) [battery heater on]",
		},
		{
			name:       "heater off with auto enabled",
			heaterOn:   false,
			heaterAuto: true,
			expected:   "BATTERY: [██████░░░░] 66% (245.5 km range) [battery heater auto enabled]",
		},
		{
			name:       "heater off without auto",
			heaterOn:   false,
			heaterAuto: false,
			expected:   "BATTERY: [██████░░░░] 66% (245.5 km range)",
		},
		{
			name:       "charging with heater on",
			heaterOn:   true,
			heaterAuto: true,
			expected:   "BATTERY: [██████░░░░] 66% (245.5 km range) [charging, ~45m quick / ~3h AC, battery heater on, auto enabled]",
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
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
			require.NoError(t, err, "Unexpected error: %v")
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatFuelStatus tests fuel status formatting
func TestFormatFuelStatus(t *testing.T) {
	t.Parallel()
	colorTestMutex.Lock()
	defer colorTestMutex.Unlock()
	SetColorEnabled(false)

	tests := []struct {
		name           string
		fuelLevel      float64
		rangeKm        float64
		asJSON         bool
		expectedOutput string
		expectedJSON   map[string]any
	}{
		{
			name:           "fuel status text format",
			fuelLevel:      92,
			rangeKm:        630.0,
			asJSON:         false,
			expectedOutput: "FUEL: [█████████░] 92% (630.0 km range)",
		},
		{
			name:      "fuel status JSON format",
			fuelLevel: 92,
			rangeKm:   630.0,
			asJSON:    true,
			expectedJSON: map[string]any{
				"fuel_level": float64(92),
				"range_km":   630.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fuelInfo := api.FuelInfo{
				FuelLevel: tt.fuelLevel,
				RangeKm:   tt.rangeKm,
			}
			result, err := formatFuelStatus(fuelInfo, tt.asJSON)
			require.NoError(t, err, "Unexpected error: %v")

			if tt.asJSON {
				data := parseJSONToMap(t, result)
				for key, expected := range tt.expectedJSON {
					assertMapValue(t, data, key, expected)
				}
			} else {
				assert.Equal(t, tt.expectedOutput, result)
			}
		})
	}
}

// TestFormatDoorsStatus tests doors status formatting
func TestFormatDoorsStatus(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			result, err := formatDoorsStatus(tt.doorStatus, false)
			require.NoError(t, err, "Unexpected error: %v")
			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}

// TestFormatTiresStatus tests tire status formatting
func TestFormatTiresStatus(t *testing.T) {
	t.Parallel()
	colorTestMutex.Lock()
	defer colorTestMutex.Unlock()
	SetColorEnabled(false)

	tests := []struct {
		name          string
		frontLeftPsi  float64
		frontRightPsi float64
		rearLeftPsi   float64
		rearRightPsi  float64
		expectedPart  string
	}{
		{
			name:          "typical tire pressures",
			frontLeftPsi:  32.5,
			frontRightPsi: 32.0,
			rearLeftPsi:   31.5,
			rearRightPsi:  31.8,
			expectedPart:  "TIRES: FL:32.5 FR:32.0 RL:31.5 RR:31.8 PSI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tireInfo := api.TireInfo{
				FrontLeftPsi:  tt.frontLeftPsi,
				FrontRightPsi: tt.frontRightPsi,
				RearLeftPsi:   tt.rearLeftPsi,
				RearRightPsi:  tt.rearRightPsi,
			}
			result, err := formatTiresStatus(tireInfo, false)
			require.NoError(t, err, "Unexpected error: %v")

			assert.Contains(t, result, tt.expectedPart)
		})
	}
}

// TestFormatLocationStatus tests location status formatting
func TestFormatLocationStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		latitude         float64
		longitude        float64
		timestamp        string
		expectedContains []string
	}{
		{
			name:      "location with coordinates",
			latitude:  37.7749,
			longitude: 122.4194,
			timestamp: "20231201120000",
			expectedContains: []string{
				"LOCATION:",
				"37.7749",
				"122.4194",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			locationInfo := api.LocationInfo{
				Latitude:  tt.latitude,
				Longitude: tt.longitude,
				Timestamp: tt.timestamp,
			}
			result, err := formatLocationStatus(locationInfo, false)
			require.NoError(t, err, "Unexpected error: %v")

			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

// TestGetInternalVIN tests getting internal VIN from vehicle base info
func TestGetInternalVIN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		response    *api.VecBaseInfosResponse
		expectedVIN string
		expectError bool
	}{
		{
			name: "successful VIN retrieval",
			response: &api.VecBaseInfosResponse{
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
			},
			expectedVIN: "INTERNAL123",
			expectError: false,
		},
		{
			name: "no vehicles found",
			response: &api.VecBaseInfosResponse{
				ResultCode:   "ResultCodeSuccess",
				VecBaseInfos: []api.VecBaseInfo{},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			vin, err := tt.response.GetInternalVIN()

			if tt.expectError {
				require.Error(t, err, "Expected error, got nil")
			} else {
				require.NoError(t, err, "Expected no error, got: %v")
				assert.Equal(t, tt.expectedVIN, vin)
			}
		})
	}
}

// TestRunStatus_Integration tests the full status command integration
func TestRunStatus_Integration(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
			hvacInfo := api.HVACInfo{
				HVACOn:         tt.hvacOn,
				FrontDefroster: tt.frontDefroster,
				RearDefroster:  tt.rearDefroster,
				InteriorTempC:  tt.interiorTempC,
				TargetTempC:    tt.targetTempC,
			}
			result, err := formatHvacStatus(hvacInfo, false)
			require.NoError(t, err, "Unexpected error: %v")
			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}

// TestFormatHvacStatus_JSON tests HVAC status JSON formatting
func TestFormatHvacStatus_JSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		hvacInfo     api.HVACInfo
		expectedJSON map[string]any
	}{
		{
			name: "HVAC status JSON format",
			hvacInfo: api.HVACInfo{
				HVACOn:         true,
				FrontDefroster: true,
				RearDefroster:  false,
				InteriorTempC:  21,
				TargetTempC:    22,
			},
			expectedJSON: map[string]any{
				"hvac_on":                true,
				"front_defroster":        true,
				"rear_defroster":         false,
				"interior_temperature_c": float64(21),
				"target_temperature_c":   float64(22),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := formatHvacStatus(tt.hvacInfo, true)
			require.NoError(t, err, "Unexpected error: %v")

			data := parseJSONToMap(t, result)
			for key, expected := range tt.expectedJSON {
				assertMapValue(t, data, key, expected)
			}
		})
	}
}

// TestGetHvacInfo tests extracting HVAC info from EV status
func TestGetHvacInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		response          *api.EVVehicleStatusResponse
		expectError       bool
		expectedHVACOn    bool
		expectedFrontDefr bool
		expectedRearDefr  bool
		expectedInteriorC float64
		expectedTargetC   float64
	}{
		{
			name: "successful HVAC info extraction",
			response: &api.EVVehicleStatusResponse{
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
			},
			expectError:       false,
			expectedHVACOn:    true,
			expectedFrontDefr: true,
			expectedRearDefr:  false,
			expectedInteriorC: 21.5,
			expectedTargetC:   22.0,
		},
		{
			name: "missing HVAC data",
			response: &api.EVVehicleStatusResponse{
				ResultCode: api.ResultCodeSuccess,
				ResultData: []api.EVResultData{
					{
						OccurrenceDate: "20231201120000",
						PlusBInformation: api.PlusBInformation{
							VehicleInfo: api.EVVehicleInfo{
								ChargeInfo:     api.ChargeInfo{},
								RemoteHvacInfo: nil,
							},
						},
					},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			hvacInfo, err := tt.response.GetHvacInfo()

			if tt.expectError {
				require.Error(t, err, "Expected error, got nil")
			} else {
				require.NoError(t, err, "Unexpected error: %v")
				assert.Equal(t, tt.expectedHVACOn, hvacInfo.HVACOn)
				assert.Equal(t, tt.expectedFrontDefr, hvacInfo.FrontDefroster)
				assert.Equal(t, tt.expectedRearDefr, hvacInfo.RearDefroster)
				assert.InDelta(t, tt.expectedInteriorC, hvacInfo.InteriorTempC, 0.0001)
				assert.InDelta(t, tt.expectedTargetC, hvacInfo.TargetTempC, 0.0001)
			}
		})
	}
}

// TestExtractHvacData tests extracting HVAC data for JSON output
func TestExtractHvacData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		response     *api.EVVehicleStatusResponse
		expectedData map[string]any
	}{
		{
			name: "HVAC data extraction",
			response: &api.EVVehicleStatusResponse{
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
			},
			expectedData: map[string]any{
				"hvac_on":                true,
				"front_defroster":        false,
				"rear_defroster":         true,
				"interior_temperature_c": float64(18),
				"target_temperature_c":   float64(22),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := extractHvacData(tt.response)

			for key, expected := range tt.expectedData {
				assertMapValue(t, data, key, expected)
			}
		})
	}
}

// TestFormatOdometerStatus tests odometer status formatting
func TestFormatOdometerStatus(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			odometerInfo := api.OdometerInfo{OdometerKm: tt.odometerKm}
			result, err := formatOdometerStatus(odometerInfo, false)
			require.NoError(t, err, "Unexpected error: %v")
			assert.Equal(t, tt.expectedOutput, result)
		})
	}
}

// TestFormatOdometerStatus_JSON tests odometer status JSON formatting
func TestFormatOdometerStatus_JSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		odometerKm   float64
		expectedJSON map[string]any
	}{
		{
			name:       "odometer JSON format",
			odometerKm: 12345.6,
			expectedJSON: map[string]any{
				"odometer_km": 12345.6,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			odometerInfo := api.OdometerInfo{OdometerKm: tt.odometerKm}
			result, err := formatOdometerStatus(odometerInfo, true)
			require.NoError(t, err, "Unexpected error: %v")

			data := parseJSONToMap(t, result)
			for key, expected := range tt.expectedJSON {
				assertMapValue(t, data, key, expected)
			}
		})
	}
}

// TestGetOdometerInfo tests extracting odometer info from vehicle status
func TestGetOdometerInfo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		response         *api.VehicleStatusResponse
		expectedOdometer float64
	}{
		{
			name: "successful odometer extraction",
			response: &api.VehicleStatusResponse{
				ResultCode: api.ResultCodeSuccess,
				RemoteInfos: []api.RemoteInfo{
					{
						DriveInformation: api.DriveInformation{
							OdoDispValue: 12345.6,
						},
					},
				},
			},
			expectedOdometer: 12345.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			odometerInfo, err := tt.response.GetOdometerInfo()
			require.NoError(t, err, "Expected no error, got: %v")

			assert.InDelta(t, tt.expectedOdometer, odometerInfo.OdometerKm, 0.0001)
		})
	}
}

// TestExtractOdometerData tests extracting odometer data for JSON output
func TestExtractOdometerData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		response     *api.VehicleStatusResponse
		expectedData map[string]any
	}{
		{
			name: "odometer data extraction",
			response: &api.VehicleStatusResponse{
				ResultCode: api.ResultCodeSuccess,
				RemoteInfos: []api.RemoteInfo{
					{
						DriveInformation: api.DriveInformation{
							OdoDispValue: 12345.6,
						},
					},
				},
			},
			expectedData: map[string]any{
				"odometer_km": 12345.6,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := extractOdometerData(tt.response)

			for key, expected := range tt.expectedData {
				assertMapValue(t, data, key, expected)
			}
		})
	}
}

// TestFormatChargeTime tests charge time formatting
func TestFormatChargeTime(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			result := formatChargeTime(tt.acMinutes, tt.qbcMinutes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatVehicleHeader tests vehicle header formatting
func TestFormatVehicleHeader(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			result := formatVehicleHeader(tt.info)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestExtractVehicleInfoData tests vehicle info extraction for JSON
func TestExtractVehicleInfoData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		vehicleInfo  VehicleInfo
		expectedData map[string]any
	}{
		{
			name: "complete vehicle info extraction",
			vehicleInfo: VehicleInfo{
				VIN:       "JM3KKEHC1R0123456",
				Nickname:  "My CX-90",
				ModelName: "CX-90 PHEV",
				ModelYear: "2024",
			},
			expectedData: map[string]any{
				"vin":        "JM3KKEHC1R0123456",
				"nickname":   "My CX-90",
				"model_name": "CX-90 PHEV",
				"model_year": "2024",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data := extractVehicleInfoData(tt.vehicleInfo)

			for key, expected := range tt.expectedData {
				assertMapValue(t, data, key, expected)
			}
		})
	}
}

// TestFormatRelativeTime tests the formatRelativeTime function
func TestFormatRelativeTime(t *testing.T) {
	t.Parallel()
	now := time.Now()

	tests := []struct {
		name     string
		time     time.Time
		expected string
	}{
		{
			name:     "just now for future times",
			time:     now.Add(5 * time.Minute),
			expected: "just now",
		},
		{
			name:     "seconds ago",
			time:     now.Add(-30 * time.Second),
			expected: "30 sec ago",
		},
		{
			name:     "one minute ago",
			time:     now.Add(-1 * time.Minute),
			expected: "1 min ago",
		},
		{
			name:     "multiple minutes ago",
			time:     now.Add(-45 * time.Minute),
			expected: "45 min ago",
		},
		{
			name:     "one hour ago",
			time:     now.Add(-1 * time.Hour),
			expected: "1 hour ago",
		},
		{
			name:     "multiple hours ago",
			time:     now.Add(-5 * time.Hour),
			expected: "5 hours ago",
		},
		{
			name:     "one day ago",
			time:     now.Add(-24 * time.Hour),
			expected: "1 day ago",
		},
		{
			name:     "multiple days ago",
			time:     now.Add(-72 * time.Hour),
			expected: "3 days ago",
		},
		{
			name:     "less than one second",
			time:     now.Add(-500 * time.Millisecond),
			expected: "0 sec ago",
		},
		{
			name:     "59 seconds",
			time:     now.Add(-59 * time.Second),
			expected: "59 sec ago",
		},
		{
			name:     "59 minutes",
			time:     now.Add(-59 * time.Minute),
			expected: "59 min ago",
		},
		{
			name:     "23 hours",
			time:     now.Add(-23 * time.Hour),
			expected: "23 hours ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatRelativeTime(tt.time)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatTimestamp tests the formatTimestamp function
func TestFormatTimestamp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		timestamp        string
		expectedContains []string
		expectedFormat   string
	}{
		{
			name:           "valid timestamp format",
			timestamp:      "20250115120000",
			expectedFormat: "2025-01-15 12:00:00",
		},
		{
			name:           "another valid timestamp",
			timestamp:      "20241225093045",
			expectedFormat: "2024-12-25 09:30:45",
		},
		{
			name:      "invalid timestamp length",
			timestamp: "2025011512",
			expectedContains: []string{
				"2025011512",
			},
		},
		{
			name:      "invalid timestamp format",
			timestamp: "not-a-timestamp",
			expectedContains: []string{
				"not-a-timestamp",
			},
		},
		{
			name:      "empty timestamp",
			timestamp: "",
			expectedContains: []string{
				"",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatTimestamp(tt.timestamp)

			if tt.expectedFormat != "" {
				assert.Contains(t, result, tt.expectedFormat)
				// Should also contain relative time in parentheses
				assert.Contains(t, result, "(")
				assert.Contains(t, result, ")")

			}

			for _, expected := range tt.expectedContains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

// TestFormatWindowsStatus tests the formatWindowsStatus function
func TestFormatWindowsStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		windowsInfo    api.WindowStatus
		jsonOutput     bool
		expectedOutput string
		expectJSON     bool
	}{
		{
			name: "all windows closed",
			windowsInfo: api.WindowStatus{
				DriverPosition:    api.WindowClosed,
				PassengerPosition: api.WindowClosed,
				RearLeftPosition:  api.WindowClosed,
				RearRightPosition: api.WindowClosed,
			},
			jsonOutput:     false,
			expectedOutput: "WINDOWS: All closed",
		},
		{
			name: "driver window partially open",
			windowsInfo: api.WindowStatus{
				DriverPosition:    25,
				PassengerPosition: api.WindowClosed,
				RearLeftPosition:  api.WindowClosed,
				RearRightPosition: api.WindowClosed,
			},
			jsonOutput:     false,
			expectedOutput: "WINDOWS: Driver 25%",
		},
		{
			name: "multiple windows open",
			windowsInfo: api.WindowStatus{
				DriverPosition:    50,
				PassengerPosition: 75,
				RearLeftPosition:  api.WindowClosed,
				RearRightPosition: api.WindowClosed,
			},
			jsonOutput:     false,
			expectedOutput: "WINDOWS: Driver 50%, Passenger 75%",
		},
		{
			name: "all windows fully open",
			windowsInfo: api.WindowStatus{
				DriverPosition:    api.WindowFullyOpen,
				PassengerPosition: api.WindowFullyOpen,
				RearLeftPosition:  api.WindowFullyOpen,
				RearRightPosition: api.WindowFullyOpen,
			},
			jsonOutput:     false,
			expectedOutput: "WINDOWS: Driver 100%, Passenger 100%, Rear left 100%, Rear right 100%",
		},
		{
			name: "rear windows open",
			windowsInfo: api.WindowStatus{
				DriverPosition:    api.WindowClosed,
				PassengerPosition: api.WindowClosed,
				RearLeftPosition:  30,
				RearRightPosition: 40,
			},
			jsonOutput:     false,
			expectedOutput: "WINDOWS: Rear left 30%, Rear right 40%",
		},
		{
			name: "JSON output",
			windowsInfo: api.WindowStatus{
				DriverPosition:    25,
				PassengerPosition: 50,
				RearLeftPosition:  75,
				RearRightPosition: 100,
			},
			jsonOutput: true,
			expectJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := formatWindowsStatus(tt.windowsInfo, tt.jsonOutput)
			require.NoError(t, err, "Unexpected error: %v")

			if tt.expectJSON {
				data := parseJSONToMap(t, result)
				assertMapValue(t, data, "driver_position", float64(tt.windowsInfo.DriverPosition))
				assertMapValue(t, data, "passenger_position", float64(tt.windowsInfo.PassengerPosition))
				assertMapValue(t, data, "rear_left_position", float64(tt.windowsInfo.RearLeftPosition))
				assertMapValue(t, data, "rear_right_position", float64(tt.windowsInfo.RearRightPosition))
			} else if tt.expectedOutput != "" {
				assert.Equal(t, tt.expectedOutput, result)
			}
		})
	}
}

// TestDisplayStatusWithVehicle tests the displayStatusWithVehicle function
func TestDisplayStatusWithVehicle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		statusType     StatusType
		vehicleStatus  *api.VehicleStatusResponse
		evStatus       *api.EVVehicleStatusResponse
		vehicleInfo    VehicleInfo
		jsonOutput     bool
		expectedOutput []string
	}{
		{
			name:          "battery status",
			statusType:    StatusBattery,
			vehicleStatus: NewMockVehicleStatus().Build(),
			evStatus:      NewMockEVVehicleStatus().Build(),
			vehicleInfo: VehicleInfo{
				VIN:       "JM3KKEHC1R0123456",
				ModelName: "CX-90 PHEV",
				ModelYear: "2024",
			},
			jsonOutput: false,
			expectedOutput: []string{
				"BATTERY:",
				"80%",
			},
		},
		{
			name:       "fuel status",
			statusType: StatusFuel,
			vehicleStatus: &api.VehicleStatusResponse{
				ResultCode: api.ResultCodeSuccess,
				RemoteInfos: []api.RemoteInfo{
					{
						ResidualFuel: api.ResidualFuel{
							FuelSegmentDActl:  92,
							RemDrvDistDActlKm: 630.0,
						},
					},
				},
			},
			evStatus: NewMockEVVehicleStatus().Build(),
			vehicleInfo: VehicleInfo{
				VIN: "JM3KKEHC1R0123456",
			},
			jsonOutput: false,
			expectedOutput: []string{
				"FUEL:",
				"92%",
			},
		},
		{
			name:       "location status",
			statusType: StatusLocation,
			vehicleStatus: &api.VehicleStatusResponse{
				ResultCode: api.ResultCodeSuccess,
				AlertInfos: []api.AlertInfo{
					{
						PositionInfo: api.PositionInfo{
							Latitude:  37.7749,
							Longitude: -122.4194,
						},
					},
				},
			},
			evStatus: NewMockEVVehicleStatus().Build(),
			vehicleInfo: VehicleInfo{
				VIN: "JM3KKEHC1R0123456",
			},
			jsonOutput: false,
			expectedOutput: []string{
				"LOCATION:",
				"37.774900",
				"-122.419400",
			},
		},
		{
			name:       "tires status",
			statusType: StatusTires,
			vehicleStatus: &api.VehicleStatusResponse{
				ResultCode: api.ResultCodeSuccess,
				RemoteInfos: []api.RemoteInfo{
					{
						TPMSInformation: api.TPMSInformation{
							FLTPrsDispPsi: 32.0,
							FRTPrsDispPsi: 32.0,
							RLTPrsDispPsi: 31.0,
							RRTPrsDispPsi: 31.0,
						},
					},
				},
			},
			evStatus: NewMockEVVehicleStatus().Build(),
			vehicleInfo: VehicleInfo{
				VIN: "JM3KKEHC1R0123456",
			},
			jsonOutput: false,
			expectedOutput: []string{
				"TIRES:",
				"FL:",
				"FR:",
				"RL:",
				"RR:",
				"PSI",
			},
		},
		{
			name:       "doors status",
			statusType: StatusDoors,
			vehicleStatus: NewMockVehicleStatus().WithDoorStatus(api.DoorStatus{
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
			}).Build(),
			evStatus: NewMockEVVehicleStatus().Build(),
			vehicleInfo: VehicleInfo{
				VIN: "JM3KKEHC1R0123456",
			},
			jsonOutput: false,
			expectedOutput: []string{
				"DOORS:",
				"All locked",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create a cobra command with a buffer for output
			buf := new(bytes.Buffer)
			cmd := NewStatusCmd()
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := displayStatusWithVehicle(cmd, tt.statusType, tt.vehicleStatus, tt.evStatus, tt.vehicleInfo, tt.jsonOutput)
			require.NoError(t, err, "Unexpected error: %v")

			output := buf.String()
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, output, expected)
			}
		})
	}
}

// TestDisplayAllStatus tests the displayAllStatus function
func TestDisplayAllStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		vehicleStatus  *api.VehicleStatusResponse
		evStatus       *api.EVVehicleStatusResponse
		vehicleInfo    VehicleInfo
		jsonOutput     bool
		expectedOutput []string
		expectJSON     bool
	}{
		{
			name: "full status display",
			vehicleStatus: &api.VehicleStatusResponse{
				ResultCode: api.ResultCodeSuccess,
				AlertInfos: []api.AlertInfo{
					{
						Door: api.DoorInfo{
							DrStatDrv:       float64(api.DoorClosed),
							DrStatPsngr:     float64(api.DoorClosed),
							DrStatRl:        float64(api.DoorClosed),
							DrStatRr:        float64(api.DoorClosed),
							DrStatTrnkLg:    float64(api.DoorClosed),
							DrStatHood:      float64(api.DoorClosed),
							LockLinkSwDrv:   float64(api.DoorLocked),
							LockLinkSwPsngr: float64(api.DoorLocked),
							LockLinkSwRl:    float64(api.DoorLocked),
							LockLinkSwRr:    float64(api.DoorLocked),
						},
					},
				},
				RemoteInfos: []api.RemoteInfo{
					{
						ResidualFuel: api.ResidualFuel{
							FuelSegmentDActl:  92,
							RemDrvDistDActlKm: 630.0,
						},
						TPMSInformation: api.TPMSInformation{
							FLTPrsDispPsi: 32.0,
							FRTPrsDispPsi: 32.0,
							RLTPrsDispPsi: 31.0,
							RRTPrsDispPsi: 31.0,
						},
						DriveInformation: api.DriveInformation{
							OdoDispValue: 12345.6,
						},
					},
				},
			},
			evStatus: &api.EVVehicleStatusResponse{
				ResultCode: api.ResultCodeSuccess,
				ResultData: []api.EVResultData{
					{
						OccurrenceDate: "20250115120000",
						PlusBInformation: api.PlusBInformation{
							VehicleInfo: api.EVVehicleInfo{
								ChargeInfo: api.ChargeInfo{
									SmaphSOC:          80.0,
									SmaphRemDrvDistKm: 200.0,
								},
								RemoteHvacInfo: &api.RemoteHvacInfo{
									HVAC:           float64(api.HVACStatusOff),
									FrontDefroster: float64(api.DefrosterOff),
									RearDefogger:   float64(api.DefrosterOff),
									InCarTeDC:      20.0,
									TargetTemp:     22.0,
								},
							},
						},
					},
				},
			},
			vehicleInfo: VehicleInfo{
				VIN:       "JM3KKEHC1R0123456",
				ModelName: "CX-90 PHEV",
				ModelYear: "2024",
			},
			jsonOutput: false,
			expectedOutput: []string{
				"CX-90 PHEV (2024)",
				"VIN: JM3KKEHC1R0123456",
				"Status as of",
				"BATTERY:",
				"FUEL:",
				"CLIMATE:",
				"DOORS:",
				"WINDOWS:",
				"TIRES:",
				"ODOMETER:",
			},
		},
		{
			name: "status with hazards on",
			vehicleStatus: &api.VehicleStatusResponse{
				ResultCode: api.ResultCodeSuccess,
				AlertInfos: []api.AlertInfo{
					{
						Door: api.DoorInfo{
							DrStatDrv:       float64(api.DoorClosed),
							DrStatPsngr:     float64(api.DoorClosed),
							DrStatRl:        float64(api.DoorClosed),
							DrStatRr:        float64(api.DoorClosed),
							DrStatTrnkLg:    float64(api.DoorClosed),
							DrStatHood:      float64(api.DoorClosed),
							LockLinkSwDrv:   float64(api.DoorLocked),
							LockLinkSwPsngr: float64(api.DoorLocked),
							LockLinkSwRl:    float64(api.DoorLocked),
							LockLinkSwRr:    float64(api.DoorLocked),
						},
						HazardLamp: api.HazardLamp{
							HazardSw: float64(api.HazardLightsOn),
						},
					},
				},
				RemoteInfos: []api.RemoteInfo{
					{
						ResidualFuel: api.ResidualFuel{
							FuelSegmentDActl:  92,
							RemDrvDistDActlKm: 630.0,
						},
						TPMSInformation: api.TPMSInformation{
							FLTPrsDispPsi: 32.0,
							FRTPrsDispPsi: 32.0,
							RLTPrsDispPsi: 31.0,
							RRTPrsDispPsi: 31.0,
						},
						DriveInformation: api.DriveInformation{
							OdoDispValue: 12345.6,
						},
					},
				},
			},
			evStatus: NewMockEVVehicleStatus().Build(),
			vehicleInfo: VehicleInfo{
				VIN: "JM3KKEHC1R0123456",
			},
			jsonOutput: false,
			expectedOutput: []string{
				"HAZARDS: On",
			},
		},
		{
			name:          "JSON output",
			vehicleStatus: NewMockVehicleStatus().Build(),
			evStatus:      NewMockEVVehicleStatus().Build(),
			vehicleInfo: VehicleInfo{
				VIN:       "JM3KKEHC1R0123456",
				ModelName: "CX-90 PHEV",
				ModelYear: "2024",
			},
			jsonOutput: true,
			expectJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := displayAllStatus(tt.vehicleStatus, tt.evStatus, tt.vehicleInfo, tt.jsonOutput)
			require.NoError(t, err, "Unexpected error: %v")

			if tt.expectJSON {
				data := parseJSONToMap(t, result)
				// Verify JSON structure has expected top-level keys
				expectedKeys := []string{"vehicle", "battery", "fuel", "location", "tires", "doors", "windows", "hazards", "climate", "odometer"}
				for _, key := range expectedKeys {
					assert.Contains(t, data, key)
				}
			} else {
				for _, expected := range tt.expectedOutput {
					assert.Contains(t, result, expected)
				}
			}
		})
	}
}

// TestDisplayAllStatus_ErrorHandling tests error cases in displayAllStatus
func TestDisplayAllStatus_ErrorHandling(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		vehicleStatus *api.VehicleStatusResponse
		evStatus      *api.EVVehicleStatusResponse
		vehicleInfo   VehicleInfo
		expectError   bool
	}{
		{
			name:          "missing occurrence date",
			vehicleStatus: NewMockVehicleStatus().Build(),
			evStatus: &api.EVVehicleStatusResponse{
				ResultCode: api.ResultCodeSuccess,
				ResultData: []api.EVResultData{},
			},
			vehicleInfo: VehicleInfo{
				VIN: "JM3KKEHC1R0123456",
			},
			expectError: true,
		},
		{
			name:          "missing HVAC info",
			vehicleStatus: NewMockVehicleStatus().Build(),
			evStatus:      NewMockEVVehicleStatus().WithoutHVAC().Build(),
			vehicleInfo: VehicleInfo{
				VIN: "JM3KKEHC1R0123456",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := displayAllStatus(tt.vehicleStatus, tt.evStatus, tt.vehicleInfo, false)
			if tt.expectError {
				require.Error(t, err, "Expected error, got nil")
			} else {
				require.NoErrorf(t, err, "Expected no error, got: %v", err)
			}
		})
	}
}

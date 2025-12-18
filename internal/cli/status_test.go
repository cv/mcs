package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

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

// findSubcommand finds a subcommand by name
func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == name {
			return subCmd
		}
	}
	return nil
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
		name           string
		batteryLevel   float64
		range_         float64
		pluggedIn      bool
		charging       bool
		expectedOutput string
	}{
		{
			name:           "charging",
			batteryLevel:   66,
			range_:         245.5,
			pluggedIn:      true,
			charging:       true,
			expectedOutput: "BATTERY: 66% (245.5 km range) [plugged in, charging]",
		},
		{
			name:           "plugged not charging",
			batteryLevel:   100,
			range_:         300.0,
			pluggedIn:      true,
			charging:       false,
			expectedOutput: "BATTERY: 100% (300.0 km range) [plugged in, not charging]",
		},
		{
			name:           "unplugged",
			batteryLevel:   50,
			range_:         150.0,
			pluggedIn:      false,
			charging:       false,
			expectedOutput: "BATTERY: 50% (150.0 km range)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBatteryStatus(tt.batteryLevel, tt.range_, tt.pluggedIn, tt.charging, false)
			if !strings.Contains(result, tt.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.expectedOutput, result)
			}
		})
	}
}

// TestFormatBatteryStatus_JSON tests battery status JSON formatting
func TestFormatBatteryStatus_JSON(t *testing.T) {
	result := formatBatteryStatus(66, 245.5, true, true, true)

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}

	if data["battery_level"] != float64(66) {
		t.Errorf("Expected battery_level 66, got %v", data["battery_level"])
	}

	if data["range_km"] != 245.5 {
		t.Errorf("Expected range_km 245.5, got %v", data["range_km"])
	}

	if data["plugged_in"] != true {
		t.Errorf("Expected plugged_in true, got %v", data["plugged_in"])
	}

	if data["charging"] != true {
		t.Errorf("Expected charging true, got %v", data["charging"])
	}
}

// TestFormatFuelStatus tests fuel status formatting
func TestFormatFuelStatus(t *testing.T) {
	result := formatFuelStatus(92, 630.0, false)
	expected := "FUEL: 92% (630.0 km range)"

	if !strings.Contains(result, expected) {
		t.Errorf("Expected output to contain '%s', got '%s'", expected, result)
	}
}

// TestFormatFuelStatus_JSON tests fuel status JSON formatting
func TestFormatFuelStatus_JSON(t *testing.T) {
	result := formatFuelStatus(92, 630.0, true)

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}

	if data["fuel_level"] != float64(92) {
		t.Errorf("Expected fuel_level 92, got %v", data["fuel_level"])
	}

	if data["range_km"] != 630.0 {
		t.Errorf("Expected range_km 630.0, got %v", data["range_km"])
	}
}

// TestFormatDoorsStatus tests doors status formatting
func TestFormatDoorsStatus(t *testing.T) {
	tests := []struct {
		name           string
		allLocked      bool
		expectedOutput string
	}{
		{
			name:           "all locked",
			allLocked:      true,
			expectedOutput: "DOORS: All locked",
		},
		{
			name:           "not all locked",
			allLocked:      false,
			expectedOutput: "DOORS: Not all locked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDoorsStatus(tt.allLocked, false)
			if !strings.Contains(result, tt.expectedOutput) {
				t.Errorf("Expected output to contain '%s', got '%s'", tt.expectedOutput, result)
			}
		})
	}
}

// TestFormatTiresStatus tests tire status formatting
func TestFormatTiresStatus(t *testing.T) {
	result := formatTiresStatus(32.5, 32.0, 31.5, 31.8, false)
	expected := "TIRES: FL:32.5 FR:32.0 RL:31.5 RR:31.8 PSI"

	if !strings.Contains(result, expected) {
		t.Errorf("Expected output to contain '%s', got '%s'", expected, result)
	}
}

// TestFormatLocationStatus tests location status formatting
func TestFormatLocationStatus(t *testing.T) {
	result := formatLocationStatus(37.7749, 122.4194, "20231201120000", false)

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
		ResultCode: "200S00",
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
		ResultCode:   "200S00",
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
		name            string
		hvacOn          bool
		frontDefroster  bool
		rearDefroster   bool
		interiorTempC   float64
		expectedOutput  string
	}{
		{
			name:           "hvac on with both defrosters",
			hvacOn:         true,
			frontDefroster: true,
			rearDefroster:  true,
			interiorTempC:  21,
			expectedOutput: "CLIMATE: On, 21°C (front and rear defrosters on)",
		},
		{
			name:           "hvac on with front defroster only",
			hvacOn:         true,
			frontDefroster: true,
			rearDefroster:  false,
			interiorTempC:  19,
			expectedOutput: "CLIMATE: On, 19°C (front defroster on)",
		},
		{
			name:           "hvac on with rear defroster only",
			hvacOn:         true,
			frontDefroster: false,
			rearDefroster:  true,
			interiorTempC:  22,
			expectedOutput: "CLIMATE: On, 22°C (rear defroster on)",
		},
		{
			name:           "hvac on no defrosters",
			hvacOn:         true,
			frontDefroster: false,
			rearDefroster:  false,
			interiorTempC:  20,
			expectedOutput: "CLIMATE: On, 20°C",
		},
		{
			name:           "hvac off",
			hvacOn:         false,
			frontDefroster: false,
			rearDefroster:  false,
			interiorTempC:  15,
			expectedOutput: "CLIMATE: Off, 15°C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHvacStatus(tt.hvacOn, tt.frontDefroster, tt.rearDefroster, tt.interiorTempC, false)
			if result != tt.expectedOutput {
				t.Errorf("Expected '%s', got '%s'", tt.expectedOutput, result)
			}
		})
	}
}

// TestFormatHvacStatus_JSON tests HVAC status JSON formatting
func TestFormatHvacStatus_JSON(t *testing.T) {
	result := formatHvacStatus(true, true, false, 21, true)

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(result), &data); err != nil {
		t.Fatalf("Expected valid JSON, got error: %v", err)
	}

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
}

// TestGetHvacInfo tests extracting HVAC info from EV status
func TestGetHvacInfo(t *testing.T) {
	evStatus := &api.EVVehicleStatusResponse{
		ResultCode: "200S00",
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
						},
					},
				},
			},
		},
	}

	hvacOn, frontDefroster, rearDefroster, interiorTempC := evStatus.GetHvacInfo()

	if !hvacOn {
		t.Error("Expected hvacOn to be true")
	}
	if !frontDefroster {
		t.Error("Expected frontDefroster to be true")
	}
	if rearDefroster {
		t.Error("Expected rearDefroster to be false")
	}
	if interiorTempC != 21.5 {
		t.Errorf("Expected interiorTempC 21.5, got %v", interiorTempC)
	}
}

// TestGetHvacInfo_MissingData tests extracting HVAC info when data is missing
func TestGetHvacInfo_MissingData(t *testing.T) {
	evStatus := &api.EVVehicleStatusResponse{
		ResultCode: "200S00",
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

	hvacOn, frontDefroster, rearDefroster, interiorTempC := evStatus.GetHvacInfo()

	if hvacOn {
		t.Error("Expected hvacOn to be false when data missing")
	}
	if frontDefroster {
		t.Error("Expected frontDefroster to be false when data missing")
	}
	if rearDefroster {
		t.Error("Expected rearDefroster to be false when data missing")
	}
	if interiorTempC != 0 {
		t.Errorf("Expected interiorTempC 0 when data missing, got %v", interiorTempC)
	}
}

// TestExtractHvacData tests extracting HVAC data for JSON output
func TestExtractHvacData(t *testing.T) {
	evStatus := &api.EVVehicleStatusResponse{
		ResultCode: "200S00",
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
						},
					},
				},
			},
		},
	}

	data := extractHvacData(evStatus)

	if data["hvac_on"] != true {
		t.Errorf("Expected hvac_on true, got %v", data["hvac_on"])
	}
	if data["front_defroster"] != false {
		t.Errorf("Expected front_defroster false, got %v", data["front_defroster"])
	}
	if data["rear_defroster"] != true {
		t.Errorf("Expected rear_defroster true, got %v", data["rear_defroster"])
	}
	if data["interior_temperature_c"] != float64(18) {
		t.Errorf("Expected interior_temperature_c 18, got %v", data["interior_temperature_c"])
	}
}

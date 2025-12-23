package cli

import (
	"encoding/json"
	"testing"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// findSubcommand finds a subcommand by name in the given parent command.
// Returns nil if not found.
func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == name {
			return subCmd
		}
	}
	return nil
}

// assertCommandBasics tests that a command has the expected Use field
// and a non-empty Short description. This helper eliminates duplicated
// test patterns across command test files.
func assertCommandBasics(t *testing.T, cmd *cobra.Command, expectedUse string) {
	t.Helper()

	assert.Equal(t, expectedUse, cmd.Use)

	assert.NotEmpty(t, cmd.Short, "Expected Short description to be set")
}

// assertNoArgsCommand tests that a command accepts no arguments.
func assertNoArgsCommand(t *testing.T, cmd *cobra.Command) {
	t.Helper()

	require.NoError(t, cmd.ValidateArgs([]string{}))
}

// assertSubcommandExists verifies that a subcommand exists, has a description,
// and optionally validates that it accepts no arguments.
func assertSubcommandExists(t *testing.T, parent *cobra.Command, subcommandName string, shouldAcceptNoArgs bool) *cobra.Command {
	t.Helper()

	subCmd := findSubcommand(parent, subcommandName)
	require.NotNilf(t, subCmd, "Expected %s subcommand to exist", subcommandName)

	assert.NotEmptyf(t, subCmd.Short, "Expected %s subcommand to have a description", subcommandName)

	if shouldAcceptNoArgs {
		require.NoError(t, subCmd.ValidateArgs([]string{}))
	}

	return subCmd
}

// assertSubcommandsExist verifies that multiple subcommands exist with descriptions
// and optionally validates that they accept no arguments.
func assertSubcommandsExist(t *testing.T, parent *cobra.Command, subcommandNames []string, shouldAcceptNoArgs bool) {
	t.Helper()

	for _, name := range subcommandNames {
		t.Run(name, func(t *testing.T) {
			assertSubcommandExists(t, parent, name, shouldAcceptNoArgs)
		})
	}
}

// FlagAssertion represents expectations for a command flag.
type FlagAssertion struct {
	Name         string
	DefaultValue string
}

// assertFlagExists verifies that a flag exists and optionally checks its default value.
func assertFlagExists(t *testing.T, cmd *cobra.Command, assertion FlagAssertion) {
	t.Helper()

	flag := cmd.Flags().Lookup(assertion.Name)
	assert.NotNilf(t, flag, "Expected --%s flag to exist", assertion.Name)

	if assertion.DefaultValue != "" {
		assert.Equal(t, assertion.DefaultValue, flag.DefValue)
	}

}

// MockVehicleStatusBuilder provides a fluent API for building mock VehicleStatusResponse objects.
type MockVehicleStatusBuilder struct {
	response *api.VehicleStatusResponse
}

// NewMockVehicleStatus creates a new builder with sensible defaults.
func NewMockVehicleStatus() *MockVehicleStatusBuilder {
	return &MockVehicleStatusBuilder{
		response: &api.VehicleStatusResponse{
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
						LockLinkSwDrv:   float64(api.DoorUnlocked),
						LockLinkSwPsngr: float64(api.DoorUnlocked),
						LockLinkSwRl:    float64(api.DoorUnlocked),
						LockLinkSwRr:    float64(api.DoorUnlocked),
					},
				},
			},
			RemoteInfos: []api.RemoteInfo{{}},
		},
	}
}

// WithDoorStatus sets the door status for the mock response.
func (b *MockVehicleStatusBuilder) WithDoorStatus(status api.DoorStatus) *MockVehicleStatusBuilder {
	doorInfo := &b.response.AlertInfos[0].Door

	doorInfo.DrStatDrv = boolToDoorState(status.DriverOpen)
	doorInfo.DrStatPsngr = boolToDoorState(status.PassengerOpen)
	doorInfo.DrStatRl = boolToDoorState(status.RearLeftOpen)
	doorInfo.DrStatRr = boolToDoorState(status.RearRightOpen)
	doorInfo.DrStatTrnkLg = boolToDoorState(status.TrunkOpen)
	doorInfo.DrStatHood = boolToDoorState(status.HoodOpen)

	doorInfo.LockLinkSwDrv = boolToLockState(status.DriverLocked)
	doorInfo.LockLinkSwPsngr = boolToLockState(status.PassengerLocked)
	doorInfo.LockLinkSwRl = boolToLockState(status.RearLeftLocked)
	doorInfo.LockLinkSwRr = boolToLockState(status.RearRightLocked)

	return b
}

// Build returns the constructed VehicleStatusResponse.
func (b *MockVehicleStatusBuilder) Build() *api.VehicleStatusResponse {
	return b.response
}

// boolToDoorState converts a boolean door state to the API's numeric representation.
func boolToDoorState(isOpen bool) float64 {
	if isOpen {
		return float64(api.DoorOpen)
	}
	return float64(api.DoorClosed)
}

// boolToLockState converts a boolean lock state to the API's numeric representation.
func boolToLockState(isLocked bool) float64 {
	if isLocked {
		return float64(api.DoorLocked)
	}
	return float64(api.DoorUnlocked)
}

// MockEVVehicleStatusBuilder provides a fluent API for building mock EVVehicleStatusResponse objects.
type MockEVVehicleStatusBuilder struct {
	response *api.EVVehicleStatusResponse
}

// NewMockEVVehicleStatus creates a new builder with sensible defaults.
func NewMockEVVehicleStatus() *MockEVVehicleStatusBuilder {
	return &MockEVVehicleStatusBuilder{
		response: &api.EVVehicleStatusResponse{
			ResultCode: api.ResultCodeSuccess,
			ResultData: []api.EVResultData{
				{
					OccurrenceDate: "2025-01-15 12:00:00",
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
	}
}

// WithHVAC sets the HVAC on/off state.
func (b *MockEVVehicleStatusBuilder) WithHVAC(on bool) *MockEVVehicleStatusBuilder {
	if on {
		b.response.ResultData[0].PlusBInformation.VehicleInfo.RemoteHvacInfo.HVAC = float64(api.HVACStatusOn)
	} else {
		b.response.ResultData[0].PlusBInformation.VehicleInfo.RemoteHvacInfo.HVAC = float64(api.HVACStatusOff)
	}
	return b
}

// WithHVACSettings sets detailed HVAC configuration.
func (b *MockEVVehicleStatusBuilder) WithHVACSettings(on bool, targetTemp float64, frontDefrost, rearDefrost bool) *MockEVVehicleStatusBuilder {
	hvac := b.response.ResultData[0].PlusBInformation.VehicleInfo.RemoteHvacInfo

	if on {
		hvac.HVAC = float64(api.HVACStatusOn)
	} else {
		hvac.HVAC = float64(api.HVACStatusOff)
	}

	hvac.TargetTemp = targetTemp

	if frontDefrost {
		hvac.FrontDefroster = float64(api.DefrosterOn)
	} else {
		hvac.FrontDefroster = float64(api.DefrosterOff)
	}

	if rearDefrost {
		hvac.RearDefogger = float64(api.DefrosterOn)
	} else {
		hvac.RearDefogger = float64(api.DefrosterOff)
	}

	return b
}

// WithCharging sets the charging state.
func (b *MockEVVehicleStatusBuilder) WithCharging(charging bool) *MockEVVehicleStatusBuilder {
	chargeInfo := &b.response.ResultData[0].PlusBInformation.VehicleInfo.ChargeInfo
	chargeInfo.ChargerConnectorFitting = float64(api.ChargerConnected)
	if charging {
		chargeInfo.ChargeStatusSub = float64(api.ChargeStatusCharging)
	} else {
		chargeInfo.ChargeStatusSub = 0
	}
	return b
}

// WithoutHVAC sets the RemoteHvacInfo to nil (simulates vehicle without HVAC data).
func (b *MockEVVehicleStatusBuilder) WithoutHVAC() *MockEVVehicleStatusBuilder {
	b.response.ResultData[0].PlusBInformation.VehicleInfo.RemoteHvacInfo = nil
	return b
}

// Build returns the constructed EVVehicleStatusResponse.
func (b *MockEVVehicleStatusBuilder) Build() *api.EVVehicleStatusResponse {
	return b.response
}

// parseJSONToMap is a test helper that parses a JSON string into a map.
func parseJSONToMap(t *testing.T, jsonStr string) map[string]interface{} {
	t.Helper()
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	require.NoError(t, err, "Expected valid JSON, got error: %v")

	return data
}

// assertMapValue is a test helper that asserts a map value equals expected.
func assertMapValue(t *testing.T, data map[string]interface{}, key string, expected interface{}) {
	t.Helper()
	actual, ok := data[key]
	assert.Truef(t, ok, "Expected key %q to exist in map", key)
	assert.Equalf(t, expected, actual, "Expected %s to be %v, got %v", key, expected, actual)
}

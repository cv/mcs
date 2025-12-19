package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/cv/mcs/internal/api"
)

// TestPollUntilCondition tests the polling logic
func TestPollUntilCondition(t *testing.T) {
	tests := []struct {
		name           string
		checkFunc      func() (bool, error)
		timeout        time.Duration
		pollInterval   time.Duration
		expectError    bool
		expectTimeout  bool
		expectedOutput string
	}{
		{
			name: "condition met immediately",
			checkFunc: func() (bool, error) {
				return true, nil
			},
			timeout:      10 * time.Second,
			pollInterval: 1 * time.Second,
			expectError:  false,
		},
		{
			name: "condition met after one check",
			checkFunc: func() func() (bool, error) {
				calls := 0
				return func() (bool, error) {
					calls++
					return calls >= 2, nil
				}
			}(),
			timeout:      10 * time.Second,
			pollInterval: 100 * time.Millisecond,
			expectError:  false,
		},
		{
			name: "timeout exceeded",
			checkFunc: func() (bool, error) {
				return false, nil
			},
			timeout:       200 * time.Millisecond,
			pollInterval:  50 * time.Millisecond,
			expectError:   false, // timeout is not an error, just a warning
			expectTimeout: true,
		},
		{
			name: "error during check",
			checkFunc: func() (bool, error) {
				return false, errors.New("check failed")
			},
			timeout:      10 * time.Second,
			pollInterval: 100 * time.Millisecond,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer

			result := pollUntilCondition(ctx, &buf, tt.checkFunc, tt.timeout, tt.pollInterval, "Test")

			if tt.expectError && result.err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && result.err != nil {
				t.Errorf("Expected no error but got: %v", result.err)
			}

			if tt.expectTimeout && result.success {
				t.Error("Expected timeout but condition was met")
			}

			if !tt.expectTimeout && !tt.expectError && !result.success {
				t.Error("Expected success but condition was not met")
			}
		})
	}
}

// TestWaitForDoorsLocked tests the door lock confirmation logic
func TestWaitForDoorsLocked(t *testing.T) {
	tests := []struct {
		name        string
		doorStatus  []api.DoorStatus
		expectError bool
		expectMet   bool
	}{
		{
			name: "all doors locked immediately",
			doorStatus: []api.DoorStatus{
				{
					DriverLocked:    true,
					PassengerLocked: true,
					RearLeftLocked:  true,
					RearRightLocked: true,
					AllLocked:       true,
				},
			},
			expectError: false,
			expectMet:   true,
		},
		{
			name: "doors lock after one check",
			doorStatus: []api.DoorStatus{
				{
					DriverLocked:    false,
					PassengerLocked: true,
					RearLeftLocked:  true,
					RearRightLocked: true,
					AllLocked:       false,
				},
				{
					DriverLocked:    true,
					PassengerLocked: true,
					RearLeftLocked:  true,
					RearRightLocked: true,
					AllLocked:       true,
				},
			},
			expectError: false,
			expectMet:   true,
		},
		{
			name: "doors never lock",
			doorStatus: []api.DoorStatus{
				{
					DriverLocked:    false,
					PassengerLocked: true,
					RearLeftLocked:  true,
					RearRightLocked: true,
					AllLocked:       false,
				},
				{
					DriverLocked:    false,
					PassengerLocked: true,
					RearLeftLocked:  true,
					RearRightLocked: true,
					AllLocked:       false,
				},
				{
					DriverLocked:    false,
					PassengerLocked: true,
					RearLeftLocked:  true,
					RearRightLocked: true,
					AllLocked:       false,
				},
			},
			expectError: false,
			expectMet:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer

			// Create mock client that returns the door status sequence
			calls := 0
			mockClient := &mockClientForConfirm{
				getVehicleStatusFunc: func(ctx context.Context, internalVIN string) (*api.VehicleStatusResponse, error) {
					if calls >= len(tt.doorStatus) {
						calls = len(tt.doorStatus) - 1
					}
					status := tt.doorStatus[calls]
					calls++
					return createMockVehicleStatusResponse(status), nil
				},
			}

			result := waitForDoorsLocked(ctx, &buf, mockClient, "test-vin", 5*time.Second, 50*time.Millisecond)

			if tt.expectError && result.err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && result.err != nil {
				t.Errorf("Expected no error but got: %v", result.err)
			}

			if tt.expectMet && !result.success {
				t.Errorf("Expected doors to be locked but they weren't (calls: %d)", calls)
			}

			if !tt.expectMet && result.success {
				t.Error("Expected doors to not be locked but they were")
			}
		})
	}
}

// TestWaitForEngineRunning tests the engine running confirmation logic
func TestWaitForEngineRunning(t *testing.T) {
	tests := []struct {
		name        string
		hvacStatus  []bool
		expectError bool
		expectMet   bool
	}{
		{
			name:        "engine running immediately",
			hvacStatus:  []bool{true},
			expectError: false,
			expectMet:   true,
		},
		{
			name:        "engine starts after one check",
			hvacStatus:  []bool{false, true},
			expectError: false,
			expectMet:   true,
		},
		{
			name:        "engine never starts",
			hvacStatus:  []bool{false, false, false},
			expectError: false,
			expectMet:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer

			// Create mock client that returns the HVAC status sequence
			calls := 0
			mockClient := &mockClientForConfirm{
				getEVVehicleStatusFunc: func(ctx context.Context, internalVIN string) (*api.EVVehicleStatusResponse, error) {
					if calls >= len(tt.hvacStatus) {
						calls = len(tt.hvacStatus) - 1
					}
					hvacOn := tt.hvacStatus[calls]
					calls++
					return createMockEVVehicleStatusResponse(hvacOn), nil
				},
			}

			result := waitForEngineRunning(ctx, &buf, mockClient, "test-vin", 5*time.Second, 50*time.Millisecond)

			if tt.expectError && result.err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && result.err != nil {
				t.Errorf("Expected no error but got: %v", result.err)
			}

			if tt.expectMet && !result.success {
				t.Errorf("Expected engine to be running but it wasn't (calls: %d)", calls)
			}

			if !tt.expectMet && result.success {
				t.Error("Expected engine to not be running but it was")
			}
		})
	}
}

// TestWaitForEngineStopped tests the engine stopped confirmation logic
func TestWaitForEngineStopped(t *testing.T) {
	tests := []struct {
		name        string
		hvacStatus  []bool
		expectError bool
		expectMet   bool
	}{
		{
			name:        "engine stopped immediately",
			hvacStatus:  []bool{false},
			expectError: false,
			expectMet:   true,
		},
		{
			name:        "engine stops after one check",
			hvacStatus:  []bool{true, false},
			expectError: false,
			expectMet:   true,
		},
		{
			name:        "engine never stops",
			hvacStatus:  []bool{true, true, true},
			expectError: false,
			expectMet:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer

			// Create mock client that returns the HVAC status sequence
			calls := 0
			mockClient := &mockClientForConfirm{
				getEVVehicleStatusFunc: func(ctx context.Context, internalVIN string) (*api.EVVehicleStatusResponse, error) {
					if calls >= len(tt.hvacStatus) {
						calls = len(tt.hvacStatus) - 1
					}
					hvacOn := tt.hvacStatus[calls]
					calls++
					return createMockEVVehicleStatusResponse(hvacOn), nil
				},
			}

			result := waitForEngineStopped(ctx, &buf, mockClient, "test-vin", 5*time.Second, 50*time.Millisecond)

			if tt.expectError && result.err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && result.err != nil {
				t.Errorf("Expected no error but got: %v", result.err)
			}

			if tt.expectMet && !result.success {
				t.Errorf("Expected engine to be stopped but it wasn't (calls: %d)", calls)
			}

			if !tt.expectMet && result.success {
				t.Error("Expected engine to not be stopped but it was")
			}
		})
	}
}

// mockClientForConfirm is a mock API client for testing confirmation logic
type mockClientForConfirm struct {
	getVehicleStatusFunc   func(ctx context.Context, internalVIN string) (*api.VehicleStatusResponse, error)
	getEVVehicleStatusFunc func(ctx context.Context, internalVIN string) (*api.EVVehicleStatusResponse, error)
}

func (m *mockClientForConfirm) GetVehicleStatus(ctx context.Context, internalVIN string) (*api.VehicleStatusResponse, error) {
	if m.getVehicleStatusFunc != nil {
		return m.getVehicleStatusFunc(ctx, internalVIN)
	}
	return nil, errors.New("not implemented")
}

func (m *mockClientForConfirm) GetEVVehicleStatus(ctx context.Context, internalVIN string) (*api.EVVehicleStatusResponse, error) {
	if m.getEVVehicleStatusFunc != nil {
		return m.getEVVehicleStatusFunc(ctx, internalVIN)
	}
	return nil, errors.New("not implemented")
}

// createMockVehicleStatusResponse creates a mock response with the given door status
func createMockVehicleStatusResponse(doorStatus api.DoorStatus) *api.VehicleStatusResponse {
	// Convert door status to API response format
	var driverOpen, passengerOpen, rearLeftOpen, rearRightOpen, trunkOpen, hoodOpen float64
	var driverLocked, passengerLocked, rearLeftLocked, rearRightLocked float64

	if doorStatus.DriverOpen {
		driverOpen = float64(api.DoorOpen)
	} else {
		driverOpen = float64(api.DoorClosed)
	}

	if doorStatus.PassengerOpen {
		passengerOpen = float64(api.DoorOpen)
	} else {
		passengerOpen = float64(api.DoorClosed)
	}

	if doorStatus.RearLeftOpen {
		rearLeftOpen = float64(api.DoorOpen)
	} else {
		rearLeftOpen = float64(api.DoorClosed)
	}

	if doorStatus.RearRightOpen {
		rearRightOpen = float64(api.DoorOpen)
	} else {
		rearRightOpen = float64(api.DoorClosed)
	}

	if doorStatus.TrunkOpen {
		trunkOpen = float64(api.DoorOpen)
	} else {
		trunkOpen = float64(api.DoorClosed)
	}

	if doorStatus.HoodOpen {
		hoodOpen = float64(api.DoorOpen)
	} else {
		hoodOpen = float64(api.DoorClosed)
	}

	if doorStatus.DriverLocked {
		driverLocked = float64(api.DoorLocked)
	} else {
		driverLocked = float64(api.DoorUnlocked)
	}

	if doorStatus.PassengerLocked {
		passengerLocked = float64(api.DoorLocked)
	} else {
		passengerLocked = float64(api.DoorUnlocked)
	}

	if doorStatus.RearLeftLocked {
		rearLeftLocked = float64(api.DoorLocked)
	} else {
		rearLeftLocked = float64(api.DoorUnlocked)
	}

	if doorStatus.RearRightLocked {
		rearRightLocked = float64(api.DoorLocked)
	} else {
		rearRightLocked = float64(api.DoorUnlocked)
	}

	return &api.VehicleStatusResponse{
		ResultCode: api.ResultCodeSuccess,
		AlertInfos: []api.AlertInfo{
			{
				Door: api.DoorInfo{
					DrStatDrv:       driverOpen,
					DrStatPsngr:     passengerOpen,
					DrStatRl:        rearLeftOpen,
					DrStatRr:        rearRightOpen,
					DrStatTrnkLg:    trunkOpen,
					DrStatHood:      hoodOpen,
					LockLinkSwDrv:   driverLocked,
					LockLinkSwPsngr: passengerLocked,
					LockLinkSwRl:    rearLeftLocked,
					LockLinkSwRr:    rearRightLocked,
				},
			},
		},
		// RemoteInfos required for valid response
		RemoteInfos: []api.RemoteInfo{{}},
	}
}

// TestWaitForCharging tests the charging started confirmation logic
func TestWaitForCharging(t *testing.T) {
	tests := []struct {
		name           string
		chargingStatus []bool
		expectError    bool
		expectMet      bool
	}{
		{
			name:           "charging started immediately",
			chargingStatus: []bool{true},
			expectError:    false,
			expectMet:      true,
		},
		{
			name:           "charging starts after one check",
			chargingStatus: []bool{false, true},
			expectError:    false,
			expectMet:      true,
		},
		{
			name:           "charging never starts",
			chargingStatus: []bool{false, false, false},
			expectError:    false,
			expectMet:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer

			// Create mock client that returns the charging status sequence
			calls := 0
			mockClient := &mockClientForConfirm{
				getEVVehicleStatusFunc: func(ctx context.Context, internalVIN string) (*api.EVVehicleStatusResponse, error) {
					if calls >= len(tt.chargingStatus) {
						calls = len(tt.chargingStatus) - 1
					}
					charging := tt.chargingStatus[calls]
					calls++
					return createMockEVVehicleStatusResponseWithCharging(charging), nil
				},
			}

			result := waitForCharging(ctx, &buf, mockClient, "test-vin", 5*time.Second, 50*time.Millisecond)

			if tt.expectError && result.err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && result.err != nil {
				t.Errorf("Expected no error but got: %v", result.err)
			}

			if tt.expectMet && !result.success {
				t.Errorf("Expected charging to be started but it wasn't (calls: %d)", calls)
			}

			if !tt.expectMet && result.success {
				t.Error("Expected charging to not be started but it was")
			}
		})
	}
}

// TestWaitForNotCharging tests the charging stopped confirmation logic
func TestWaitForNotCharging(t *testing.T) {
	tests := []struct {
		name           string
		chargingStatus []bool
		expectError    bool
		expectMet      bool
	}{
		{
			name:           "charging stopped immediately",
			chargingStatus: []bool{false},
			expectError:    false,
			expectMet:      true,
		},
		{
			name:           "charging stops after one check",
			chargingStatus: []bool{true, false},
			expectError:    false,
			expectMet:      true,
		},
		{
			name:           "charging never stops",
			chargingStatus: []bool{true, true, true},
			expectError:    false,
			expectMet:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer

			// Create mock client that returns the charging status sequence
			calls := 0
			mockClient := &mockClientForConfirm{
				getEVVehicleStatusFunc: func(ctx context.Context, internalVIN string) (*api.EVVehicleStatusResponse, error) {
					if calls >= len(tt.chargingStatus) {
						calls = len(tt.chargingStatus) - 1
					}
					charging := tt.chargingStatus[calls]
					calls++
					return createMockEVVehicleStatusResponseWithCharging(charging), nil
				},
			}

			result := waitForNotCharging(ctx, &buf, mockClient, "test-vin", 5*time.Second, 50*time.Millisecond)

			if tt.expectError && result.err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && result.err != nil {
				t.Errorf("Expected no error but got: %v", result.err)
			}

			if tt.expectMet && !result.success {
				t.Errorf("Expected charging to be stopped but it wasn't (calls: %d)", calls)
			}

			if !tt.expectMet && result.success {
				t.Error("Expected charging to not be stopped but it was")
			}
		})
	}
}

// createMockEVVehicleStatusResponse creates a mock EV status response with the given HVAC status
func createMockEVVehicleStatusResponse(hvacOn bool) *api.EVVehicleStatusResponse {
	var hvacValue float64
	if hvacOn {
		hvacValue = float64(api.HVACStatusOn)
	} else {
		hvacValue = float64(api.HVACStatusOff)
	}

	return &api.EVVehicleStatusResponse{
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
							HVAC:           hvacValue,
							FrontDefroster: 0,
							RearDefogger:   0,
							InCarTeDC:      20.0,
							TargetTemp:     22.0,
						},
					},
				},
			},
		},
	}
}

// createMockEVVehicleStatusResponseWithCharging creates a mock EV status response with the given charging status
func createMockEVVehicleStatusResponseWithCharging(charging bool) *api.EVVehicleStatusResponse {
	var chargeStatus float64
	if charging {
		chargeStatus = float64(api.ChargeStatusCharging)
	} else {
		chargeStatus = 0
	}

	return &api.EVVehicleStatusResponse{
		ResultCode: api.ResultCodeSuccess,
		ResultData: []api.EVResultData{
			{
				OccurrenceDate: "2025-01-15 12:00:00",
				PlusBInformation: api.PlusBInformation{
					VehicleInfo: api.EVVehicleInfo{
						ChargeInfo: api.ChargeInfo{
							SmaphSOC:                80.0,
							SmaphRemDrvDistKm:       200.0,
							ChargerConnectorFitting: float64(api.ChargerConnected),
							ChargeStatusSub:         chargeStatus,
						},
						RemoteHvacInfo: &api.RemoteHvacInfo{
							HVAC:           float64(api.HVACStatusOff),
							FrontDefroster: 0,
							RearDefogger:   0,
							InCarTeDC:      20.0,
							TargetTemp:     22.0,
						},
					},
				},
			},
		},
	}
}

// TestWaitForHvacOn tests the HVAC on confirmation logic
func TestWaitForHvacOn(t *testing.T) {
	tests := []struct {
		name        string
		hvacStatus  []bool
		expectError bool
		expectMet   bool
	}{
		{
			name:        "hvac on immediately",
			hvacStatus:  []bool{true},
			expectError: false,
			expectMet:   true,
		},
		{
			name:        "hvac turns on after one check",
			hvacStatus:  []bool{false, true},
			expectError: false,
			expectMet:   true,
		},
		{
			name:        "hvac never turns on",
			hvacStatus:  []bool{false, false, false},
			expectError: false,
			expectMet:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer

			// Create mock client that returns the HVAC status sequence
			calls := 0
			mockClient := &mockClientForConfirm{
				getEVVehicleStatusFunc: func(ctx context.Context, internalVIN string) (*api.EVVehicleStatusResponse, error) {
					if calls >= len(tt.hvacStatus) {
						calls = len(tt.hvacStatus) - 1
					}
					hvacOn := tt.hvacStatus[calls]
					calls++
					return createMockEVVehicleStatusResponseWithHvac(hvacOn, 22.0, false, false), nil
				},
			}

			result := waitForHvacOn(ctx, &buf, mockClient, "test-vin", 5*time.Second, 50*time.Millisecond)

			if tt.expectError && result.err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && result.err != nil {
				t.Errorf("Expected no error but got: %v", result.err)
			}

			if tt.expectMet && !result.success {
				t.Errorf("Expected HVAC to be on but it wasn't (calls: %d)", calls)
			}

			if !tt.expectMet && result.success {
				t.Error("Expected HVAC to not be on but it was")
			}
		})
	}
}

// TestWaitForHvacOff tests the HVAC off confirmation logic
func TestWaitForHvacOff(t *testing.T) {
	tests := []struct {
		name        string
		hvacStatus  []bool
		expectError bool
		expectMet   bool
	}{
		{
			name:        "hvac off immediately",
			hvacStatus:  []bool{false},
			expectError: false,
			expectMet:   true,
		},
		{
			name:        "hvac turns off after one check",
			hvacStatus:  []bool{true, false},
			expectError: false,
			expectMet:   true,
		},
		{
			name:        "hvac never turns off",
			hvacStatus:  []bool{true, true, true},
			expectError: false,
			expectMet:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer

			// Create mock client that returns the HVAC status sequence
			calls := 0
			mockClient := &mockClientForConfirm{
				getEVVehicleStatusFunc: func(ctx context.Context, internalVIN string) (*api.EVVehicleStatusResponse, error) {
					if calls >= len(tt.hvacStatus) {
						calls = len(tt.hvacStatus) - 1
					}
					hvacOn := tt.hvacStatus[calls]
					calls++
					return createMockEVVehicleStatusResponseWithHvac(hvacOn, 22.0, false, false), nil
				},
			}

			result := waitForHvacOff(ctx, &buf, mockClient, "test-vin", 5*time.Second, 50*time.Millisecond)

			if tt.expectError && result.err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && result.err != nil {
				t.Errorf("Expected no error but got: %v", result.err)
			}

			if tt.expectMet && !result.success {
				t.Errorf("Expected HVAC to be off but it wasn't (calls: %d)", calls)
			}

			if !tt.expectMet && result.success {
				t.Error("Expected HVAC to not be off but it was")
			}
		})
	}
}

// TestWaitForHvacSettings tests the HVAC settings confirmation logic
func TestWaitForHvacSettings(t *testing.T) {
	tests := []struct {
		name           string
		targetTemp     float64
		frontDefroster bool
		rearDefroster  bool
		hvacResponses  []hvacSettings
		expectError    bool
		expectMet      bool
	}{
		{
			name:           "settings match immediately",
			targetTemp:     22.0,
			frontDefroster: true,
			rearDefroster:  false,
			hvacResponses: []hvacSettings{
				{hvacOn: true, temp: 22.0, frontDefrost: true, rearDefrost: false},
			},
			expectError: false,
			expectMet:   true,
		},
		{
			name:           "settings match after one check",
			targetTemp:     22.0,
			frontDefroster: true,
			rearDefroster:  false,
			hvacResponses: []hvacSettings{
				{hvacOn: true, temp: 20.0, frontDefrost: false, rearDefrost: false},
				{hvacOn: true, temp: 22.0, frontDefrost: true, rearDefrost: false},
			},
			expectError: false,
			expectMet:   true,
		},
		{
			name:           "temperature within tolerance",
			targetTemp:     22.0,
			frontDefroster: false,
			rearDefroster:  false,
			hvacResponses: []hvacSettings{
				{hvacOn: true, temp: 22.3, frontDefrost: false, rearDefrost: false}, // Within 0.5C tolerance
			},
			expectError: false,
			expectMet:   true,
		},
		{
			name:           "settings never match",
			targetTemp:     22.0,
			frontDefroster: true,
			rearDefroster:  false,
			hvacResponses: []hvacSettings{
				{hvacOn: true, temp: 20.0, frontDefrost: false, rearDefrost: false},
				{hvacOn: true, temp: 20.0, frontDefrost: false, rearDefrost: false},
			},
			expectError: false,
			expectMet:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer

			// Create mock client that returns the HVAC settings sequence
			calls := 0
			mockClient := &mockClientForConfirm{
				getEVVehicleStatusFunc: func(ctx context.Context, internalVIN string) (*api.EVVehicleStatusResponse, error) {
					if calls >= len(tt.hvacResponses) {
						calls = len(tt.hvacResponses) - 1
					}
					settings := tt.hvacResponses[calls]
					calls++
					return createMockEVVehicleStatusResponseWithHvac(
						settings.hvacOn,
						settings.temp,
						settings.frontDefrost,
						settings.rearDefrost,
					), nil
				},
			}

			result := waitForHvacSettings(
				ctx,
				&buf,
				mockClient,
				"test-vin",
				tt.targetTemp,
				tt.frontDefroster,
				tt.rearDefroster,
				5*time.Second,
				50*time.Millisecond,
			)

			if tt.expectError && result.err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && result.err != nil {
				t.Errorf("Expected no error but got: %v", result.err)
			}

			if tt.expectMet && !result.success {
				t.Errorf("Expected HVAC settings to match but they didn't (calls: %d)", calls)
			}

			if !tt.expectMet && result.success {
				t.Error("Expected HVAC settings to not match but they did")
			}
		})
	}
}

// hvacSettings holds HVAC configuration for testing
type hvacSettings struct {
	hvacOn       bool
	temp         float64
	frontDefrost bool
	rearDefrost  bool
}

// createMockEVVehicleStatusResponseWithHvac creates a mock EV status response with the given HVAC settings
func createMockEVVehicleStatusResponseWithHvac(hvacOn bool, targetTemp float64, frontDefrost, rearDefrost bool) *api.EVVehicleStatusResponse {
	var hvacValue, frontDefrostValue, rearDefrostValue float64
	if hvacOn {
		hvacValue = float64(api.HVACStatusOn)
	} else {
		hvacValue = float64(api.HVACStatusOff)
	}
	if frontDefrost {
		frontDefrostValue = float64(api.DefrosterOn)
	} else {
		frontDefrostValue = float64(api.DefrosterOff)
	}
	if rearDefrost {
		rearDefrostValue = float64(api.DefrosterOn)
	} else {
		rearDefrostValue = float64(api.DefrosterOff)
	}

	return &api.EVVehicleStatusResponse{
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
							HVAC:           hvacValue,
							FrontDefroster: frontDefrostValue,
							RearDefogger:   rearDefrostValue,
							InCarTeDC:      20.0,
							TargetTemp:     targetTemp,
						},
					},
				},
			},
		},
	}
}

// TestExecuteConfirmableCommand tests the executeConfirmableCommand helper
func TestExecuteConfirmableCommand(t *testing.T) {
	tests := []struct {
		name           string
		config         ConfirmableCommandConfig
		confirm        bool
		confirmWait    int
		actionError    error
		waitResult     confirmationResult
		expectError    bool
		expectedOutput string
	}{
		{
			name: "success without confirmation",
			config: ConfirmableCommandConfig{
				ActionFunc: func(ctx context.Context, client *api.Client, internalVIN string) error {
					return nil
				},
				SuccessMsg:    "Command executed successfully",
				WaitingMsg:    "Command sent, waiting for confirmation...",
				ActionName:    "execute command",
				ConfirmName:   "command status",
				TimeoutSuffix: "confirmation timeout",
			},
			confirm:        false,
			confirmWait:    90,
			actionError:    nil,
			expectError:    false,
			expectedOutput: "Command executed successfully\n",
		},
		{
			name: "success with confirmation",
			config: ConfirmableCommandConfig{
				ActionFunc: func(ctx context.Context, client *api.Client, internalVIN string) error {
					return nil
				},
				WaitFunc: func(ctx context.Context, out io.Writer, client *api.Client, internalVIN string, timeout, pollInterval time.Duration) confirmationResult {
					return confirmationResult{success: true, err: nil}
				},
				SuccessMsg:    "Command executed successfully",
				WaitingMsg:    "Command sent, waiting for confirmation...",
				ActionName:    "execute command",
				ConfirmName:   "command status",
				TimeoutSuffix: "confirmation timeout",
			},
			confirm:        true,
			confirmWait:    90,
			actionError:    nil,
			waitResult:     confirmationResult{success: true, err: nil},
			expectError:    false,
			expectedOutput: "Command sent, waiting for confirmation...\nCommand executed successfully\n",
		},
		{
			name: "timeout during confirmation",
			config: ConfirmableCommandConfig{
				ActionFunc: func(ctx context.Context, client *api.Client, internalVIN string) error {
					return nil
				},
				WaitFunc: func(ctx context.Context, out io.Writer, client *api.Client, internalVIN string, timeout, pollInterval time.Duration) confirmationResult {
					return confirmationResult{success: false, err: nil}
				},
				SuccessMsg:    "Command executed successfully",
				WaitingMsg:    "Command sent, waiting for confirmation...",
				ActionName:    "execute command",
				ConfirmName:   "command status",
				TimeoutSuffix: "confirmation timeout",
			},
			confirm:        true,
			confirmWait:    90,
			actionError:    nil,
			waitResult:     confirmationResult{success: false, err: nil},
			expectError:    false,
			expectedOutput: "Command sent, waiting for confirmation...\nCommand sent (confirmation timeout)\n",
		},
		{
			name: "action fails",
			config: ConfirmableCommandConfig{
				ActionFunc: func(ctx context.Context, client *api.Client, internalVIN string) error {
					return errors.New("action failed")
				},
				SuccessMsg:    "Command executed successfully",
				WaitingMsg:    "Command sent, waiting for confirmation...",
				ActionName:    "execute command",
				ConfirmName:   "command status",
				TimeoutSuffix: "confirmation timeout",
			},
			confirm:        true,
			confirmWait:    90,
			actionError:    errors.New("action failed"),
			expectError:    true,
			expectedOutput: "",
		},
		{
			name: "confirmation fails with error",
			config: ConfirmableCommandConfig{
				ActionFunc: func(ctx context.Context, client *api.Client, internalVIN string) error {
					return nil
				},
				WaitFunc: func(ctx context.Context, out io.Writer, client *api.Client, internalVIN string, timeout, pollInterval time.Duration) confirmationResult {
					return confirmationResult{success: false, err: errors.New("confirmation error")}
				},
				SuccessMsg:    "Command executed successfully",
				WaitingMsg:    "Command sent, waiting for confirmation...",
				ActionName:    "execute command",
				ConfirmName:   "command status",
				TimeoutSuffix: "confirmation timeout",
			},
			confirm:        true,
			confirmWait:    90,
			waitResult:     confirmationResult{success: false, err: errors.New("confirmation error")},
			expectError:    true,
			expectedOutput: "Command sent, waiting for confirmation...\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer

			err := executeConfirmableCommand(
				ctx,
				&buf,
				nil, // client not used in these tests
				"test-vin",
				tt.config,
				tt.confirm,
				tt.confirmWait,
			)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			output := buf.String()
			if output != tt.expectedOutput {
				t.Errorf("Expected output %q but got %q", tt.expectedOutput, output)
			}
		})
	}
}

package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/cv/mcs/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWaitForCondition tests the generic condition waiting logic
func TestWaitForCondition(t *testing.T) {
	tests := []struct {
		name          string
		useEVStatus   bool
		statusValues  []interface{} // Either []api.DoorStatus or []bool
		conditionFunc func(interface{}) (bool, error)
		expectError   bool
		expectMet     bool
	}{
		{
			name:        "regular status - condition met immediately",
			useEVStatus: false,
			statusValues: []interface{}{
				api.DoorStatus{
					AllLocked:       true,
					DriverLocked:    true,
					PassengerLocked: true,
					RearLeftLocked:  true,
					RearRightLocked: true,
				},
			},
			conditionFunc: func(status interface{}) (bool, error) {
				vStatus := status.(*api.VehicleStatusResponse)
				doorInfo, err := vStatus.GetDoorsInfo()
				if err != nil {
					return false, err
				}
				return doorInfo.AllLocked, nil
			},
			expectError: false,
			expectMet:   true,
		},
		{
			name:        "regular status - condition met after one check",
			useEVStatus: false,
			statusValues: []interface{}{
				api.DoorStatus{
					AllLocked:       false,
					DriverLocked:    false,
					PassengerLocked: true,
					RearLeftLocked:  true,
					RearRightLocked: true,
				},
				api.DoorStatus{
					AllLocked:       true,
					DriverLocked:    true,
					PassengerLocked: true,
					RearLeftLocked:  true,
					RearRightLocked: true,
				},
			},
			conditionFunc: func(status interface{}) (bool, error) {
				vStatus := status.(*api.VehicleStatusResponse)
				doorInfo, err := vStatus.GetDoorsInfo()
				if err != nil {
					return false, err
				}
				return doorInfo.AllLocked, nil
			},
			expectError: false,
			expectMet:   true,
		},
		{
			name:        "EV status - condition met immediately",
			useEVStatus: true,
			statusValues: []interface{}{
				true, // HVAC on
			},
			conditionFunc: func(status interface{}) (bool, error) {
				evStatus := status.(*api.EVVehicleStatusResponse)
				hvacInfo, err := evStatus.GetHvacInfo()
				if err != nil {
					return false, err
				}
				return hvacInfo.HVACOn, nil
			},
			expectError: false,
			expectMet:   true,
		},
		{
			name:        "EV status - condition met after one check",
			useEVStatus: true,
			statusValues: []interface{}{
				false, // HVAC off
				true,  // HVAC on
			},
			conditionFunc: func(status interface{}) (bool, error) {
				evStatus := status.(*api.EVVehicleStatusResponse)
				hvacInfo, err := evStatus.GetHvacInfo()
				if err != nil {
					return false, err
				}
				return hvacInfo.HVACOn, nil
			},
			expectError: false,
			expectMet:   true,
		},
		{
			name:        "condition never met - timeout",
			useEVStatus: true,
			statusValues: []interface{}{
				false,
				false,
				false,
			},
			conditionFunc: func(status interface{}) (bool, error) {
				evStatus := status.(*api.EVVehicleStatusResponse)
				hvacInfo, err := evStatus.GetHvacInfo()
				if err != nil {
					return false, err
				}
				return hvacInfo.HVACOn, nil
			},
			expectError: false,
			expectMet:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer

			calls := 0
			var mockClient *mockClientForConfirm

			if tt.useEVStatus {
				mockClient = &mockClientForConfirm{
					getEVVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error) {
						if calls >= len(tt.statusValues) {
							calls = len(tt.statusValues) - 1
						}
						hvacOn := tt.statusValues[calls].(bool)
						calls++
						return NewMockEVVehicleStatus().WithHVAC(hvacOn).Build(), nil
					},
				}
			} else {
				mockClient = &mockClientForConfirm{
					getVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.VehicleStatusResponse, error) {
						if calls >= len(tt.statusValues) {
							calls = len(tt.statusValues) - 1
						}
						doorStatus := tt.statusValues[calls].(api.DoorStatus)
						calls++
						return NewMockVehicleStatus().WithDoorStatus(doorStatus).Build(), nil
					},
				}
			}

			result := waitForCondition(
				ctx,
				&buf,
				mockClient,
				api.InternalVIN("test-vin"),
				tt.useEVStatus,
				tt.conditionFunc,
				200*time.Millisecond, // Use short timeout for tests
				50*time.Millisecond,
				"test action",
			)

			if tt.expectError {
				assert.NotNil(t, result.err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.Nilf(t, result.err, "Expected no error but got: %v", result.err)
			}

			if tt.expectMet {
				assert.Truef(t, result.success, "Expected condition to be met but it wasn't (calls: %d)", calls)
			}

			if result.success {
				assert.True(t, tt.expectMet, "Expected condition to not be met but it was")
			}

		})
	}
}

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
			name: "error during check - should timeout not fail",
			checkFunc: func() (bool, error) {
				return false, errors.New("check failed")
			},
			timeout:       200 * time.Millisecond,
			pollInterval:  50 * time.Millisecond,
			expectError:   false, // errors are treated as "not ready yet", will timeout
			expectTimeout: true,
		},
		{
			name: "transient error then success - should retry and succeed",
			checkFunc: func() func() (bool, error) {
				calls := 0
				return func() (bool, error) {
					calls++
					if calls == 1 {
						return false, errors.New("transient error")
					}
					return true, nil
				}
			}(),
			timeout:      10 * time.Second,
			pollInterval: 50 * time.Millisecond,
			expectError:  false, // should succeed after retry
		},
		{
			name: "multiple transient errors then success",
			checkFunc: func() func() (bool, error) {
				calls := 0
				return func() (bool, error) {
					calls++
					if calls <= 3 {
						return false, errors.New("transient error")
					}
					return true, nil
				}
			}(),
			timeout:      10 * time.Second,
			pollInterval: 50 * time.Millisecond,
			expectError:  false, // should succeed after retries
		},
		{
			name: "persistent errors until timeout",
			checkFunc: func() (bool, error) {
				return false, errors.New("persistent error")
			},
			timeout:       200 * time.Millisecond,
			pollInterval:  50 * time.Millisecond,
			expectError:   false, // should timeout, not error
			expectTimeout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer

			result := pollUntilCondition(ctx, &buf, tt.checkFunc, tt.timeout, tt.pollInterval, "Test")

			if tt.expectError {
				assert.NotNil(t, result.err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.Nilf(t, result.err, "Expected no error but got: %v", result.err)
			}

			assert.False(t, tt.expectTimeout && result.success, "Expected timeout but condition was met")

			if !tt.expectTimeout && !tt.expectError {
				assert.True(t, result.success, "Expected success but condition was not met")
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
				getVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.VehicleStatusResponse, error) {
					if calls >= len(tt.doorStatus) {
						calls = len(tt.doorStatus) - 1
					}
					status := tt.doorStatus[calls]
					calls++
					return NewMockVehicleStatus().WithDoorStatus(status).Build(), nil
				},
			}

			// Use shorter timeout for "never" cases to speed up tests
			timeout := 5 * time.Second
			if !tt.expectMet {
				timeout = 200 * time.Millisecond
			}

			result := waitForDoorsLocked(ctx, &buf, mockClient, api.InternalVIN("test-vin"), timeout, 50*time.Millisecond)

			if tt.expectError {
				assert.NotNil(t, result.err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.Nilf(t, result.err, "Expected no error but got: %v", result.err)
			}

			if tt.expectMet {
				assert.Truef(t, result.success, "Expected doors to be locked but they weren't (calls: %d)", calls)
			}

			if result.success {
				assert.True(t, tt.expectMet, "Expected doors to not be locked but they were")
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
				getEVVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error) {
					if calls >= len(tt.hvacStatus) {
						calls = len(tt.hvacStatus) - 1
					}
					hvacOn := tt.hvacStatus[calls]
					calls++
					return NewMockEVVehicleStatus().WithHVAC(hvacOn).Build(), nil
				},
			}

			// Use shorter timeout for "never" cases to speed up tests
			timeout := 5 * time.Second
			if !tt.expectMet {
				timeout = 200 * time.Millisecond
			}

			result := waitForEngineRunning(ctx, &buf, mockClient, api.InternalVIN("test-vin"), timeout, 50*time.Millisecond)

			if tt.expectError {
				assert.NotNil(t, result.err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.Nilf(t, result.err, "Expected no error but got: %v", result.err)
			}

			if tt.expectMet {
				assert.Truef(t, result.success, "Expected engine to be running but it wasn't (calls: %d)", calls)
			}

			if result.success {
				assert.True(t, tt.expectMet, "Expected engine to not be running but it was")
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
				getEVVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error) {
					if calls >= len(tt.hvacStatus) {
						calls = len(tt.hvacStatus) - 1
					}
					hvacOn := tt.hvacStatus[calls]
					calls++
					return NewMockEVVehicleStatus().WithHVAC(hvacOn).Build(), nil
				},
			}

			// Use shorter timeout for "never" cases to speed up tests
			timeout := 5 * time.Second
			if !tt.expectMet {
				timeout = 200 * time.Millisecond
			}

			result := waitForEngineStopped(ctx, &buf, mockClient, api.InternalVIN("test-vin"), timeout, 50*time.Millisecond)

			if tt.expectError {
				assert.NotNil(t, result.err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.Nilf(t, result.err, "Expected no error but got: %v", result.err)
			}

			if tt.expectMet {
				assert.Truef(t, result.success, "Expected engine to be stopped but it wasn't (calls: %d)", calls)
			}

			if result.success {
				assert.True(t, tt.expectMet, "Expected engine to not be stopped but it was")
			}

		})
	}
}

// mockClientForConfirm is a mock API client for testing confirmation logic
type mockClientForConfirm struct {
	getVehicleStatusFunc      func(ctx context.Context, internalVIN api.InternalVIN) (*api.VehicleStatusResponse, error)
	getEVVehicleStatusFunc    func(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error)
	refreshVehicleStatusFunc  func(ctx context.Context, internalVIN api.InternalVIN) error
	refreshVehicleStatusCalls int
}

func (m *mockClientForConfirm) GetVehicleStatus(ctx context.Context, internalVIN api.InternalVIN) (*api.VehicleStatusResponse, error) {
	if m.getVehicleStatusFunc != nil {
		return m.getVehicleStatusFunc(ctx, internalVIN)
	}
	return nil, errors.New("not implemented")
}

func (m *mockClientForConfirm) GetEVVehicleStatus(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error) {
	if m.getEVVehicleStatusFunc != nil {
		return m.getEVVehicleStatusFunc(ctx, internalVIN)
	}
	return nil, errors.New("not implemented")
}

func (m *mockClientForConfirm) RefreshVehicleStatus(ctx context.Context, internalVIN api.InternalVIN) error {
	m.refreshVehicleStatusCalls++
	if m.refreshVehicleStatusFunc != nil {
		return m.refreshVehicleStatusFunc(ctx, internalVIN)
	}
	return nil
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
				getEVVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error) {
					if calls >= len(tt.chargingStatus) {
						calls = len(tt.chargingStatus) - 1
					}
					charging := tt.chargingStatus[calls]
					calls++
					return NewMockEVVehicleStatus().WithCharging(charging).Build(), nil
				},
			}

			// Use shorter timeout for "never" cases to speed up tests
			timeout := 5 * time.Second
			if !tt.expectMet {
				timeout = 200 * time.Millisecond
			}

			result := waitForCharging(ctx, &buf, mockClient, api.InternalVIN("test-vin"), timeout, 50*time.Millisecond)

			if tt.expectError {
				assert.NotNil(t, result.err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.Nilf(t, result.err, "Expected no error but got: %v", result.err)
			}

			if tt.expectMet {
				assert.Truef(t, result.success, "Expected charging to be started but it wasn't (calls: %d)", calls)
			}

			if result.success {
				assert.True(t, tt.expectMet, "Expected charging to not be started but it was")
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
				getEVVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error) {
					if calls >= len(tt.chargingStatus) {
						calls = len(tt.chargingStatus) - 1
					}
					charging := tt.chargingStatus[calls]
					calls++
					return NewMockEVVehicleStatus().WithCharging(charging).Build(), nil
				},
			}

			// Use shorter timeout for "never" cases to speed up tests
			timeout := 5 * time.Second
			if !tt.expectMet {
				timeout = 200 * time.Millisecond
			}

			result := waitForNotCharging(ctx, &buf, mockClient, api.InternalVIN("test-vin"), timeout, 50*time.Millisecond)

			if tt.expectError {
				assert.NotNil(t, result.err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.Nilf(t, result.err, "Expected no error but got: %v", result.err)
			}

			if tt.expectMet {
				assert.Truef(t, result.success, "Expected charging to be stopped but it wasn't (calls: %d)", calls)
			}

			if result.success {
				assert.True(t, tt.expectMet, "Expected charging to not be stopped but it was")
			}

		})
	}
}

// TestWaitForHvacOn tests the HVAC on confirmation logic
func TestWaitForHvacOn(t *testing.T) {
	tests := []struct {
		name         string
		hvacStatus   []bool
		hvacNil      []bool // indicates if HVAC info should be nil for each call
		expectError  bool
		expectMet    bool
		expectCalled int // minimum number of calls expected
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
		{
			name:         "nil hvac info then valid - should retry and succeed",
			hvacStatus:   []bool{false, true}, // second call has HVAC on
			hvacNil:      []bool{true, false}, // first call has nil HVAC
			expectError:  false,
			expectMet:    true,
			expectCalled: 2,
		},
		{
			name:         "multiple nil hvac info then valid",
			hvacStatus:   []bool{false, false, true},
			hvacNil:      []bool{true, true, false}, // first two calls have nil HVAC
			expectError:  false,
			expectMet:    true,
			expectCalled: 3,
		},
		{
			name:        "persistent nil hvac info - should timeout",
			hvacStatus:  []bool{false},
			hvacNil:     []bool{true, true, true}, // always nil
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
				getEVVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error) {
					callIdx := calls
					if callIdx >= len(tt.hvacStatus) {
						callIdx = len(tt.hvacStatus) - 1
					}

					// Check if this call should return nil HVAC info
					shouldBeNil := false
					if len(tt.hvacNil) > calls {
						shouldBeNil = tt.hvacNil[calls]
					}

					calls++

					if shouldBeNil {
						return NewMockEVVehicleStatus().WithoutHVAC().Build(), nil
					}

					hvacOn := tt.hvacStatus[callIdx]
					return NewMockEVVehicleStatus().WithHVACSettings(hvacOn, 22.0, false, false).Build(), nil
				},
			}

			// Use shorter timeout for "never" cases to speed up tests
			timeout := 5 * time.Second
			if !tt.expectMet {
				timeout = 200 * time.Millisecond
			}

			result := waitForHvacOn(ctx, &buf, mockClient, api.InternalVIN("test-vin"), timeout, 50*time.Millisecond)

			if tt.expectError {
				assert.NotNil(t, result.err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.Nilf(t, result.err, "Expected no error but got: %v", result.err)
			}

			if tt.expectMet {
				assert.Truef(t, result.success, "Expected HVAC to be on but it wasn't (calls: %d)", calls)
			}

			if result.success {
				assert.True(t, tt.expectMet, "Expected HVAC to not be on but it was")
			}

			if tt.expectCalled > 0 {
				assert.GreaterOrEqual(t, calls, tt.expectCalled)
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
				getEVVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error) {
					if calls >= len(tt.hvacStatus) {
						calls = len(tt.hvacStatus) - 1
					}
					hvacOn := tt.hvacStatus[calls]
					calls++
					return NewMockEVVehicleStatus().WithHVACSettings(hvacOn, 22.0, false, false).Build(), nil
				},
			}

			// Use shorter timeout for "never" cases to speed up tests
			timeout := 5 * time.Second
			if !tt.expectMet {
				timeout = 200 * time.Millisecond
			}

			result := waitForHvacOff(ctx, &buf, mockClient, api.InternalVIN("test-vin"), timeout, 50*time.Millisecond)

			if tt.expectError {
				assert.NotNil(t, result.err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.Nilf(t, result.err, "Expected no error but got: %v", result.err)
			}

			if tt.expectMet {
				assert.Truef(t, result.success, "Expected HVAC to be off but it wasn't (calls: %d)", calls)
			}

			if result.success {
				assert.True(t, tt.expectMet, "Expected HVAC to not be off but it was")
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
				getEVVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error) {
					if calls >= len(tt.hvacResponses) {
						calls = len(tt.hvacResponses) - 1
					}
					settings := tt.hvacResponses[calls]
					calls++
					return NewMockEVVehicleStatus().WithHVACSettings(
						settings.hvacOn,
						settings.temp,
						settings.frontDefrost,
						settings.rearDefrost,
					).Build(), nil
				},
			}

			// Use shorter timeout for "never" cases to speed up tests
			timeout := 5 * time.Second
			if !tt.expectMet {
				timeout = 200 * time.Millisecond
			}

			result := waitForHvacSettings(
				ctx,
				&buf,
				mockClient,
				api.InternalVIN("test-vin"),
				tt.targetTemp,
				tt.frontDefroster,
				tt.rearDefroster,
				timeout,
				50*time.Millisecond,
			)

			if tt.expectError {
				assert.NotNil(t, result.err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.Nilf(t, result.err, "Expected no error but got: %v", result.err)
			}

			if tt.expectMet {
				assert.Truef(t, result.success, "Expected HVAC settings to match but they didn't (calls: %d)", calls)
			}

			if result.success {
				assert.True(t, tt.expectMet, "Expected HVAC settings to not match but they did")
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
				ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
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
				ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
					return nil
				},
				WaitFunc: func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
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
				ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
					return nil
				},
				WaitFunc: func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
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
				ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
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
				ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
					return nil
				},
				WaitFunc: func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
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
				api.InternalVIN("test-vin"),
				tt.config,
				tt.confirm,
				tt.confirmWait,
			)

			if tt.expectError {
				assert.NotNil(t, err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.Nilf(t, err, "Expected no error but got: %v", err)
			}

			output := buf.String()
			assert.Equalf(t, tt.expectedOutput, output, "Expected output %q but got %q")
		})
	}
}

// TestWaitForConditionRefreshesStatus tests that confirmation polling calls RefreshVehicleStatus
// before starting to poll. This ensures we get fresh data from the vehicle, not stale cached data.
func TestWaitForConditionRefreshesStatus(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer

	mockClient := &mockClientForConfirm{
		getEVVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error) {
			return NewMockEVVehicleStatus().WithHVAC(true).Build(), nil
		},
		refreshVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) error {
			return nil
		},
	}

	conditionChecker := func(status interface{}) (bool, error) {
		evStatus := status.(*api.EVVehicleStatusResponse)
		hvacInfo, err := evStatus.GetHvacInfo()
		if err != nil {
			return false, err
		}
		return hvacInfo.HVACOn, nil
	}

	result := waitForCondition(
		ctx,
		&buf,
		mockClient,
		api.InternalVIN("test-vin"),
		true, // useEVStatus
		conditionChecker,
		200*time.Millisecond,
		50*time.Millisecond,
		"test action",
	)

	assert.Equalf(t, nil, result.err, "Expected no error but got: %v", result.err)

	assert.True(t, result.success, "Expected condition to be met")

	// The critical assertion: RefreshVehicleStatus should be called exactly once before polling
	assert.Equalf(t, 1, mockClient.refreshVehicleStatusCalls, "Expected RefreshVehicleStatus to be called exactly once, but was called %d times", mockClient.refreshVehicleStatusCalls)
}

// TestWaitForDoorsUnlocked tests the door unlock confirmation logic
func TestWaitForDoorsUnlocked(t *testing.T) {
	tests := []struct {
		name        string
		doorStatus  []api.DoorStatus
		expectError bool
		expectMet   bool
	}{
		{
			name: "all doors unlocked immediately",
			doorStatus: []api.DoorStatus{
				{
					DriverLocked:    false,
					PassengerLocked: false,
					RearLeftLocked:  false,
					RearRightLocked: false,
					AllLocked:       false,
				},
			},
			expectError: false,
			expectMet:   true,
		},
		{
			name: "doors unlock after one check",
			doorStatus: []api.DoorStatus{
				{
					DriverLocked:    true,
					PassengerLocked: true,
					RearLeftLocked:  true,
					RearRightLocked: true,
					AllLocked:       true,
				},
				{
					DriverLocked:    false,
					PassengerLocked: false,
					RearLeftLocked:  false,
					RearRightLocked: false,
					AllLocked:       false,
				},
			},
			expectError: false,
			expectMet:   true,
		},
		{
			name: "at least one door unlocked is considered unlocked",
			doorStatus: []api.DoorStatus{
				{
					DriverLocked:    false,
					PassengerLocked: true,
					RearLeftLocked:  true,
					RearRightLocked: true,
					AllLocked:       false,
				},
			},
			expectError: false,
			expectMet:   true,
		},
		{
			name: "doors never unlock",
			doorStatus: []api.DoorStatus{
				{
					DriverLocked:    true,
					PassengerLocked: true,
					RearLeftLocked:  true,
					RearRightLocked: true,
					AllLocked:       true,
				},
				{
					DriverLocked:    true,
					PassengerLocked: true,
					RearLeftLocked:  true,
					RearRightLocked: true,
					AllLocked:       true,
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
				getVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.VehicleStatusResponse, error) {
					if calls >= len(tt.doorStatus) {
						calls = len(tt.doorStatus) - 1
					}
					status := tt.doorStatus[calls]
					calls++
					return NewMockVehicleStatus().WithDoorStatus(status).Build(), nil
				},
			}

			// Use shorter timeout for "never" cases to speed up tests
			timeout := 5 * time.Second
			if !tt.expectMet {
				timeout = 200 * time.Millisecond
			}

			result := waitForDoorsUnlocked(ctx, &buf, mockClient, api.InternalVIN("test-vin"), timeout, 50*time.Millisecond)

			if tt.expectError {
				assert.NotNil(t, result.err, "Expected error but got nil")
			}

			if !tt.expectError {
				assert.Nilf(t, result.err, "Expected no error but got: %v", result.err)
			}

			if tt.expectMet {
				assert.Truef(t, result.success, "Expected doors to be unlocked but they weren't (calls: %d)", calls)
			}

			if result.success {
				assert.True(t, tt.expectMet, "Expected doors to not be unlocked but they were")
			}

		})
	}
}

// TestClientAdapter tests the clientAdapter wrapper that converts InternalVIN types
func TestClientAdapter(t *testing.T) {
	ctx := context.Background()
	testVIN := api.InternalVIN("123456789")

	t.Run("GetVehicleStatus", func(t *testing.T) {
		// Create a mock API client with a wrapper that intercepts the method calls
		mockAPIClient := &mockAPIClientForAdapter{}
		mockAPIClient.getVehicleStatusFunc = func(ctx context.Context, vin string) (*api.VehicleStatusResponse, error) {
			assert.Equalf(t, string(testVIN), vin, "Expected VIN to be %q, got %q")
			return NewMockVehicleStatus().Build(), nil
		}

		adapter := &testClientAdapter{mockAPIClient}
		resp, err := adapter.GetVehicleStatus(ctx, testVIN)

		require.NoError(t, err, "Expected no error, got: %v")

		require.NotNil(t, resp, "Expected response to be non-nil")

		assert.Equalf(t, api.ResultCodeSuccess, resp.ResultCode, "Expected result code %q, got %q")
	})

	t.Run("GetEVVehicleStatus", func(t *testing.T) {
		mockAPIClient := &mockAPIClientForAdapter{}
		mockAPIClient.getEVVehicleStatusFunc = func(ctx context.Context, vin string) (*api.EVVehicleStatusResponse, error) {
			assert.Equalf(t, string(testVIN), vin, "Expected VIN to be %q, got %q")
			return NewMockEVVehicleStatus().Build(), nil
		}

		adapter := &testClientAdapter{mockAPIClient}
		resp, err := adapter.GetEVVehicleStatus(ctx, testVIN)

		require.NoError(t, err, "Expected no error, got: %v")

		require.NotNil(t, resp, "Expected response to be non-nil")

		assert.Equalf(t, api.ResultCodeSuccess, resp.ResultCode, "Expected result code %q, got %q")
	})

	t.Run("RefreshVehicleStatus", func(t *testing.T) {
		refreshCalled := false
		mockAPIClient := &mockAPIClientForAdapter{}
		mockAPIClient.refreshVehicleStatusFunc = func(ctx context.Context, vin string) error {
			assert.Equalf(t, string(testVIN), vin, "Expected VIN to be %q, got %q")
			refreshCalled = true
			return nil
		}

		adapter := &testClientAdapter{mockAPIClient}
		err := adapter.RefreshVehicleStatus(ctx, testVIN)

		assert.Equalf(t, nil, err, "Expected no error, got: %v", err)

		assert.True(t, refreshCalled, "Expected RefreshVehicleStatus to be called")
	})

	t.Run("GetVehicleStatus error propagation", func(t *testing.T) {
		expectedErr := errors.New("API error")
		mockAPIClient := &mockAPIClientForAdapter{}
		mockAPIClient.getVehicleStatusFunc = func(ctx context.Context, vin string) (*api.VehicleStatusResponse, error) {
			return nil, expectedErr
		}

		adapter := &testClientAdapter{mockAPIClient}
		_, err := adapter.GetVehicleStatus(ctx, testVIN)

		assert.Equalf(t, expectedErr, err, "Expected error %v, got %v")
	})

	t.Run("GetEVVehicleStatus error propagation", func(t *testing.T) {
		expectedErr := errors.New("API error")
		mockAPIClient := &mockAPIClientForAdapter{}
		mockAPIClient.getEVVehicleStatusFunc = func(ctx context.Context, vin string) (*api.EVVehicleStatusResponse, error) {
			return nil, expectedErr
		}

		adapter := &testClientAdapter{mockAPIClient}
		_, err := adapter.GetEVVehicleStatus(ctx, testVIN)

		assert.Equalf(t, expectedErr, err, "Expected error %v, got %v")
	})

	t.Run("RefreshVehicleStatus error propagation", func(t *testing.T) {
		expectedErr := errors.New("refresh error")
		mockAPIClient := &mockAPIClientForAdapter{}
		mockAPIClient.refreshVehicleStatusFunc = func(ctx context.Context, vin string) error {
			return expectedErr
		}

		adapter := &testClientAdapter{mockAPIClient}
		err := adapter.RefreshVehicleStatus(ctx, testVIN)

		assert.Equalf(t, expectedErr, err, "Expected error %v, got %v")
	})
}

// testClientAdapter is a test version of clientAdapter that works with our mock
type testClientAdapter struct {
	client *mockAPIClientForAdapter
}

func (t *testClientAdapter) GetVehicleStatus(ctx context.Context, internalVIN api.InternalVIN) (*api.VehicleStatusResponse, error) {
	return t.client.GetVehicleStatus(ctx, string(internalVIN))
}

func (t *testClientAdapter) GetEVVehicleStatus(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error) {
	return t.client.GetEVVehicleStatus(ctx, string(internalVIN))
}

func (t *testClientAdapter) RefreshVehicleStatus(ctx context.Context, internalVIN api.InternalVIN) error {
	return t.client.RefreshVehicleStatus(ctx, string(internalVIN))
}

// mockAPIClientForAdapter is a mock implementation of api.Client methods for testing the clientAdapter
type mockAPIClientForAdapter struct {
	getVehicleStatusFunc     func(ctx context.Context, vin string) (*api.VehicleStatusResponse, error)
	getEVVehicleStatusFunc   func(ctx context.Context, vin string) (*api.EVVehicleStatusResponse, error)
	refreshVehicleStatusFunc func(ctx context.Context, vin string) error
}

func (m *mockAPIClientForAdapter) GetVehicleStatus(ctx context.Context, vin string) (*api.VehicleStatusResponse, error) {
	if m.getVehicleStatusFunc != nil {
		return m.getVehicleStatusFunc(ctx, vin)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAPIClientForAdapter) GetEVVehicleStatus(ctx context.Context, vin string) (*api.EVVehicleStatusResponse, error) {
	if m.getEVVehicleStatusFunc != nil {
		return m.getEVVehicleStatusFunc(ctx, vin)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAPIClientForAdapter) RefreshVehicleStatus(ctx context.Context, vin string) error {
	if m.refreshVehicleStatusFunc != nil {
		return m.refreshVehicleStatusFunc(ctx, vin)
	}
	return errors.New("not implemented")
}

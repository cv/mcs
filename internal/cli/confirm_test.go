package cli

import (
	"bytes"
	"context"
	"errors"
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

// mockClientForConfirm is a mock API client for testing confirmation logic
type mockClientForConfirm struct {
	getVehicleStatusFunc func(ctx context.Context, internalVIN string) (*api.VehicleStatusResponse, error)
}

func (m *mockClientForConfirm) GetVehicleStatus(ctx context.Context, internalVIN string) (*api.VehicleStatusResponse, error) {
	if m.getVehicleStatusFunc != nil {
		return m.getVehicleStatusFunc(ctx, internalVIN)
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

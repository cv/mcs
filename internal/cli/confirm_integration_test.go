package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/cv/mcs/internal/api"
)

// TestConfirmationFlow_Integration tests the full confirmation flow end-to-end
func TestConfirmationFlow_Integration(t *testing.T) {
	tests := []struct {
		name              string
		actionError       error
		statusSequence    []api.DoorStatus
		confirm           bool
		confirmWait       int
		expectError       bool
		expectedOutput    string
		verifyStatusCalls bool
		minStatusCalls    int
	}{
		{
			name:        "success with immediate confirmation",
			actionError: nil,
			statusSequence: []api.DoorStatus{
				{AllLocked: true, DriverLocked: true, PassengerLocked: true, RearLeftLocked: true, RearRightLocked: true},
			},
			confirm:           true,
			confirmWait:       1,
			expectError:       false,
			expectedOutput:    "Command sent, waiting for confirmation...\nCommand executed successfully\n",
			verifyStatusCalls: true,
			minStatusCalls:    1,
		},
		{
			name:        "success after status check",
			actionError: nil,
			statusSequence: []api.DoorStatus{
				{AllLocked: false, DriverLocked: false, PassengerLocked: true, RearLeftLocked: true, RearRightLocked: true},
				{AllLocked: true, DriverLocked: true, PassengerLocked: true, RearLeftLocked: true, RearRightLocked: true},
			},
			confirm:           true,
			confirmWait:       2,
			expectError:       false,
			expectedOutput:    "Command executed successfully",
			verifyStatusCalls: true,
			minStatusCalls:    2,
		},
		{
			name:        "timeout during confirmation",
			actionError: nil,
			statusSequence: []api.DoorStatus{
				{AllLocked: false, DriverLocked: false, PassengerLocked: true, RearLeftLocked: true, RearRightLocked: true},
				{AllLocked: false, DriverLocked: false, PassengerLocked: true, RearLeftLocked: true, RearRightLocked: true},
			},
			confirm:        true,
			confirmWait:    1,
			expectError:    false,
			expectedOutput: "Warning: test not confirmed within timeout period",
		},
		{
			name:           "no confirmation",
			actionError:    nil,
			statusSequence: []api.DoorStatus{},
			confirm:        false,
			confirmWait:    90,
			expectError:    false,
			expectedOutput: "Command executed successfully\n",
		},
		{
			name:           "action fails",
			actionError:    errors.New("vehicle is offline"),
			statusSequence: []api.DoorStatus{},
			confirm:        true,
			confirmWait:    90,
			expectError:    true,
			expectedOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer

			// Track action and status calls
			actionCalled := false
			statusCallCount := 0

			// Create mock action function
			actionFunc := func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
				actionCalled = true
				return tt.actionError
			}

			// Create mock wait function
			var waitFunc func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult
			if tt.confirm && len(tt.statusSequence) > 0 {
				waitFunc = func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
					// Create mock client for status checks
					mockClient := &mockClientForConfirm{
						getVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.VehicleStatusResponse, error) {
							if statusCallCount >= len(tt.statusSequence) {
								statusCallCount = len(tt.statusSequence) - 1
							}
							status := tt.statusSequence[statusCallCount]
							statusCallCount++
							return NewMockVehicleStatus().WithDoorStatus(status).Build(), nil
						},
					}

					conditionChecker := func(status interface{}) (bool, error) {
						vStatus := status.(*api.VehicleStatusResponse)
						doorStatus, err := vStatus.GetDoorsInfo()
						if err != nil {
							return false, err
						}
						return doorStatus.AllLocked, nil
					}

					return waitForCondition(ctx, out, mockClient, internalVIN, false, conditionChecker, timeout, pollInterval, "test")
				}
			}

			// Create config
			config := ConfirmableCommandConfig{
				ActionFunc:    actionFunc,
				WaitFunc:      waitFunc,
				PollInterval:  100 * time.Millisecond, // fast polling for tests
				SuccessMsg:    "Command executed successfully",
				WaitingMsg:    "Command sent, waiting for confirmation...",
				ActionName:    "test action",
				ConfirmName:   "test status",
				TimeoutSuffix: "confirmation timeout",
			}

			// Execute confirmable command
			err := executeConfirmableCommand(ctx, &buf, nil, api.InternalVIN("TEST-VIN"), config, tt.confirm, tt.confirmWait)

			// Verify error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Verify action was called (unless we expected an error before action)
			if !tt.expectError && !actionCalled {
				t.Error("Expected action to be called")
			}

			// Verify status call count if specified
			if tt.verifyStatusCalls && statusCallCount < tt.minStatusCalls {
				t.Errorf("Expected at least %d status calls, got %d", tt.minStatusCalls, statusCallCount)
			}

			// Verify output contains expected messages
			output := buf.String()
			if tt.expectedOutput != "" && !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("Expected output to contain %q but got %q", tt.expectedOutput, output)
			}
		})
	}
}

// TestConfirmationFlow_EVStatus tests confirmation flow with EV status
func TestConfirmationFlow_EVStatus(t *testing.T) {
	tests := []struct {
		name           string
		hvacSequence   []bool
		confirm        bool
		confirmWait    int
		expectSuccess  bool
		expectedOutput string
	}{
		{
			name:           "HVAC turns on immediately",
			hvacSequence:   []bool{true},
			confirm:        true,
			confirmWait:    1,
			expectSuccess:  true,
			expectedOutput: "HVAC command sent, waiting for confirmation...\nHVAC turned on successfully\n",
		},
		{
			name:           "HVAC turns on after check",
			hvacSequence:   []bool{false, true},
			confirm:        true,
			confirmWait:    2,
			expectSuccess:  true,
			expectedOutput: "HVAC turned on successfully",
		},
		{
			name:           "timeout",
			hvacSequence:   []bool{false, false},
			confirm:        true,
			confirmWait:    1,
			expectSuccess:  true,
			expectedOutput: "Warning: HVAC not confirmed within timeout period",
		},
		{
			name:           "no confirmation",
			hvacSequence:   []bool{},
			confirm:        false,
			confirmWait:    90,
			expectSuccess:  true,
			expectedOutput: "HVAC turned on successfully\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			var buf bytes.Buffer

			statusCallCount := 0

			// Create mock action function
			actionFunc := func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
				return nil
			}

			// Create mock wait function
			var waitFunc func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult
			if tt.confirm && len(tt.hvacSequence) > 0 {
				waitFunc = func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
					// Create mock client for status checks
					mockClient := &mockClientForConfirm{
						getEVVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error) {
							if statusCallCount >= len(tt.hvacSequence) {
								statusCallCount = len(tt.hvacSequence) - 1
							}
							hvacOn := tt.hvacSequence[statusCallCount]
							statusCallCount++
							return NewMockEVVehicleStatus().WithHVAC(hvacOn).Build(), nil
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

					return waitForCondition(ctx, out, mockClient, internalVIN, true, conditionChecker, timeout, pollInterval, "HVAC")
				}
			}

			// Create config
			config := ConfirmableCommandConfig{
				ActionFunc:    actionFunc,
				WaitFunc:      waitFunc,
				PollInterval:  100 * time.Millisecond, // fast polling for tests
				SuccessMsg:    "HVAC turned on successfully",
				WaitingMsg:    "HVAC command sent, waiting for confirmation...",
				ActionName:    "turn on HVAC",
				ConfirmName:   "HVAC status",
				TimeoutSuffix: "confirmation timeout",
			}

			// Execute confirmable command
			err := executeConfirmableCommand(ctx, &buf, nil, api.InternalVIN("TEST-VIN"), config, tt.confirm, tt.confirmWait)

			// Verify error expectation
			if !tt.expectSuccess && err == nil {
				t.Error("Expected error but got nil")
			}

			if tt.expectSuccess && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Verify output contains expected messages
			output := buf.String()
			if tt.expectedOutput != "" && !strings.Contains(output, tt.expectedOutput) {
				t.Errorf("Expected output to contain %q but got %q", tt.expectedOutput, output)
			}
		})
	}
}

// TestConfirmationFlow_StatusError tests error handling during status check
func TestConfirmationFlow_StatusError(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer

	// Create mock action function
	actionFunc := func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
		return nil
	}

	// Create mock wait function that returns error
	waitFunc := func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
		// Create mock client that returns error
		mockClient := &mockClientForConfirm{
			getVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.VehicleStatusResponse, error) {
				return nil, errors.New("network error")
			},
		}

		conditionChecker := func(status interface{}) (bool, error) {
			vStatus := status.(*api.VehicleStatusResponse)
			doorStatus, err := vStatus.GetDoorsInfo()
			if err != nil {
				return false, err
			}
			return doorStatus.AllLocked, nil
		}

		return waitForCondition(ctx, out, mockClient, internalVIN, false, conditionChecker, 1*time.Second, 200*time.Millisecond, "lock")
	}

	// Create config
	config := ConfirmableCommandConfig{
		ActionFunc:    actionFunc,
		WaitFunc:      waitFunc,
		PollInterval:  100 * time.Millisecond, // fast polling for tests
		SuccessMsg:    "Doors locked successfully",
		WaitingMsg:    "Lock command sent, waiting for confirmation...",
		ActionName:    "lock doors",
		ConfirmName:   "lock status",
		TimeoutSuffix: "confirmation timeout",
	}

	// Execute confirmable command with short timeout to avoid long test
	err := executeConfirmableCommand(ctx, &buf, nil, api.InternalVIN("TEST-VIN"), config, true, 1)

	// Verify: No error should be returned (errors are treated as "not ready yet" and timeout)
	if err != nil {
		t.Errorf("Expected no error (timeout), got: %v", err)
	}

	// Verify: Output should contain timeout message
	output := buf.String()
	if !strings.Contains(output, "Lock command sent") {
		t.Errorf("Expected timeout message in output, got: %q", output)
	}
}

// TestConfirmationFlow_ContextCancellation tests confirmation flow with context cancellation
func TestConfirmationFlow_ContextCancellation(t *testing.T) {
	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	var buf bytes.Buffer

	// Create mock action function
	actionFunc := func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
		return nil
	}

	// Create mock wait function that checks for cancellation
	waitFunc := func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
		// Cancel context to simulate cancellation
		cancel()

		// Create mock client
		mockClient := &mockClientForConfirm{
			getVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.VehicleStatusResponse, error) {
				// Check if context is cancelled
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				default:
					return NewMockVehicleStatus().WithDoorStatus(api.DoorStatus{AllLocked: false}).Build(), nil
				}
			},
		}

		conditionChecker := func(status interface{}) (bool, error) {
			vStatus := status.(*api.VehicleStatusResponse)
			doorStatus, err := vStatus.GetDoorsInfo()
			if err != nil {
				return false, err
			}
			return doorStatus.AllLocked, nil
		}

		return waitForCondition(ctx, out, mockClient, internalVIN, false, conditionChecker, 1*time.Second, 200*time.Millisecond, "lock")
	}

	// Create config
	config := ConfirmableCommandConfig{
		ActionFunc:    actionFunc,
		WaitFunc:      waitFunc,
		PollInterval:  100 * time.Millisecond, // fast polling for tests
		SuccessMsg:    "Doors locked successfully",
		WaitingMsg:    "Lock command sent, waiting for confirmation...",
		ActionName:    "lock doors",
		ConfirmName:   "lock status",
		TimeoutSuffix: "confirmation timeout",
	}

	// Execute confirmable command
	err := executeConfirmableCommand(ctx, &buf, nil, api.InternalVIN("TEST-VIN"), config, true, 1)

	// Verify: Error should be returned
	if err == nil {
		t.Error("Expected error when context is cancelled, got nil")
	}
}

// TestConfirmationFlow_MultipleConditions tests confirmation with complex conditions
func TestConfirmationFlow_MultipleConditions(t *testing.T) {
	ctx := context.Background()
	var buf bytes.Buffer

	hvacSettings := []struct {
		on            bool
		temp          float64
		frontDefrost  bool
		rearDefrost   bool
		shouldSucceed bool
	}{
		{on: true, temp: 20.0, frontDefrost: false, rearDefrost: false, shouldSucceed: false},
		{on: true, temp: 22.0, frontDefrost: true, rearDefrost: false, shouldSucceed: true},
	}

	statusCallCount := 0

	// Create mock action function
	actionFunc := func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
		return nil
	}

	// Create mock wait function
	waitFunc := func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
		// Create mock client for status checks
		mockClient := &mockClientForConfirm{
			getEVVehicleStatusFunc: func(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error) {
				if statusCallCount >= len(hvacSettings) {
					statusCallCount = len(hvacSettings) - 1
				}
				settings := hvacSettings[statusCallCount]
				statusCallCount++
				return NewMockEVVehicleStatus().WithHVACSettings(
					settings.on,
					settings.temp,
					settings.frontDefrost,
					settings.rearDefrost,
				).Build(), nil
			},
		}

		return waitForHvacSettings(ctx, out, mockClient, internalVIN, 22.0, true, false, timeout, pollInterval)
	}

	// Create config
	config := ConfirmableCommandConfig{
		ActionFunc:    actionFunc,
		WaitFunc:      waitFunc,
		PollInterval:  100 * time.Millisecond, // fast polling for tests
		SuccessMsg:    "HVAC settings updated successfully",
		WaitingMsg:    "HVAC settings command sent, waiting for confirmation...",
		ActionName:    "update HVAC settings",
		ConfirmName:   "HVAC settings status",
		TimeoutSuffix: "confirmation timeout",
	}

	// Execute confirmable command
	err := executeConfirmableCommand(ctx, &buf, nil, api.InternalVIN("TEST-VIN"), config, true, 2)

	// Verify: No error
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}

	// Verify: Status was called at least twice
	if statusCallCount < 2 {
		t.Errorf("Expected at least 2 status calls, got %d", statusCallCount)
	}

	// Verify: Success message in output
	output := buf.String()
	if !strings.Contains(output, "HVAC settings updated successfully") {
		t.Errorf("Expected output to contain success message, got: %q", output)
	}
}

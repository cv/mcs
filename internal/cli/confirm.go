package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cv/mcs/internal/api"
)

// confirmationResult holds the result of a confirmation poll
type confirmationResult struct {
	success bool
	err     error
}

// pollUntilCondition polls a condition function until it returns true or times out.
// It returns a result indicating success or timeout, and any error encountered.
func pollUntilCondition(
	ctx context.Context,
	out io.Writer,
	checkFunc func() (bool, error),
	timeout time.Duration,
	pollInterval time.Duration,
	actionName string,
) confirmationResult {
	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	startTime := time.Now()

	// Check immediately first
	if met, err := checkFunc(); err != nil {
		return confirmationResult{success: false, err: err}
	} else if met {
		return confirmationResult{success: true, err: nil}
	}

	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(startTime)
			_, _ = fmt.Fprintf(out, "Waiting for confirmation... (%ds/%ds)\n",
				int(elapsed.Seconds()), int(timeout.Seconds()))

			met, err := checkFunc()
			if err != nil {
				return confirmationResult{success: false, err: err}
			}
			if met {
				return confirmationResult{success: true, err: nil}
			}

		case <-timeoutCtx.Done():
			if timeoutCtx.Err() == context.DeadlineExceeded {
				_, _ = fmt.Fprintf(out, "Warning: %s not confirmed within timeout period\n", actionName)
				return confirmationResult{success: false, err: nil}
			}
			return confirmationResult{success: false, err: timeoutCtx.Err()}
		}
	}
}

// vehicleStatusGetter is an interface for getting vehicle status
// This allows for easier testing by mocking the API client
type vehicleStatusGetter interface {
	GetVehicleStatus(ctx context.Context, internalVIN string) (*api.VehicleStatusResponse, error)
	GetEVVehicleStatus(ctx context.Context, internalVIN string) (*api.EVVehicleStatusResponse, error)
}

// waitForDoorsLocked polls the vehicle status until all doors are locked or timeout occurs
func waitForDoorsLocked(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN string,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	checkFunc := func() (bool, error) {
		status, err := client.GetVehicleStatus(ctx, internalVIN)
		if err != nil {
			return false, err
		}

		doorStatus, err := status.GetDoorsInfo()
		if err != nil {
			return false, err
		}

		return doorStatus.AllLocked, nil
	}

	return pollUntilCondition(ctx, out, checkFunc, timeout, pollInterval, "door lock")
}

// waitForDoorsUnlocked polls the vehicle status until all doors are unlocked or timeout occurs
func waitForDoorsUnlocked(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN string,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	checkFunc := func() (bool, error) {
		status, err := client.GetVehicleStatus(ctx, internalVIN)
		if err != nil {
			return false, err
		}

		doorStatus, err := status.GetDoorsInfo()
		if err != nil {
			return false, err
		}

		// Unlocked means at least one door is unlocked (not all locked)
		return !doorStatus.AllLocked, nil
	}

	return pollUntilCondition(ctx, out, checkFunc, timeout, pollInterval, "door unlock")
}

// waitForEngineRunning polls the vehicle status until the engine is running or timeout occurs
func waitForEngineRunning(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN string,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	checkFunc := func() (bool, error) {
		status, err := client.GetEVVehicleStatus(ctx, internalVIN)
		if err != nil {
			return false, err
		}

		hvacInfo, err := status.GetHvacInfo()
		if err != nil {
			return false, err
		}

		return hvacInfo.HVACOn, nil
	}

	return pollUntilCondition(ctx, out, checkFunc, timeout, pollInterval, "engine start")
}

// waitForEngineStopped polls the vehicle status until the engine is stopped or timeout occurs
func waitForEngineStopped(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN string,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	checkFunc := func() (bool, error) {
		status, err := client.GetEVVehicleStatus(ctx, internalVIN)
		if err != nil {
			return false, err
		}

		hvacInfo, err := status.GetHvacInfo()
		if err != nil {
			return false, err
		}

		return !hvacInfo.HVACOn, nil
	}

	return pollUntilCondition(ctx, out, checkFunc, timeout, pollInterval, "engine stop")
}

// waitForCharging polls the vehicle status until charging is active or timeout occurs
func waitForCharging(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN string,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	checkFunc := func() (bool, error) {
		status, err := client.GetEVVehicleStatus(ctx, internalVIN)
		if err != nil {
			return false, err
		}

		batteryInfo, err := status.GetBatteryInfo()
		if err != nil {
			return false, err
		}

		return batteryInfo.Charging, nil
	}

	return pollUntilCondition(ctx, out, checkFunc, timeout, pollInterval, "charging start")
}

// waitForNotCharging polls the vehicle status until charging is inactive or timeout occurs
func waitForNotCharging(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN string,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	checkFunc := func() (bool, error) {
		status, err := client.GetEVVehicleStatus(ctx, internalVIN)
		if err != nil {
			return false, err
		}

		batteryInfo, err := status.GetBatteryInfo()
		if err != nil {
			return false, err
		}

		return !batteryInfo.Charging, nil
	}

	return pollUntilCondition(ctx, out, checkFunc, timeout, pollInterval, "charging stop")
}

// waitForHvacOn polls the vehicle status until HVAC is on or timeout occurs
func waitForHvacOn(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN string,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	checkFunc := func() (bool, error) {
		status, err := client.GetEVVehicleStatus(ctx, internalVIN)
		if err != nil {
			return false, err
		}

		hvacInfo, err := status.GetHvacInfo()
		if err != nil {
			return false, err
		}

		return hvacInfo.HVACOn, nil
	}

	return pollUntilCondition(ctx, out, checkFunc, timeout, pollInterval, "HVAC on")
}

// waitForHvacOff polls the vehicle status until HVAC is off or timeout occurs
func waitForHvacOff(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN string,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	checkFunc := func() (bool, error) {
		status, err := client.GetEVVehicleStatus(ctx, internalVIN)
		if err != nil {
			return false, err
		}

		hvacInfo, err := status.GetHvacInfo()
		if err != nil {
			return false, err
		}

		return !hvacInfo.HVACOn, nil
	}

	return pollUntilCondition(ctx, out, checkFunc, timeout, pollInterval, "HVAC off")
}

// waitForHvacSettings polls the vehicle status until HVAC settings match the requested values or timeout occurs
func waitForHvacSettings(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN string,
	targetTemp float64,
	frontDefroster bool,
	rearDefroster bool,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	checkFunc := func() (bool, error) {
		status, err := client.GetEVVehicleStatus(ctx, internalVIN)
		if err != nil {
			return false, err
		}

		hvacInfo, err := status.GetHvacInfo()
		if err != nil {
			return false, err
		}

		// Check temperature with tolerance of 0.5C
		const tempTolerance = 0.5
		tempMatch := hvacInfo.TargetTempC >= targetTemp-tempTolerance &&
			hvacInfo.TargetTempC <= targetTemp+tempTolerance

		// Check defroster settings
		defrostersMatch := hvacInfo.FrontDefroster == frontDefroster &&
			hvacInfo.RearDefroster == rearDefroster

		return tempMatch && defrostersMatch, nil
	}

	return pollUntilCondition(ctx, out, checkFunc, timeout, pollInterval, "HVAC settings")
}

// ConfirmableCommandConfig holds the configuration for a confirmable command
type ConfirmableCommandConfig struct {
	// ActionFunc performs the API action (e.g., lock doors, start engine)
	ActionFunc func(ctx context.Context, client *api.Client, internalVIN string) error

	// WaitFunc waits for confirmation that the action completed
	// If nil, confirmation is skipped
	WaitFunc func(ctx context.Context, out io.Writer, client *api.Client, internalVIN string, timeout, pollInterval time.Duration) confirmationResult

	// Messages
	SuccessMsg    string // Message to show on success (e.g., "Doors locked successfully")
	WaitingMsg    string // Message to show while waiting (e.g., "Lock command sent, waiting for confirmation...")
	ActionName    string // Name for error messages (e.g., "lock doors")
	ConfirmName   string // Name for confirmation error (e.g., "lock status")
	TimeoutSuffix string // Suffix for timeout message (e.g., "confirmation timeout")
}

// executeConfirmableCommand executes a confirmable command with the given configuration
func executeConfirmableCommand(
	ctx context.Context,
	out io.Writer,
	client *api.Client,
	internalVIN string,
	config ConfirmableCommandConfig,
	confirm bool,
	confirmWait int,
) error {
	// Execute the action
	if err := config.ActionFunc(ctx, client, internalVIN); err != nil {
		return fmt.Errorf("failed to %s: %w", config.ActionName, err)
	}

	// If confirmation disabled, return immediately
	if !confirm || config.WaitFunc == nil {
		_, _ = fmt.Fprintln(out, config.SuccessMsg)
		return nil
	}

	// Wait for confirmation
	_, _ = fmt.Fprintln(out, config.WaitingMsg)
	result := config.WaitFunc(
		ctx,
		out,
		client,
		internalVIN,
		time.Duration(confirmWait)*time.Second,
		5*time.Second, // poll every 5 seconds
	)

	if result.err != nil {
		return fmt.Errorf("failed to confirm %s: %w", config.ConfirmName, result.err)
	}

	if result.success {
		_, _ = fmt.Fprintln(out, config.SuccessMsg)
	} else {
		// Extract the command part from waiting message (e.g., "Lock command sent" from "Lock command sent, waiting for confirmation...")
		// and append the timeout suffix
		commandMsg := config.WaitingMsg
		if idx := len(commandMsg) - len(", waiting for confirmation..."); idx > 0 {
			commandMsg = commandMsg[:idx]
		}
		_, _ = fmt.Fprintf(out, "%s (%s)\n", commandMsg, config.TimeoutSuffix)
	}

	return nil
}

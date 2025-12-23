package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cv/mcs/internal/api"
)

// confirmationResult holds the result of a confirmation poll.
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
		// Treat errors as "condition not yet met" - retry instead of failing immediately
		// This handles transient errors like nil HVAC info or temporary API issues
	} else if met {
		return confirmationResult{success: true, err: nil}
	}

	lastPrintedSecond := -1

	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(startTime)
			elapsedSec := int(elapsed.Seconds())

			// Only update output once per second to avoid spam
			if elapsedSec > lastPrintedSecond {
				lastPrintedSecond = elapsedSec
				// Use \r to update in place, then clear to end of line
				_, _ = fmt.Fprintf(out, "\rWaiting for confirmation... (%ds/%ds)   ",
					elapsedSec, int(timeout.Seconds()))
			}

			met, err := checkFunc()
			if err != nil {
				// Treat errors as "condition not yet met" - continue polling
				// This allows recovery from transient errors
				continue
			}
			if met {
				// Clear the progress line and move to new line
				_, _ = fmt.Fprint(out, "\r                                        \r")

				return confirmationResult{success: true, err: nil}
			}

		case <-timeoutCtx.Done():
			// Clear the progress line and move to new line
			_, _ = fmt.Fprint(out, "\r                                        \r")
			if timeoutCtx.Err() == context.DeadlineExceeded {
				_, _ = fmt.Fprintf(out, "Warning: %s not confirmed within timeout period\n", actionName)

				return confirmationResult{success: false, err: nil}
			}

			return confirmationResult{success: false, err: timeoutCtx.Err()}
		}
	}
}

// vehicleStatusGetter is an interface for getting vehicle status
// This allows for easier testing by mocking the API client.
type vehicleStatusGetter interface {
	GetVehicleStatus(ctx context.Context, internalVIN api.InternalVIN) (*api.VehicleStatusResponse, error)
	GetEVVehicleStatus(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error)
	RefreshVehicleStatus(ctx context.Context, internalVIN api.InternalVIN) error
}

// clientAdapter adapts api.Client to vehicleStatusGetter by converting InternalVIN to string.
type clientAdapter struct {
	*api.Client
}

func (c *clientAdapter) GetVehicleStatus(ctx context.Context, internalVIN api.InternalVIN) (*api.VehicleStatusResponse, error) {
	return c.Client.GetVehicleStatus(ctx, string(internalVIN))
}

func (c *clientAdapter) GetEVVehicleStatus(ctx context.Context, internalVIN api.InternalVIN) (*api.EVVehicleStatusResponse, error) {
	return c.Client.GetEVVehicleStatus(ctx, string(internalVIN))
}

func (c *clientAdapter) RefreshVehicleStatus(ctx context.Context, internalVIN api.InternalVIN) error {
	return c.Client.RefreshVehicleStatus(ctx, string(internalVIN))
}

// waitForCondition is a generic function that waits for a vehicle status condition to be met.
// It polls the vehicle status (either regular or EV) and checks the condition using the provided checker function.
//
// Parameters:
//   - ctx: context for cancellation
//   - out: writer for status messages
//   - client: API client for getting vehicle status
//   - internalVIN: vehicle identifier
//   - useEVStatus: if true, uses GetEVVehicleStatus; otherwise uses GetVehicleStatus
//   - conditionChecker: function that receives the status response and returns true if condition is met
//   - timeout: maximum time to wait for condition
//   - pollInterval: time between status checks
//   - actionName: name of the action being confirmed (for error messages)
//
// Returns: confirmationResult with success flag and any error encountered.
func waitForCondition(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN api.InternalVIN,
	useEVStatus bool,
	conditionChecker func(any) (bool, error),
	timeout time.Duration,
	pollInterval time.Duration,
	actionName string,
) confirmationResult {
	// Request fresh status from vehicle before polling
	if err := client.RefreshVehicleStatus(ctx, internalVIN); err != nil {
		// Don't fail on refresh error - just continue with potentially stale data
		// The status command handles this the same way
		_, _ = fmt.Fprintf(out, "Warning: failed to refresh vehicle status: %v\n", err)
	}

	checkFunc := func() (bool, error) {
		var status any
		var err error

		if useEVStatus {
			status, err = client.GetEVVehicleStatus(ctx, internalVIN)
		} else {
			status, err = client.GetVehicleStatus(ctx, internalVIN)
		}

		if err != nil {
			return false, err
		}

		return conditionChecker(status)
	}

	return pollUntilCondition(ctx, out, checkFunc, timeout, pollInterval, actionName)
}

// waitForDoorsLocked polls the vehicle status until all doors are locked or timeout occurs.
func waitForDoorsLocked(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN api.InternalVIN,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	conditionChecker := func(status any) (bool, error) {
		vStatus := status.(*api.VehicleStatusResponse)
		doorStatus, err := vStatus.GetDoorsInfo()
		if err != nil {
			return false, err
		}

		return doorStatus.AllLocked, nil
	}

	return waitForCondition(ctx, out, client, internalVIN, false, conditionChecker, timeout, pollInterval, "door lock")
}

// waitForDoorsUnlocked polls the vehicle status until all doors are unlocked or timeout occurs.
func waitForDoorsUnlocked(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN api.InternalVIN,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	conditionChecker := func(status any) (bool, error) {
		vStatus := status.(*api.VehicleStatusResponse)
		doorStatus, err := vStatus.GetDoorsInfo()
		if err != nil {
			return false, err
		}
		// Unlocked means at least one door is unlocked (not all locked)
		return !doorStatus.AllLocked, nil
	}

	return waitForCondition(ctx, out, client, internalVIN, false, conditionChecker, timeout, pollInterval, "door unlock")
}

// waitForEngineRunning polls the vehicle status until the engine is running or timeout occurs.
func waitForEngineRunning(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN api.InternalVIN,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	conditionChecker := func(status any) (bool, error) {
		evStatus := status.(*api.EVVehicleStatusResponse)
		hvacInfo, err := evStatus.GetHvacInfo()
		if err != nil {
			return false, err
		}

		return hvacInfo.HVACOn, nil
	}

	return waitForCondition(ctx, out, client, internalVIN, true, conditionChecker, timeout, pollInterval, "engine start")
}

// waitForEngineStopped polls the vehicle status until the engine is stopped or timeout occurs.
func waitForEngineStopped(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN api.InternalVIN,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	conditionChecker := func(status any) (bool, error) {
		evStatus := status.(*api.EVVehicleStatusResponse)
		hvacInfo, err := evStatus.GetHvacInfo()
		if err != nil {
			return false, err
		}

		return !hvacInfo.HVACOn, nil
	}

	return waitForCondition(ctx, out, client, internalVIN, true, conditionChecker, timeout, pollInterval, "engine stop")
}

// waitForCharging polls the vehicle status until charging is active or timeout occurs.
func waitForCharging(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN api.InternalVIN,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	conditionChecker := func(status any) (bool, error) {
		evStatus := status.(*api.EVVehicleStatusResponse)
		batteryInfo, err := evStatus.GetBatteryInfo()
		if err != nil {
			return false, err
		}

		return batteryInfo.Charging, nil
	}

	return waitForCondition(ctx, out, client, internalVIN, true, conditionChecker, timeout, pollInterval, "charging start")
}

// waitForNotCharging polls the vehicle status until charging is inactive or timeout occurs.
func waitForNotCharging(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN api.InternalVIN,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	conditionChecker := func(status any) (bool, error) {
		evStatus := status.(*api.EVVehicleStatusResponse)
		batteryInfo, err := evStatus.GetBatteryInfo()
		if err != nil {
			return false, err
		}

		return !batteryInfo.Charging, nil
	}

	return waitForCondition(ctx, out, client, internalVIN, true, conditionChecker, timeout, pollInterval, "charging stop")
}

// ConfirmationInitialDelay is the time to wait before polling for command confirmation.
// Commands take time to propagate to the server before status is updated.
const ConfirmationInitialDelay = 20 * time.Second

// waitForHvacOn polls the vehicle status until HVAC is on or timeout occurs.
func waitForHvacOn(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN api.InternalVIN,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	conditionChecker := func(status any) (bool, error) {
		evStatus := status.(*api.EVVehicleStatusResponse)
		hvacInfo, err := evStatus.GetHvacInfo()
		if err != nil {
			return false, err
		}

		return hvacInfo.HVACOn, nil
	}

	return waitForCondition(ctx, out, client, internalVIN, true, conditionChecker, timeout, pollInterval, "HVAC on")
}

// waitForHvacOff polls the vehicle status until HVAC is off or timeout occurs.
func waitForHvacOff(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN api.InternalVIN,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	conditionChecker := func(status any) (bool, error) {
		evStatus := status.(*api.EVVehicleStatusResponse)
		hvacInfo, err := evStatus.GetHvacInfo()
		if err != nil {
			return false, err
		}

		return !hvacInfo.HVACOn, nil
	}

	return waitForCondition(ctx, out, client, internalVIN, true, conditionChecker, timeout, pollInterval, "HVAC off")
}

// waitForHvacSettings polls the vehicle status until HVAC settings match the requested values or timeout occurs.
func waitForHvacSettings(
	ctx context.Context,
	out io.Writer,
	client vehicleStatusGetter,
	internalVIN api.InternalVIN,
	targetTemp float64,
	frontDefroster bool,
	rearDefroster bool,
	timeout time.Duration,
	pollInterval time.Duration,
) confirmationResult {
	conditionChecker := func(status any) (bool, error) {
		evStatus := status.(*api.EVVehicleStatusResponse)
		hvacInfo, err := evStatus.GetHvacInfo()
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

	return waitForCondition(ctx, out, client, internalVIN, true, conditionChecker, timeout, pollInterval, "HVAC settings")
}

// DefaultPollInterval is the default time between status checks during confirmation polling.
const DefaultPollInterval = 5 * time.Second

// ConfirmableCommandConfig holds the configuration for a confirmable command.
type ConfirmableCommandConfig struct {
	// ActionFunc performs the API action (e.g., lock doors, start engine)
	ActionFunc func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error

	// WaitFunc waits for confirmation that the action completed
	// If nil, confirmation is skipped
	WaitFunc func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult

	// InitialDelay is the time to wait before starting confirmation polling.
	// Some commands (like HVAC) need time to propagate before status is updated.
	InitialDelay time.Duration

	// PollInterval is the time between status checks. If zero, DefaultPollInterval is used.
	PollInterval time.Duration

	// Messages
	SuccessMsg    string // Message to show on success (e.g., "Doors locked successfully")
	WaitingMsg    string // Message to show while waiting (e.g., "Lock command sent, waiting for confirmation...")
	ActionName    string // Name for error messages (e.g., "lock doors")
	ConfirmName   string // Name for confirmation error (e.g., "lock status")
	TimeoutSuffix string // Suffix for timeout message (e.g., "confirmation timeout")
}

// buildTimeoutMessage constructs the timeout message from waiting message and suffix.
func buildTimeoutMessage(waitingMsg, timeoutSuffix string) string {
	// Extract the command part from waiting message
	commandMsg := waitingMsg
	if idx := len(commandMsg) - len(", waiting for confirmation..."); idx > 0 {
		commandMsg = commandMsg[:idx]
	}

	return fmt.Sprintf("%s (%s)", commandMsg, timeoutSuffix)
}

// applyInitialDelay waits for the configured initial delay, respecting context cancellation.
func applyInitialDelay(ctx context.Context, delay time.Duration, actionName string) error {
	if delay <= 0 {
		return nil
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("failed to %s: %w", actionName, ctx.Err())
	case <-time.After(delay):
		return nil
	}
}

// executeConfirmableCommand executes a confirmable command with the given configuration.
func executeConfirmableCommand(
	ctx context.Context,
	out io.Writer,
	client *api.Client,
	internalVIN api.InternalVIN,
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

	timeout := time.Duration(confirmWait) * time.Second

	// Apply initial delay if configured
	if err := applyInitialDelay(ctx, config.InitialDelay, config.ActionName); err != nil {
		return err
	}
	timeout -= config.InitialDelay

	pollInterval := config.PollInterval
	if pollInterval == 0 {
		pollInterval = DefaultPollInterval
	}

	result := config.WaitFunc(ctx, out, client, internalVIN, timeout, pollInterval)

	if result.err != nil {
		return fmt.Errorf("failed to confirm %s: %w", config.ConfirmName, result.err)
	}

	if result.success {
		_, _ = fmt.Fprintln(out, config.SuccessMsg)
	} else {
		_, _ = fmt.Fprintln(out, buildTimeoutMessage(config.WaitingMsg, config.TimeoutSuffix))
	}

	return nil
}

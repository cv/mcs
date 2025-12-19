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

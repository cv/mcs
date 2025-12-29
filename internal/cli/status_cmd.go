package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewStatusCmd creates the status command.
func NewStatusCmd() *cobra.Command {
	var jsonOutput bool
	var refresh bool
	var refreshWait int

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show vehicle status",
		Long:  `Show comprehensive vehicle status including battery, fuel, location, tires, and doors.`,
		Example: `  # Show all vehicle status information
  mcs status

  # Example output:
  # CX-90 PHEV (2024)
  # VIN: JM3XXXXXXXXXX1234
  # Status as of 2024-03-15 14:30:45 (2 min ago)
  #
  # BATTERY: 85% [plugged in, not charging]
  # FUEL: 75% (45 km EV + 450 km fuel = 495 km total)
  # CLIMATE: Off, 18°C
  # DOORS: All locked
  # WINDOWS: All closed
  # TIRES: FL:35.0 FR:35.0 RL:33.0 RR:33.0 PSI (color-coded)
  # ODOMETER: 12,345.6 km

  # Show status in JSON format
  mcs status --json

  # Request fresh status from vehicle (PHEV/EV only, waits up to 90 seconds)
  mcs status --refresh`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, refresh, refreshWait)
		},
		SilenceUsage: true,
	}

	// Add flags
	statusCmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	statusCmd.Flags().BoolVarP(&refresh, "refresh", "r", false, "request fresh status from vehicle (PHEV/EV only)")
	statusCmd.Flags().IntVar(&refreshWait, "refresh-wait", 90, "max seconds to wait for vehicle response")

	return statusCmd
}

// runStatus executes the status command.
func runStatus(cmd *cobra.Command, jsonOutput bool, refresh bool, refreshWait int) error {
	return withVehicleClientEx(cmd.Context(), func(ctx context.Context, client *api.Client, vehicleInfo VehicleInfo) error {
		// Get initial EV status (needed for refresh comparison and final display)
		evStatus, err := client.GetEVVehicleStatus(ctx, string(vehicleInfo.InternalVIN))
		if err != nil {
			return fmt.Errorf("failed to get EV status: %w", err)
		}

		// If refresh requested, trigger status refresh and poll until timestamp changes
		if refresh {
			evStatus, err = refreshAndWaitForStatus(ctx, cmd, client, vehicleInfo.InternalVIN, evStatus, refreshWait)
			if err != nil {
				return err
			}
		}

		// Get vehicle status
		vehicleStatus, err := client.GetVehicleStatus(ctx, string(vehicleInfo.InternalVIN))
		if err != nil {
			return fmt.Errorf("failed to get vehicle status: %w", err)
		}

		// Display status
		output, err := displayAllStatus(vehicleStatus, evStatus, vehicleInfo, jsonOutput)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), output)

		return nil
	})
}

// refreshAndWaitForStatus triggers a status refresh and polls until the timestamp changes.
func refreshAndWaitForStatus(ctx context.Context, cmd *cobra.Command, client *api.Client, internalVIN api.InternalVIN, evStatus *api.EVVehicleStatusResponse, refreshWait int) (*api.EVVehicleStatusResponse, error) {
	initialTimestamp, err := evStatus.GetOccurrenceDate()
	if err != nil {
		return nil, fmt.Errorf("failed to get occurrence date: %w", err)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Current status from: %s\n", formatTimestamp(initialTimestamp))
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Requesting fresh status from vehicle...")

	if err := client.RefreshVehicleStatus(ctx, string(internalVIN)); err != nil {
		return nil, fmt.Errorf("failed to refresh vehicle status: %w", err)
	}

	// Poll every 30 seconds until timestamp changes or timeout
	pollInterval := 30 * time.Second
	maxWait := time.Duration(refreshWait) * time.Second

	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, maxWait)
	defer cancel()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	startTime := time.Now()
	for {
		select {
		case <-ticker.C:
			elapsed := time.Since(startTime)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Waiting for vehicle response... (%ds/%ds)\n", int(elapsed.Seconds()), refreshWait)

			// Fetch new EV status
			newEvStatus, err := client.GetEVVehicleStatus(timeoutCtx, string(internalVIN))
			if err != nil {
				continue // Keep trying on error
			}

			newTimestamp, err := newEvStatus.GetOccurrenceDate()
			if err != nil {
				continue // Keep trying on error
			}
			if newTimestamp != initialTimestamp {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Got fresh status from: %s\n", formatTimestamp(newTimestamp))

				return newEvStatus, nil
			}

		case <-timeoutCtx.Done():
			if timeoutCtx.Err() == context.DeadlineExceeded {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Warning: status did not update within timeout period")

				return evStatus, nil
			}

			return nil, timeoutCtx.Err()
		}
	}
}

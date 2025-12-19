package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// StatusType represents the type of status information to display
type StatusType string

const (
	StatusAll      StatusType = "all"
	StatusBattery  StatusType = "battery"
	StatusFuel     StatusType = "fuel"
	StatusLocation StatusType = "location"
	StatusTires    StatusType = "tires"
	StatusDoors    StatusType = "doors"
	StatusWindows  StatusType = "windows"
	StatusOdometer StatusType = "odometer"
	StatusHVAC     StatusType = "hvac"
)

// NewStatusCmd creates the status command
func NewStatusCmd() *cobra.Command {
	var jsonOutput bool
	var refresh bool
	var refreshWait int

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show vehicle status",
		Long:  `Show comprehensive vehicle status including battery, fuel, location, tires, and doors.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, StatusAll, refresh, refreshWait)
		},
		SilenceUsage: true,
	}

	// Add persistent JSON flag
	statusCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	statusCmd.PersistentFlags().BoolVar(&refresh, "refresh", false, "request fresh status from vehicle (PHEV/EV only)")
	statusCmd.PersistentFlags().IntVar(&refreshWait, "refresh-wait", 90, "max seconds to wait for vehicle response")

	// Add subcommands using configuration slice
	subcommands := []struct {
		use        string
		short      string
		statusType StatusType
	}{
		{"battery", "Show battery status", StatusBattery},
		{"fuel", "Show fuel status", StatusFuel},
		{"location", "Show vehicle location", StatusLocation},
		{"tires", "Show tire pressure", StatusTires},
		{"doors", "Show door lock status", StatusDoors},
	}

	for _, sc := range subcommands {
		// Capture loop variable for closure
		statusType := sc.statusType
		statusCmd.AddCommand(&cobra.Command{
			Use:   sc.use,
			Short: sc.short,
			RunE: func(cmd *cobra.Command, args []string) error {
				return runStatus(cmd, jsonOutput, statusType, refresh, refreshWait)
			},
			SilenceUsage: true,
		})
	}

	return statusCmd
}

// runStatus executes the status command
func runStatus(cmd *cobra.Command, jsonOutput bool, statusType StatusType, refresh bool, refreshWait int) error {
	return withVehicleClientEx(cmd.Context(), func(ctx context.Context, client *api.Client, vehicleInfo VehicleInfo) error {
		// Get initial EV status (needed for refresh comparison and final display)
		evStatus, err := client.GetEVVehicleStatus(ctx, vehicleInfo.InternalVIN)
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
		vehicleStatus, err := client.GetVehicleStatus(ctx, vehicleInfo.InternalVIN)
		if err != nil {
			return fmt.Errorf("failed to get vehicle status: %w", err)
		}

		// Display status based on type
		if err := displayStatusWithVehicle(cmd, statusType, vehicleStatus, evStatus, vehicleInfo, jsonOutput); err != nil {
			return err
		}
		return nil
	})
}

// refreshAndWaitForStatus triggers a status refresh and polls until the timestamp changes
func refreshAndWaitForStatus(ctx context.Context, cmd *cobra.Command, client *api.Client, internalVIN string, evStatus *api.EVVehicleStatusResponse, refreshWait int) (*api.EVVehicleStatusResponse, error) {
	initialTimestamp, err := evStatus.GetOccurrenceDate()
	if err != nil {
		return nil, fmt.Errorf("failed to get occurrence date: %w", err)
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Current status from: %s\n", formatTimestamp(initialTimestamp))
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Requesting fresh status from vehicle...")

	if err := client.RefreshVehicleStatus(ctx, internalVIN); err != nil {
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
			newEvStatus, err := client.GetEVVehicleStatus(timeoutCtx, internalVIN)
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

// displayStatusWithVehicle outputs the status based on type, including vehicle info for "all"
func displayStatusWithVehicle(cmd *cobra.Command, statusType StatusType, vehicleStatus *api.VehicleStatusResponse, evStatus *api.EVVehicleStatusResponse, vehicleInfo VehicleInfo, jsonOutput bool) error {
	var output string
	var err error

	switch statusType {
	case StatusBattery:
		output, err = displayBatteryStatus(evStatus, jsonOutput)
	case StatusFuel:
		output, err = displayFuelStatus(vehicleStatus, jsonOutput)
	case StatusLocation:
		output, err = displayLocationStatus(vehicleStatus, jsonOutput)
	case StatusTires:
		output, err = displayTiresStatus(vehicleStatus, jsonOutput)
	case StatusDoors:
		output, err = displayDoorsStatus(vehicleStatus, jsonOutput)
	case StatusAll:
		output, err = displayAllStatus(vehicleStatus, evStatus, vehicleInfo, jsonOutput)
	}

	if err != nil {
		return err
	}

	// Writing to stdout rarely fails, so we ignore the error here
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), output)
	return nil
}

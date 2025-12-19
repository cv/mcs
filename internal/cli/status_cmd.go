package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
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
			return runStatus(cmd, jsonOutput, "all", refresh, refreshWait)
		},
		SilenceUsage: true,
	}

	// Add persistent JSON flag
	statusCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	statusCmd.PersistentFlags().BoolVar(&refresh, "refresh", false, "request fresh status from vehicle (PHEV/EV only)")
	statusCmd.PersistentFlags().IntVar(&refreshWait, "refresh-wait", 90, "max seconds to wait for vehicle response")

	// Add subcommands
	statusCmd.AddCommand(&cobra.Command{
		Use:   "battery",
		Short: "Show battery status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, "battery", refresh, refreshWait)
		},
		SilenceUsage: true,
	})

	statusCmd.AddCommand(&cobra.Command{
		Use:   "fuel",
		Short: "Show fuel status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, "fuel", refresh, refreshWait)
		},
		SilenceUsage: true,
	})

	statusCmd.AddCommand(&cobra.Command{
		Use:   "location",
		Short: "Show vehicle location",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, "location", refresh, refreshWait)
		},
		SilenceUsage: true,
	})

	statusCmd.AddCommand(&cobra.Command{
		Use:   "tires",
		Short: "Show tire pressure",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, "tires", refresh, refreshWait)
		},
		SilenceUsage: true,
	})

	statusCmd.AddCommand(&cobra.Command{
		Use:   "doors",
		Short: "Show door lock status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, "doors", refresh, refreshWait)
		},
		SilenceUsage: true,
	})

	return statusCmd
}

// runStatus executes the status command
func runStatus(cmd *cobra.Command, jsonOutput bool, statusType string, refresh bool, refreshWait int) error {
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
	elapsed := time.Duration(0)

	for elapsed < maxWait {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Waiting for vehicle response... (%ds/%ds)\n", int(elapsed.Seconds()), refreshWait)
		select {
		case <-time.After(pollInterval):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		elapsed += pollInterval

		// Fetch new EV status
		newEvStatus, err := client.GetEVVehicleStatus(ctx, internalVIN)
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
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Warning: status did not update within timeout period")
	return evStatus, nil
}

// displayStatusWithVehicle outputs the status based on type, including vehicle info for "all"
func displayStatusWithVehicle(cmd *cobra.Command, statusType string, vehicleStatus *api.VehicleStatusResponse, evStatus *api.EVVehicleStatusResponse, vehicleInfo VehicleInfo, jsonOutput bool) error {
	var output string
	var err error

	switch statusType {
	case "battery":
		output, err = displayBatteryStatus(evStatus, jsonOutput)
	case "fuel":
		output, err = displayFuelStatus(vehicleStatus, jsonOutput)
	case "location":
		output, err = displayLocationStatus(vehicleStatus, jsonOutput)
	case "tires":
		output, err = displayTiresStatus(vehicleStatus, jsonOutput)
	case "doors":
		output, err = displayDoorsStatus(vehicleStatus, jsonOutput)
	case "all":
		output, err = displayAllStatus(vehicleStatus, evStatus, vehicleInfo, jsonOutput)
	}

	if err != nil {
		return err
	}

	// Writing to stdout rarely fails, so we ignore the error here
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), output)
	return nil
}

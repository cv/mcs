package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// StatusType represents the type of status information to display.
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
  mcs status --refresh

  # Show specific status category
  mcs status battery
  mcs status fuel
  mcs status location`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, StatusAll, refresh, refreshWait)
		},
		SilenceUsage: true,
	}

	// Add persistent JSON flag
	statusCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	statusCmd.PersistentFlags().BoolVarP(&refresh, "refresh", "r", false, "request fresh status from vehicle (PHEV/EV only)")
	statusCmd.PersistentFlags().IntVar(&refreshWait, "refresh-wait", 90, "max seconds to wait for vehicle response")

	// Add battery subcommand
	statusCmd.AddCommand(&cobra.Command{
		Use:   "battery",
		Short: "Show battery status",
		Example: `  # Show battery status
  mcs status battery

  # Example output:
  # BATTERY: 85% (42.5 km range) [plugged in, not charging]

  # Show battery charging with time estimates
  # BATTERY: 65% (32.5 km range) [charging, ~2h 30m quick / ~4h AC]

  # Show battery status in JSON format
  mcs status battery --json

  # Example JSON output:
  # {
  #   "battery_level": 85,
  #   "range_km": 42.5,
  #   "charge_time_ac_min": 0,
  #   "charge_time_qbc_min": 0,
  #   "plugged_in": true,
  #   "charging": false,
  #   "heater_on": false,
  #   "heater_auto": true
  # }`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, StatusBattery, refresh, refreshWait)
		},
		SilenceUsage: true,
	})

	// Add fuel subcommand
	statusCmd.AddCommand(&cobra.Command{
		Use:   "fuel",
		Short: "Show fuel status",
		Example: `  # Show fuel status
  mcs status fuel

  # Example output:
  # FUEL: 75% (450.5 km range)

  # Show fuel status in JSON format
  mcs status fuel --json

  # Example JSON output:
  # {
  #   "fuel_level": 75,
  #   "range_km": 450.5
  # }`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, StatusFuel, refresh, refreshWait)
		},
		SilenceUsage: true,
	})

	// Add location subcommand
	statusCmd.AddCommand(&cobra.Command{
		Use:   "location",
		Short: "Show vehicle location",
		Example: `  # Show vehicle location
  mcs status location

  # Example output:
  # LOCATION: 37.774929, -122.419418
  #   https://maps.google.com/?q=37.774929,-122.419418

  # Show location in JSON format
  mcs status location --json

  # Example JSON output:
  # {
  #   "latitude": 37.774929,
  #   "longitude": -122.419418,
  #   "timestamp": "20240315143045"
  # }`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, StatusLocation, refresh, refreshWait)
		},
		SilenceUsage: true,
	})

	// Add tires subcommand
	statusCmd.AddCommand(&cobra.Command{
		Use:   "tires",
		Short: "Show tire pressure",
		Example: `  # Show tire pressure (color-coded based on deviation from 36 PSI target)
  mcs status tires

  # Example output:
  # TIRES: FL:35.0 FR:35.0 RL:33.0 RR:33.0 PSI
  #
  # Color coding: Green (±3 PSI), Yellow (4-6 PSI off), Red (>6 PSI off)

  # Show tire pressure in JSON format
  mcs status tires --json

  # Example JSON output:
  # {
  #   "front_left_psi": 35.0,
  #   "front_right_psi": 35.0,
  #   "rear_left_psi": 33.0,
  #   "rear_right_psi": 33.0
  # }`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, StatusTires, refresh, refreshWait)
		},
		SilenceUsage: true,
	})

	// Add doors subcommand
	statusCmd.AddCommand(&cobra.Command{
		Use:   "doors",
		Short: "Show door lock status",
		Example: `  # Show door lock status
  mcs status doors

  # Example output when all locked:
  # DOORS: All locked

  # Example output with issues:
  # DOORS: Driver unlocked, Trunk open

  # Show door status in JSON format
  mcs status doors --json

  # Example JSON output:
  # {
  #   "driver_open": false,
  #   "driver_locked": true,
  #   "passenger_open": false,
  #   "passenger_locked": true,
  #   "rear_left_open": false,
  #   "rear_left_locked": true,
  #   "rear_right_open": false,
  #   "rear_right_locked": true,
  #   "trunk_open": false,
  #   "hood_open": false,
  #   "fuel_lid_open": false,
  #   "all_locked": true
  # }`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, StatusDoors, refresh, refreshWait)
		},
		SilenceUsage: true,
	})

	return statusCmd
}

// runStatus executes the status command.
func runStatus(cmd *cobra.Command, jsonOutput bool, statusType StatusType, refresh bool, refreshWait int) error {
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

		// Display status based on type
		if err := displayStatusWithVehicle(cmd, statusType, vehicleStatus, evStatus, vehicleInfo, jsonOutput); err != nil {
			return err
		}

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

// displayStatusWithVehicle outputs the status based on type, including vehicle info for "all".
func displayStatusWithVehicle(cmd *cobra.Command, statusType StatusType, vehicleStatus *api.VehicleStatusResponse, evStatus *api.EVVehicleStatusResponse, vehicleInfo VehicleInfo, jsonOutput bool) error {
	var output string
	var err error

	switch statusType {
	case StatusBattery:
		batteryInfo, _ := evStatus.GetBatteryInfo()
		output, err = formatBatteryStatus(batteryInfo, jsonOutput)
	case StatusFuel:
		fuelInfo, _ := vehicleStatus.GetFuelInfo()
		output, err = formatFuelStatus(fuelInfo, jsonOutput)
	case StatusLocation:
		locationInfo, _ := vehicleStatus.GetLocationInfo()
		output, err = formatLocationStatus(locationInfo, jsonOutput)
	case StatusTires:
		tireInfo, _ := vehicleStatus.GetTiresInfo()
		output, err = formatTiresStatus(tireInfo, jsonOutput)
	case StatusDoors:
		doorStatus, _ := vehicleStatus.GetDoorsInfo()
		output, err = formatDoorsStatus(doorStatus, jsonOutput)
	case StatusWindows:
		windowsInfo, _ := vehicleStatus.GetWindowsInfo()
		output, err = formatWindowsStatus(windowsInfo, jsonOutput)
	case StatusOdometer:
		odometerInfo, _ := vehicleStatus.GetOdometerInfo()
		output, err = formatOdometerStatus(odometerInfo, jsonOutput)
	case StatusHVAC:
		hvacInfo, _ := evStatus.GetHvacInfo()
		output, err = formatHvacStatus(hvacInfo, jsonOutput)
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

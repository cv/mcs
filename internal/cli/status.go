package cli

import (
	"encoding/json"
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
	// Create API client (with cached credentials if available)
	client, err := createAPIClient()
	if err != nil {
		return err
	}
	defer saveClientCache(client)

	// Get vehicle base info to retrieve internal VIN
	vecBaseInfos, err := client.GetVecBaseInfos()
	if err != nil {
		return fmt.Errorf("failed to get vehicle info: %w", err)
	}

	internalVIN, err := vecBaseInfos.GetInternalVIN()
	if err != nil {
		return err
	}

	// Get initial EV status (needed for refresh comparison and final display)
	evStatus, err := client.GetEVVehicleStatus(internalVIN)
	if err != nil {
		return fmt.Errorf("failed to get EV status: %w", err)
	}

	// If refresh requested, trigger status refresh and poll until timestamp changes
	if refresh {
		initialTimestamp := evStatus.GetOccurrenceDate()
		fmt.Fprintf(cmd.OutOrStdout(), "Current status from: %s\n", formatTimestamp(initialTimestamp))
		fmt.Fprintln(cmd.OutOrStdout(), "Requesting fresh status from vehicle...")

		if err := client.RefreshVehicleStatus(internalVIN); err != nil {
			return fmt.Errorf("failed to refresh vehicle status: %w", err)
		}

		// Poll every 30 seconds until timestamp changes or timeout
		pollInterval := 30 * time.Second
		maxWait := time.Duration(refreshWait) * time.Second
		elapsed := time.Duration(0)

		for elapsed < maxWait {
			fmt.Fprintf(cmd.OutOrStdout(), "Waiting for vehicle response... (%ds/%ds)\n", int(elapsed.Seconds()), refreshWait)
			time.Sleep(pollInterval)
			elapsed += pollInterval

			// Fetch new EV status
			newEvStatus, err := client.GetEVVehicleStatus(internalVIN)
			if err != nil {
				continue // Keep trying on error
			}

			newTimestamp := newEvStatus.GetOccurrenceDate()
			if newTimestamp != initialTimestamp {
				fmt.Fprintf(cmd.OutOrStdout(), "Got fresh status from: %s\n", formatTimestamp(newTimestamp))
				evStatus = newEvStatus
				break
			}
		}

		if evStatus.GetOccurrenceDate() == initialTimestamp {
			fmt.Fprintln(cmd.OutOrStdout(), "Warning: status did not update within timeout period")
		}
	}

	// Get vehicle status
	vehicleStatus, err := client.GetVehicleStatus(internalVIN)
	if err != nil {
		return fmt.Errorf("failed to get vehicle status: %w", err)
	}

	// Display status based on type
	switch statusType {
	case "battery":
		output := displayBatteryStatus(evStatus, jsonOutput)
		fmt.Fprintln(cmd.OutOrStdout(), output)
	case "fuel":
		output := displayFuelStatus(vehicleStatus, jsonOutput)
		fmt.Fprintln(cmd.OutOrStdout(), output)
	case "location":
		output := displayLocationStatus(vehicleStatus, jsonOutput)
		fmt.Fprintln(cmd.OutOrStdout(), output)
	case "tires":
		output := displayTiresStatus(vehicleStatus, jsonOutput)
		fmt.Fprintln(cmd.OutOrStdout(), output)
	case "doors":
		output := displayDoorsStatus(vehicleStatus, jsonOutput)
		fmt.Fprintln(cmd.OutOrStdout(), output)
	case "all":
		output := displayAllStatus(vehicleStatus, evStatus, jsonOutput)
		fmt.Fprint(cmd.OutOrStdout(), output)
	}

	return nil
}

// displayAllStatus displays all status information
func displayAllStatus(vehicleStatus *api.VehicleStatusResponse, evStatus *api.EVVehicleStatusResponse, jsonOutput bool) string {
	if jsonOutput {
		data := map[string]interface{}{
			"battery":  extractBatteryData(evStatus),
			"fuel":     extractFuelData(vehicleStatus),
			"location": extractLocationData(vehicleStatus),
			"tires":    extractTiresData(vehicleStatus),
			"doors":    extractDoorsData(vehicleStatus),
			"climate":  extractHvacData(evStatus),
		}
		jsonBytes, _ := json.MarshalIndent(data, "", "  ")
		return string(jsonBytes)
	}

	// Get timestamp from EV status
	timestamp := formatTimestamp(evStatus.GetOccurrenceDate())

	// Extract HVAC info
	hvacOn, frontDefroster, rearDefroster, interiorTempC := evStatus.GetHvacInfo()

	output := fmt.Sprintf("\nVehicle Status (Last Updated: %s)\n\n", timestamp)
	output += displayBatteryStatus(evStatus, false) + "\n"
	output += displayFuelStatus(vehicleStatus, false) + "\n"
	output += formatHvacStatus(hvacOn, frontDefroster, rearDefroster, interiorTempC, false) + "\n"
	output += displayDoorsStatus(vehicleStatus, false) + "\n"
	output += displayTiresStatus(vehicleStatus, false) + "\n"

	return output
}

// displayBatteryStatus displays battery status
func displayBatteryStatus(evStatus *api.EVVehicleStatusResponse, jsonOutput bool) string {
	batteryLevel, range_, pluggedIn, charging, _ := evStatus.GetBatteryInfo()
	return formatBatteryStatus(batteryLevel, range_, pluggedIn, charging, jsonOutput)
}

// displayFuelStatus displays fuel status
func displayFuelStatus(vehicleStatus *api.VehicleStatusResponse, jsonOutput bool) string {
	fuelLevel, range_, _ := vehicleStatus.GetFuelInfo()
	return formatFuelStatus(fuelLevel, range_, jsonOutput)
}

// displayLocationStatus displays location status
func displayLocationStatus(vehicleStatus *api.VehicleStatusResponse, jsonOutput bool) string {
	lat, lon, timestamp, _ := vehicleStatus.GetLocationInfo()
	return formatLocationStatus(lat, lon, timestamp, jsonOutput)
}

// displayTiresStatus displays tire pressure status
func displayTiresStatus(vehicleStatus *api.VehicleStatusResponse, jsonOutput bool) string {
	fl, fr, rl, rr, _ := vehicleStatus.GetTiresInfo()
	return formatTiresStatus(fl, fr, rl, rr, jsonOutput)
}

// displayDoorsStatus displays door lock status
func displayDoorsStatus(vehicleStatus *api.VehicleStatusResponse, jsonOutput bool) string {
	allLocked, _ := vehicleStatus.GetDoorsInfo()
	return formatDoorsStatus(allLocked, jsonOutput)
}


// extractBatteryData extracts battery data for JSON output
func extractBatteryData(evStatus *api.EVVehicleStatusResponse) map[string]interface{} {
	batteryLevel, range_, pluggedIn, charging, _ := evStatus.GetBatteryInfo()
	return map[string]interface{}{
		"battery_level": batteryLevel,
		"range_km":      range_,
		"plugged_in":    pluggedIn,
		"charging":      charging,
	}
}

// extractFuelData extracts fuel data for JSON output
func extractFuelData(vehicleStatus *api.VehicleStatusResponse) map[string]interface{} {
	fuelLevel, range_, _ := vehicleStatus.GetFuelInfo()
	return map[string]interface{}{
		"fuel_level": fuelLevel,
		"range_km":   range_,
	}
}

// extractLocationData extracts location data for JSON output
func extractLocationData(vehicleStatus *api.VehicleStatusResponse) map[string]interface{} {
	lat, lon, timestamp, _ := vehicleStatus.GetLocationInfo()
	mapsURL := fmt.Sprintf("https://maps.google.com/?q=%f,%f", lat, lon)
	return map[string]interface{}{
		"latitude":  lat,
		"longitude": lon,
		"timestamp": timestamp,
		"maps_url":  mapsURL,
	}
}

// extractTiresData extracts tire data for JSON output
func extractTiresData(vehicleStatus *api.VehicleStatusResponse) map[string]interface{} {
	fl, fr, rl, rr, _ := vehicleStatus.GetTiresInfo()
	return map[string]interface{}{
		"front_left_psi":  fl,
		"front_right_psi": fr,
		"rear_left_psi":   rl,
		"rear_right_psi":  rr,
	}
}

// extractDoorsData extracts door data for JSON output
func extractDoorsData(vehicleStatus *api.VehicleStatusResponse) map[string]interface{} {
	allLocked, _ := vehicleStatus.GetDoorsInfo()
	return map[string]interface{}{
		"all_locked": allLocked,
	}
}

// extractHvacData extracts HVAC data for JSON output
func extractHvacData(evStatus *api.EVVehicleStatusResponse) map[string]interface{} {
	hvacOn, frontDefroster, rearDefroster, interiorTempC := evStatus.GetHvacInfo()
	return map[string]interface{}{
		"hvac_on":                hvacOn,
		"front_defroster":        frontDefroster,
		"rear_defroster":         rearDefroster,
		"interior_temperature_c": interiorTempC,
	}
}

// formatBatteryStatus formats battery status for display
func formatBatteryStatus(batteryLevel, range_ float64, pluggedIn, charging bool, jsonOutput bool) string {
	if jsonOutput {
		data := map[string]interface{}{
			"battery_level": batteryLevel,
			"range_km":      range_,
			"plugged_in":    pluggedIn,
			"charging":      charging,
		}
		jsonBytes, _ := json.MarshalIndent(data, "", "  ")
		return string(jsonBytes)
	}

	status := fmt.Sprintf("BATTERY: %.0f%% (%.1f km range)", batteryLevel, range_)
	if pluggedIn {
		if charging {
			status += " [plugged in, charging]"
		} else {
			status += " [plugged in, not charging]"
		}
	}

	return status
}

// formatFuelStatus formats fuel status for display
func formatFuelStatus(fuelLevel, range_ float64, jsonOutput bool) string {
	if jsonOutput {
		data := map[string]interface{}{
			"fuel_level": fuelLevel,
			"range_km":   range_,
		}
		jsonBytes, _ := json.MarshalIndent(data, "", "  ")
		return string(jsonBytes)
	}

	return fmt.Sprintf("FUEL: %.0f%% (%.1f km range)", fuelLevel, range_)
}

// formatLocationStatus formats location status for display
func formatLocationStatus(lat, lon float64, timestamp string, jsonOutput bool) string {
	if jsonOutput {
		mapsURL := fmt.Sprintf("https://maps.google.com/?q=%f,%f", lat, lon)
		data := map[string]interface{}{
			"latitude":  lat,
			"longitude": lon,
			"timestamp": timestamp,
			"maps_url":  mapsURL,
		}
		jsonBytes, _ := json.MarshalIndent(data, "", "  ")
		return string(jsonBytes)
	}

	mapsURL := fmt.Sprintf("https://maps.google.com/?q=%f,%f", lat, lon)
	return fmt.Sprintf("LOCATION: %.6f, %.6f\n  %s", lat, lon, mapsURL)
}

// formatTiresStatus formats tire status for display
func formatTiresStatus(fl, fr, rl, rr float64, jsonOutput bool) string {
	if jsonOutput {
		data := map[string]interface{}{
			"front_left_psi":  fl,
			"front_right_psi": fr,
			"rear_left_psi":   rl,
			"rear_right_psi":  rr,
		}
		jsonBytes, _ := json.MarshalIndent(data, "", "  ")
		return string(jsonBytes)
	}

	return fmt.Sprintf("TIRES: FL:%.1f FR:%.1f RL:%.1f RR:%.1f PSI", fl, fr, rl, rr)
}

// formatDoorsStatus formats door status for display
func formatDoorsStatus(allLocked bool, jsonOutput bool) string {
	if jsonOutput {
		data := map[string]interface{}{
			"all_locked": allLocked,
		}
		jsonBytes, _ := json.MarshalIndent(data, "", "  ")
		return string(jsonBytes)
	}

	if allLocked {
		return "DOORS: All locked"
	}
	return "DOORS: Not all locked"
}

// formatHvacStatus formats HVAC status for display
func formatHvacStatus(hvacOn, frontDefroster, rearDefroster bool, interiorTempC float64, jsonOutput bool) string {
	if jsonOutput {
		data := map[string]interface{}{
			"hvac_on":                hvacOn,
			"front_defroster":        frontDefroster,
			"rear_defroster":         rearDefroster,
			"interior_temperature_c": interiorTempC,
		}
		jsonBytes, _ := json.MarshalIndent(data, "", "  ")
		return string(jsonBytes)
	}

	var status string
	if hvacOn {
		status = fmt.Sprintf("CLIMATE: On, %.0f°C", interiorTempC)
	} else {
		status = fmt.Sprintf("CLIMATE: Off, %.0f°C", interiorTempC)
	}

	// Build defroster status
	var defrosters []string
	if frontDefroster {
		defrosters = append(defrosters, "front")
	}
	if rearDefroster {
		defrosters = append(defrosters, "rear")
	}

	if len(defrosters) == 2 {
		status += " (front and rear defrosters on)"
	} else if len(defrosters) == 1 {
		status += fmt.Sprintf(" (%s defroster on)", defrosters[0])
	}

	return status
}

// formatTimestamp converts timestamp from API format to readable format
func formatTimestamp(timestamp string) string {
	// API returns timestamp in format: YYYYMMDDHHmmss
	// Convert to: YYYY-MM-DD HH:mm:ss
	if len(timestamp) != 14 {
		return timestamp
	}

	t, err := time.Parse("20060102150405", timestamp)
	if err != nil {
		return timestamp
	}

	return t.Format("2006-01-02 15:04:05")
}

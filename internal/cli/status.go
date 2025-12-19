package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
		displayStatusWithVehicle(cmd, statusType, vehicleStatus, evStatus, vehicleInfo, jsonOutput)
		return nil
	})
}

// refreshAndWaitForStatus triggers a status refresh and polls until the timestamp changes
func refreshAndWaitForStatus(ctx context.Context, cmd *cobra.Command, client *api.Client, internalVIN string, evStatus *api.EVVehicleStatusResponse, refreshWait int) (*api.EVVehicleStatusResponse, error) {
	initialTimestamp := evStatus.GetOccurrenceDate()
	fmt.Fprintf(cmd.OutOrStdout(), "Current status from: %s\n", formatTimestamp(initialTimestamp))
	fmt.Fprintln(cmd.OutOrStdout(), "Requesting fresh status from vehicle...")

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

		fmt.Fprintf(cmd.OutOrStdout(), "Waiting for vehicle response... (%ds/%ds)\n", int(elapsed.Seconds()), refreshWait)
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

		newTimestamp := newEvStatus.GetOccurrenceDate()
		if newTimestamp != initialTimestamp {
			fmt.Fprintf(cmd.OutOrStdout(), "Got fresh status from: %s\n", formatTimestamp(newTimestamp))
			return newEvStatus, nil
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Warning: status did not update within timeout period")
	return evStatus, nil
}

// displayStatusWithVehicle outputs the status based on type, including vehicle info for "all"
func displayStatusWithVehicle(cmd *cobra.Command, statusType string, vehicleStatus *api.VehicleStatusResponse, evStatus *api.EVVehicleStatusResponse, vehicleInfo VehicleInfo, jsonOutput bool) {
	switch statusType {
	case "battery":
		fmt.Fprintln(cmd.OutOrStdout(), displayBatteryStatus(evStatus, jsonOutput))
	case "fuel":
		fmt.Fprintln(cmd.OutOrStdout(), displayFuelStatus(vehicleStatus, jsonOutput))
	case "location":
		fmt.Fprintln(cmd.OutOrStdout(), displayLocationStatus(vehicleStatus, jsonOutput))
	case "tires":
		fmt.Fprintln(cmd.OutOrStdout(), displayTiresStatus(vehicleStatus, jsonOutput))
	case "doors":
		fmt.Fprintln(cmd.OutOrStdout(), displayDoorsStatus(vehicleStatus, jsonOutput))
	case "all":
		fmt.Fprint(cmd.OutOrStdout(), displayAllStatus(vehicleStatus, evStatus, vehicleInfo, jsonOutput))
	}
}

// displayAllStatus displays all status information
func displayAllStatus(vehicleStatus *api.VehicleStatusResponse, evStatus *api.EVVehicleStatusResponse, vehicleInfo VehicleInfo, jsonOutput bool) string {
	if jsonOutput {
		hazardsOn, _ := vehicleStatus.GetHazardInfo()
		data := map[string]interface{}{
			"vehicle":  extractVehicleInfoData(vehicleInfo),
			"battery":  extractBatteryData(evStatus),
			"fuel":     extractFuelData(vehicleStatus),
			"location": extractLocationData(vehicleStatus),
			"tires":    extractTiresData(vehicleStatus),
			"doors":    extractDoorsData(vehicleStatus),
			"windows":  extractWindowsData(vehicleStatus),
			"hazards":  hazardsOn,
			"climate":  extractHvacData(evStatus),
			"odometer": extractOdometerData(vehicleStatus),
		}
		jsonBytes, _ := json.MarshalIndent(data, "", "  ")
		return string(jsonBytes)
	}

	// Get timestamp from EV status
	timestamp := formatTimestamp(evStatus.GetOccurrenceDate())

	// Extract HVAC info
	hvacOn, frontDefroster, rearDefroster, interiorTempC, targetTempC := evStatus.GetHvacInfo()

	// Extract odometer
	odometer, _ := vehicleStatus.GetOdometerInfo()

	// Extract windows info
	driver, passenger, rearLeft, rearRight, _ := vehicleStatus.GetWindowsInfo()

	// Extract hazard info
	hazardsOn, _ := vehicleStatus.GetHazardInfo()

	// Extract battery and fuel info
	batteryLevel, evRange, chargeTimeACMin, chargeTimeQBCMin, pluggedIn, charging, heaterOn, heaterAuto, _ := evStatus.GetBatteryInfo()
	fuelLevel, fuelRange, _ := vehicleStatus.GetFuelInfo()

	// Build vehicle header
	output := "\n" + formatVehicleHeader(vehicleInfo) + "\n"
	output += fmt.Sprintf("Status as of %s\n\n", timestamp)
	output += formatBatteryStatusCompact(batteryLevel, chargeTimeACMin, chargeTimeQBCMin, pluggedIn, charging, heaterOn, heaterAuto) + "\n"
	output += formatFuelStatusWithRange(fuelLevel, fuelRange, evRange) + "\n"
	output += formatHvacStatus(hvacOn, frontDefroster, rearDefroster, interiorTempC, targetTempC, false) + "\n"
	output += displayDoorsStatus(vehicleStatus, false) + "\n"
	output += formatWindowsStatus(driver, passenger, rearLeft, rearRight, false) + "\n"

	// Only show hazards if they're on
	if hazardsOn {
		output += "HAZARDS: On\n"
	}

	output += displayTiresStatus(vehicleStatus, false) + "\n"
	output += formatOdometerStatus(odometer, false) + "\n"

	return output
}

// extractVehicleInfoData extracts vehicle info for JSON output
func extractVehicleInfoData(vehicleInfo VehicleInfo) map[string]interface{} {
	return map[string]interface{}{
		"vin":        vehicleInfo.VIN,
		"nickname":   vehicleInfo.Nickname,
		"model_name": vehicleInfo.ModelName,
		"model_year": vehicleInfo.ModelYear,
	}
}

// formatVehicleHeader formats vehicle identification for display
func formatVehicleHeader(vehicleInfo VehicleInfo) string {
	var header string

	// Build model line: "CX-90 PHEV (2024)" or just model name
	if vehicleInfo.ModelName != "" {
		header = vehicleInfo.ModelName
		if vehicleInfo.ModelYear != "" {
			header += fmt.Sprintf(" (%s)", vehicleInfo.ModelYear)
		}
		// Add nickname in quotes if present
		if vehicleInfo.Nickname != "" {
			header += fmt.Sprintf(" \"%s\"", vehicleInfo.Nickname)
		}
		header += "\n"
	} else if vehicleInfo.Nickname != "" {
		// No model but has nickname
		header = fmt.Sprintf("\"%s\"\n", vehicleInfo.Nickname)
	}

	// Add VIN line if available
	if vehicleInfo.VIN != "" {
		header += fmt.Sprintf("VIN: %s\n", vehicleInfo.VIN)
	}

	return header
}

// displayBatteryStatus displays battery status
func displayBatteryStatus(evStatus *api.EVVehicleStatusResponse, jsonOutput bool) string {
	batteryLevel, range_, chargeTimeACMin, chargeTimeQBCMin, pluggedIn, charging, heaterOn, heaterAuto, _ := evStatus.GetBatteryInfo()
	return formatBatteryStatus(batteryLevel, range_, chargeTimeACMin, chargeTimeQBCMin, pluggedIn, charging, heaterOn, heaterAuto, jsonOutput)
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
	doorStatus, _ := vehicleStatus.GetDoorsInfo()
	return formatDoorsStatus(doorStatus, jsonOutput)
}


// extractBatteryData extracts battery data for JSON output
func extractBatteryData(evStatus *api.EVVehicleStatusResponse) map[string]interface{} {
	batteryLevel, range_, chargeTimeACMin, chargeTimeQBCMin, pluggedIn, charging, heaterOn, heaterAuto, _ := evStatus.GetBatteryInfo()
	data := map[string]interface{}{
		"battery_level":   batteryLevel,
		"range_km":        range_,
		"plugged_in":      pluggedIn,
		"charging":        charging,
		"heater_on":       heaterOn,
		"heater_auto":     heaterAuto,
	}
	if charging {
		data["charge_time_ac_minutes"] = chargeTimeACMin
		data["charge_time_qbc_minutes"] = chargeTimeQBCMin
	}
	return data
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
	doorStatus, _ := vehicleStatus.GetDoorsInfo()
	return map[string]interface{}{
		"all_locked":        doorStatus.AllLocked,
		"driver_open":       doorStatus.DriverOpen,
		"passenger_open":    doorStatus.PassengerOpen,
		"rear_left_open":    doorStatus.RearLeftOpen,
		"rear_right_open":   doorStatus.RearRightOpen,
		"trunk_open":        doorStatus.TrunkOpen,
		"hood_open":         doorStatus.HoodOpen,
		"fuel_lid_open":     doorStatus.FuelLidOpen,
		"driver_locked":     doorStatus.DriverLocked,
		"passenger_locked":  doorStatus.PassengerLocked,
		"rear_left_locked":  doorStatus.RearLeftLocked,
		"rear_right_locked": doorStatus.RearRightLocked,
	}
}

// extractOdometerData extracts odometer data for JSON output
func extractOdometerData(vehicleStatus *api.VehicleStatusResponse) map[string]interface{} {
	odometer, _ := vehicleStatus.GetOdometerInfo()
	return map[string]interface{}{
		"odometer_km": odometer,
	}
}

// extractHvacData extracts HVAC data for JSON output
func extractHvacData(evStatus *api.EVVehicleStatusResponse) map[string]interface{} {
	hvacOn, frontDefroster, rearDefroster, interiorTempC, targetTempC := evStatus.GetHvacInfo()
	return map[string]interface{}{
		"hvac_on":                hvacOn,
		"front_defroster":        frontDefroster,
		"rear_defroster":         rearDefroster,
		"interior_temperature_c": interiorTempC,
		"target_temperature_c":   targetTempC,
	}
}

// extractWindowsData extracts window data for JSON output
func extractWindowsData(vehicleStatus *api.VehicleStatusResponse) map[string]interface{} {
	driver, passenger, rearLeft, rearRight, _ := vehicleStatus.GetWindowsInfo()
	return map[string]interface{}{
		"driver_position":    driver,
		"passenger_position": passenger,
		"rear_left_position": rearLeft,
		"rear_right_position": rearRight,
	}
}

// toJSON converts a map to formatted JSON string
func toJSON(data map[string]interface{}) string {
	jsonBytes, _ := json.MarshalIndent(data, "", "  ")
	return string(jsonBytes)
}

// formatBatteryStatus formats battery status for display
func formatBatteryStatus(batteryLevel, range_, chargeTimeACMin, chargeTimeQBCMin float64, pluggedIn, charging, heaterOn, heaterAuto bool, jsonOutput bool) string {
	if jsonOutput {
		data := map[string]interface{}{
			"battery_level": batteryLevel,
			"range_km":      range_,
			"plugged_in":    pluggedIn,
			"charging":      charging,
			"heater_on":     heaterOn,
			"heater_auto":   heaterAuto,
		}
		if charging {
			data["charge_time_ac_minutes"] = chargeTimeACMin
			data["charge_time_qbc_minutes"] = chargeTimeQBCMin
		}
		return toJSON(data)
	}

	status := fmt.Sprintf("BATTERY: %.0f%% (%.1f km range)", batteryLevel, range_)

	// Build status flags
	var flags []string

	if pluggedIn {
		if charging {
			// Show charging time estimates
			timeStr := formatChargeTime(chargeTimeACMin, chargeTimeQBCMin)
			if timeStr != "" {
				flags = append(flags, fmt.Sprintf("charging, %s", timeStr))
			} else {
				flags = append(flags, "charging")
			}
		} else {
			flags = append(flags, "plugged in, not charging")
		}
	}

	// Add heater status
	if heaterOn {
		if heaterAuto {
			flags = append(flags, "battery heater on, auto enabled")
		} else {
			flags = append(flags, "battery heater on")
		}
	} else if heaterAuto {
		flags = append(flags, "battery heater auto enabled")
	}

	if len(flags) > 0 {
		status += fmt.Sprintf(" [%s]", strings.Join(flags, ", "))
	}

	return status
}

// formatFuelStatus formats fuel status for display
func formatFuelStatus(fuelLevel, range_ float64, jsonOutput bool) string {
	if jsonOutput {
		return toJSON(map[string]interface{}{
			"fuel_level": fuelLevel,
			"range_km":   range_,
		})
	}

	return fmt.Sprintf("FUEL: %.0f%% (%.1f km range)", fuelLevel, range_)
}

// formatBatteryStatusCompact formats battery status without range (for combined view)
func formatBatteryStatusCompact(batteryLevel, chargeTimeACMin, chargeTimeQBCMin float64, pluggedIn, charging, heaterOn, heaterAuto bool) string {
	status := fmt.Sprintf("BATTERY: %.0f%%", batteryLevel)

	// Build status flags
	var flags []string

	if pluggedIn {
		if charging {
			// Show charging time estimates
			timeStr := formatChargeTime(chargeTimeACMin, chargeTimeQBCMin)
			if timeStr != "" {
				flags = append(flags, fmt.Sprintf("charging, %s", timeStr))
			} else {
				flags = append(flags, "charging")
			}
		} else {
			flags = append(flags, "plugged in, not charging")
		}
	}

	// Add heater status
	if heaterOn {
		if heaterAuto {
			flags = append(flags, "battery heater on, auto enabled")
		} else {
			flags = append(flags, "battery heater on")
		}
	} else if heaterAuto {
		flags = append(flags, "battery heater auto enabled")
	}

	if len(flags) > 0 {
		status += fmt.Sprintf(" [%s]", strings.Join(flags, ", "))
	}

	return status
}

// formatFuelStatusWithRange formats fuel status with combined range display for PHEVs
func formatFuelStatusWithRange(fuelLevel, fuelRange, evRange float64) string {
	// If both ranges are the same (common for PHEVs), show as total range
	if fuelRange == evRange {
		return fmt.Sprintf("FUEL: %.0f%% (%.1f km total range)", fuelLevel, fuelRange)
	}
	// If different, show both EV and total (fuelRange is usually total for PHEVs)
	return fmt.Sprintf("FUEL: %.0f%% (%.1f km EV / %.1f km total range)", fuelLevel, evRange, fuelRange)
}

// formatLocationStatus formats location status for display
func formatLocationStatus(lat, lon float64, timestamp string, jsonOutput bool) string {
	mapsURL := fmt.Sprintf("https://maps.google.com/?q=%f,%f", lat, lon)
	if jsonOutput {
		return toJSON(map[string]interface{}{
			"latitude":  lat,
			"longitude": lon,
			"timestamp": timestamp,
			"maps_url":  mapsURL,
		})
	}

	return fmt.Sprintf("LOCATION: %.6f, %.6f\n  %s", lat, lon, mapsURL)
}

// formatTiresStatus formats tire status for display
func formatTiresStatus(fl, fr, rl, rr float64, jsonOutput bool) string {
	if jsonOutput {
		return toJSON(map[string]interface{}{
			"front_left_psi":  fl,
			"front_right_psi": fr,
			"rear_left_psi":   rl,
			"rear_right_psi":  rr,
		})
	}

	return fmt.Sprintf("TIRES: FL:%.1f FR:%.1f RL:%.1f RR:%.1f PSI", fl, fr, rl, rr)
}

// formatDoorsStatus formats door status for display
func formatDoorsStatus(doorStatus api.DoorStatus, jsonOutput bool) string {
	if jsonOutput {
		return toJSON(map[string]interface{}{
			"all_locked":        doorStatus.AllLocked,
			"driver_open":       doorStatus.DriverOpen,
			"passenger_open":    doorStatus.PassengerOpen,
			"rear_left_open":    doorStatus.RearLeftOpen,
			"rear_right_open":   doorStatus.RearRightOpen,
			"trunk_open":        doorStatus.TrunkOpen,
			"hood_open":         doorStatus.HoodOpen,
			"fuel_lid_open":     doorStatus.FuelLidOpen,
			"driver_locked":     doorStatus.DriverLocked,
			"passenger_locked":  doorStatus.PassengerLocked,
			"rear_left_locked":  doorStatus.RearLeftLocked,
			"rear_right_locked": doorStatus.RearRightLocked,
		})
	}

	// If all locked and closed, show simple message
	if doorStatus.AllLocked {
		return "DOORS: All locked"
	}

	// Otherwise, build a list of issues
	var issues []string

	// Check unlocked doors (closed but not locked)
	if !doorStatus.DriverLocked && !doorStatus.DriverOpen {
		issues = append(issues, "Driver unlocked")
	}
	if !doorStatus.PassengerLocked && !doorStatus.PassengerOpen {
		issues = append(issues, "Passenger unlocked")
	}
	if !doorStatus.RearLeftLocked && !doorStatus.RearLeftOpen {
		issues = append(issues, "Rear left unlocked")
	}
	if !doorStatus.RearRightLocked && !doorStatus.RearRightOpen {
		issues = append(issues, "Rear right unlocked")
	}

	// Check open doors/trunk/hood
	if doorStatus.DriverOpen {
		issues = append(issues, "Driver open")
	}
	if doorStatus.PassengerOpen {
		issues = append(issues, "Passenger open")
	}
	if doorStatus.RearLeftOpen {
		issues = append(issues, "Rear left open")
	}
	if doorStatus.RearRightOpen {
		issues = append(issues, "Rear right open")
	}
	if doorStatus.TrunkOpen {
		issues = append(issues, "Trunk open")
	}
	if doorStatus.HoodOpen {
		issues = append(issues, "Hood open")
	}
	if doorStatus.FuelLidOpen {
		issues = append(issues, "Fuel lid open")
	}

	if len(issues) == 0 {
		return "DOORS: Status unknown"
	}

	return fmt.Sprintf("DOORS: %s", strings.Join(issues, ", "))
}

// formatOdometerStatus formats odometer status for display
func formatOdometerStatus(odometerKm float64, jsonOutput bool) string {
	if jsonOutput {
		return toJSON(map[string]interface{}{
			"odometer_km": odometerKm,
		})
	}

	return fmt.Sprintf("ODOMETER: %s km", formatThousands(odometerKm))
}

// formatHvacStatus formats HVAC status for display
func formatHvacStatus(hvacOn, frontDefroster, rearDefroster bool, interiorTempC, targetTempC float64, jsonOutput bool) string {
	if jsonOutput {
		return toJSON(map[string]interface{}{
			"hvac_on":                hvacOn,
			"front_defroster":        frontDefroster,
			"rear_defroster":         rearDefroster,
			"interior_temperature_c": interiorTempC,
			"target_temperature_c":   targetTempC,
		})
	}

	var status string
	if hvacOn {
		// Show current temp → target temp when HVAC is on and temps differ
		if targetTempC > 0 && targetTempC != interiorTempC {
			status = fmt.Sprintf("CLIMATE: On, %.0f°C → %.0f°C", interiorTempC, targetTempC)
		} else {
			status = fmt.Sprintf("CLIMATE: On, %.0f°C", interiorTempC)
		}
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

// formatThousands formats a float with comma separators for thousands
func formatThousands(value float64) string {
	// Format the full number with one decimal place
	formatted := fmt.Sprintf("%.1f", value)

	// Find the decimal point position
	dotPos := -1
	for i, c := range formatted {
		if c == '.' {
			dotPos = i
			break
		}
	}

	if dotPos == -1 {
		return formatted
	}

	// Add commas to the integer part
	intPart := formatted[:dotPos]
	decPart := formatted[dotPos:]

	var result string
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}

	return result + decPart
}

// formatChargeTime formats charging time estimates for display
func formatChargeTime(acMinutes, qbcMinutes float64) string {
	// If both are zero or negative, no charging time info available
	if acMinutes <= 0 && qbcMinutes <= 0 {
		return ""
	}

	// Helper to format minutes as "Xh Ym" or "Xm"
	formatMinutes := func(minutes float64) string {
		if minutes <= 0 {
			return ""
		}
		hours := int(minutes) / 60
		mins := int(minutes) % 60
		if hours > 0 {
			if mins > 0 {
				return fmt.Sprintf("%dh %dm", hours, mins)
			}
			return fmt.Sprintf("%dh", hours)
		}
		return fmt.Sprintf("%dm", mins)
	}

	// If both are available and different, show both
	if qbcMinutes > 0 && acMinutes > 0 && qbcMinutes != acMinutes {
		qbcStr := formatMinutes(qbcMinutes)
		acStr := formatMinutes(acMinutes)
		return fmt.Sprintf("~%s quick / ~%s AC", qbcStr, acStr)
	}

	// Otherwise, show whichever is available
	if qbcMinutes > 0 {
		return fmt.Sprintf("~%s to full", formatMinutes(qbcMinutes))
	}
	if acMinutes > 0 {
		return fmt.Sprintf("~%s to full", formatMinutes(acMinutes))
	}

	return ""
}

// formatWindowsStatus formats window status for display
func formatWindowsStatus(driver, passenger, rearLeft, rearRight float64, jsonOutput bool) string {
	if jsonOutput {
		return toJSON(map[string]interface{}{
			"driver_position":    driver,
			"passenger_position": passenger,
			"rear_left_position": rearLeft,
			"rear_right_position": rearRight,
		})
	}

	// If all windows are closed, show simple message
	if driver == 0 && passenger == 0 && rearLeft == 0 && rearRight == 0 {
		return "WINDOWS: All closed"
	}

	// Otherwise, build a list of open windows with percentages
	var openWindows []string

	if driver > 0 {
		openWindows = append(openWindows, fmt.Sprintf("Driver %.0f%%", driver))
	}
	if passenger > 0 {
		openWindows = append(openWindows, fmt.Sprintf("Passenger %.0f%%", passenger))
	}
	if rearLeft > 0 {
		openWindows = append(openWindows, fmt.Sprintf("Rear left %.0f%%", rearLeft))
	}
	if rearRight > 0 {
		openWindows = append(openWindows, fmt.Sprintf("Rear right %.0f%%", rearRight))
	}

	if len(openWindows) == 0 {
		return "WINDOWS: All closed"
	}

	return fmt.Sprintf("WINDOWS: %s", strings.Join(openWindows, ", "))
}

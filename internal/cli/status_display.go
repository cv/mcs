package cli

import (
	"encoding/json"
	"fmt"

	"github.com/cv/mcs/internal/api"
)

// appendFormattedSection appends a formatted section to the output string with a newline.
// It calls the provided formatter function and handles errors.
func appendFormattedSection(output *string, formatter func() (string, error)) error {
	result, err := formatter()
	if err != nil {
		return err
	}
	*output += result + "\n"

	return nil
}

// displayAllStatusJSON formats all status as JSON.
func displayAllStatusJSON(vehicleStatus *api.VehicleStatusResponse, evStatus *api.EVVehicleStatusResponse, vehicleInfo VehicleInfo) (string, error) {
	hazardsOn, _ := vehicleStatus.GetHazardInfo()
	data := map[string]any{
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
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// displayAllStatusText formats all status as human-readable text.
func displayAllStatusText(vehicleStatus *api.VehicleStatusResponse, evStatus *api.EVVehicleStatusResponse, vehicleInfo VehicleInfo) (string, error) {
	// Get timestamp from EV status
	occurrenceDate, err := evStatus.GetOccurrenceDate()
	if err != nil {
		return "", fmt.Errorf("failed to get occurrence date: %w", err)
	}
	timestamp := formatTimestamp(occurrenceDate)

	// Extract HVAC info
	hvacInfo, err := evStatus.GetHvacInfo()
	if err != nil {
		return "", fmt.Errorf("failed to get HVAC info: %w", err)
	}

	// Extract all other info
	odometerInfo, _ := vehicleStatus.GetOdometerInfo()
	windowsInfo, _ := vehicleStatus.GetWindowsInfo()
	hazardsOn, _ := vehicleStatus.GetHazardInfo()
	batteryInfo, _ := evStatus.GetBatteryInfo()
	fuelInfo, _ := vehicleStatus.GetFuelInfo()
	doorStatus, _ := vehicleStatus.GetDoorsInfo()
	tireInfo, _ := vehicleStatus.GetTiresInfo()

	// Build vehicle header
	output := formatVehicleHeader(vehicleInfo) + "\n"
	output += fmt.Sprintf("Status as of %s\n\n", timestamp)
	output += formatBatteryStatusCompact(batteryInfo) + "\n"
	output += formatFuelStatusWithRange(fuelInfo, batteryInfo) + "\n"

	if err := appendFormattedSection(&output, func() (string, error) {
		return formatHvacStatus(hvacInfo, false)
	}); err != nil {
		return "", err
	}

	if err := appendFormattedSection(&output, func() (string, error) {
		return formatDoorsStatus(doorStatus, false)
	}); err != nil {
		return "", err
	}

	if err := appendFormattedSection(&output, func() (string, error) {
		return formatWindowsStatus(windowsInfo, false)
	}); err != nil {
		return "", err
	}

	// Only show hazards if they're on
	if hazardsOn {
		output += "HAZARDS: On\n"
	}

	if err := appendFormattedSection(&output, func() (string, error) {
		return formatTiresStatus(tireInfo, false)
	}); err != nil {
		return "", err
	}

	// Note: odometer is the last section, so no trailing newline
	odometerOutput, err := formatOdometerStatus(odometerInfo, false)
	if err != nil {
		return "", err
	}
	output += odometerOutput

	return output, nil
}

// displayAllStatus displays all status information.
func displayAllStatus(vehicleStatus *api.VehicleStatusResponse, evStatus *api.EVVehicleStatusResponse, vehicleInfo VehicleInfo, jsonOutput bool) (string, error) {
	if jsonOutput {
		return displayAllStatusJSON(vehicleStatus, evStatus, vehicleInfo)
	}

	return displayAllStatusText(vehicleStatus, evStatus, vehicleInfo)
}

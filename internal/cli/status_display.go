package cli

import (
	"encoding/json"
	"fmt"

	"github.com/cv/mcs/internal/api"
)

// displayAllStatus displays all status information
func displayAllStatus(vehicleStatus *api.VehicleStatusResponse, evStatus *api.EVVehicleStatusResponse, vehicleInfo VehicleInfo, jsonOutput bool) (string, error) {
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
		jsonBytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal JSON: %w", err)
		}
		return string(jsonBytes), nil
	}

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

	// Extract odometer
	odometerInfo, _ := vehicleStatus.GetOdometerInfo()

	// Extract windows info
	windowsInfo, _ := vehicleStatus.GetWindowsInfo()

	// Extract hazard info
	hazardsOn, _ := vehicleStatus.GetHazardInfo()

	// Extract battery and fuel info
	batteryInfo, _ := evStatus.GetBatteryInfo()
	fuelInfo, _ := vehicleStatus.GetFuelInfo()

	// Build vehicle header
	output := "\n" + formatVehicleHeader(vehicleInfo) + "\n"
	output += fmt.Sprintf("Status as of %s\n\n", timestamp)
	output += formatBatteryStatusCompact(batteryInfo) + "\n"
	output += formatFuelStatusWithRange(fuelInfo, batteryInfo.RangeKm) + "\n"

	hvacOutput, err := formatHvacStatus(hvacInfo, false)
	if err != nil {
		return "", err
	}
	output += hvacOutput + "\n"

	doorsOutput, err := displayDoorsStatus(vehicleStatus, false)
	if err != nil {
		return "", err
	}
	output += doorsOutput + "\n"

	windowsOutput, err := formatWindowsStatus(windowsInfo, false)
	if err != nil {
		return "", err
	}
	output += windowsOutput + "\n"

	// Only show hazards if they're on
	if hazardsOn {
		output += "HAZARDS: On\n"
	}

	tiresOutput, err := displayTiresStatus(vehicleStatus, false)
	if err != nil {
		return "", err
	}
	output += tiresOutput + "\n"

	odometerOutput, err := formatOdometerStatus(odometerInfo, false)
	if err != nil {
		return "", err
	}
	output += odometerOutput + "\n"

	return output, nil
}

// displayBatteryStatus displays battery status
func displayBatteryStatus(evStatus *api.EVVehicleStatusResponse, jsonOutput bool) (string, error) {
	batteryInfo, _ := evStatus.GetBatteryInfo()
	return formatBatteryStatus(batteryInfo, jsonOutput)
}

// displayFuelStatus displays fuel status
func displayFuelStatus(vehicleStatus *api.VehicleStatusResponse, jsonOutput bool) (string, error) {
	fuelInfo, _ := vehicleStatus.GetFuelInfo()
	return formatFuelStatus(fuelInfo, jsonOutput)
}

// displayLocationStatus displays location status
func displayLocationStatus(vehicleStatus *api.VehicleStatusResponse, jsonOutput bool) (string, error) {
	locationInfo, _ := vehicleStatus.GetLocationInfo()
	return formatLocationStatus(locationInfo, jsonOutput)
}

// displayTiresStatus displays tire pressure status
func displayTiresStatus(vehicleStatus *api.VehicleStatusResponse, jsonOutput bool) (string, error) {
	tireInfo, _ := vehicleStatus.GetTiresInfo()
	return formatTiresStatus(tireInfo, jsonOutput)
}

// displayDoorsStatus displays door lock status
func displayDoorsStatus(vehicleStatus *api.VehicleStatusResponse, jsonOutput bool) (string, error) {
	doorStatus, _ := vehicleStatus.GetDoorsInfo()
	return formatDoorsStatus(doorStatus, jsonOutput)
}

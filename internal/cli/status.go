package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cv/cx90/internal/api"
	"github.com/cv/cx90/internal/config"
	"github.com/spf13/cobra"
)

// NewStatusCmd creates the status command
func NewStatusCmd() *cobra.Command {
	var jsonOutput bool

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show vehicle status",
		Long:  `Show comprehensive vehicle status including battery, fuel, location, tires, and doors.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, "all")
		},
		SilenceUsage: true,
	}

	// Add persistent JSON flag
	statusCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	// Add subcommands
	statusCmd.AddCommand(&cobra.Command{
		Use:   "battery",
		Short: "Show battery status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, "battery")
		},
		SilenceUsage: true,
	})

	statusCmd.AddCommand(&cobra.Command{
		Use:   "fuel",
		Short: "Show fuel status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, "fuel")
		},
		SilenceUsage: true,
	})

	statusCmd.AddCommand(&cobra.Command{
		Use:   "location",
		Short: "Show vehicle location",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, "location")
		},
		SilenceUsage: true,
	})

	statusCmd.AddCommand(&cobra.Command{
		Use:   "tires",
		Short: "Show tire pressure",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, "tires")
		},
		SilenceUsage: true,
	})

	statusCmd.AddCommand(&cobra.Command{
		Use:   "doors",
		Short: "Show door lock status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd, jsonOutput, "doors")
		},
		SilenceUsage: true,
	})

	return statusCmd
}

// runStatus executes the status command
func runStatus(cmd *cobra.Command, jsonOutput bool, statusType string) error {
	// Load configuration
	cfg, err := config.Load(ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Create API client
	client, err := api.NewClient(cfg.Email, cfg.Password, cfg.Region)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Get vehicle base info to retrieve internal VIN
	vecBaseInfos, err := client.GetVecBaseInfos()
	if err != nil {
		return fmt.Errorf("failed to get vehicle info: %w", err)
	}

	internalVIN, err := getInternalVIN(vecBaseInfos)
	if err != nil {
		return err
	}

	// Get vehicle status
	vehicleStatus, err := client.GetVehicleStatus(internalVIN)
	if err != nil {
		return fmt.Errorf("failed to get vehicle status: %w", err)
	}

	// Get EV status
	evStatus, err := client.GetEVVehicleStatus(internalVIN)
	if err != nil {
		return fmt.Errorf("failed to get EV status: %w", err)
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
		output := displayAllStatus(vecBaseInfos, vehicleStatus, evStatus, jsonOutput)
		fmt.Fprint(cmd.OutOrStdout(), output)
	}

	return nil
}

// getInternalVIN extracts the internal VIN from vehicle base info
func getInternalVIN(vecBaseInfos map[string]interface{}) (string, error) {
	vecInfos, ok := vecBaseInfos["vecBaseInfos"].([]interface{})
	if !ok || len(vecInfos) == 0 {
		return "", fmt.Errorf("no vehicles found")
	}

	firstVehicle := vecInfos[0].(map[string]interface{})
	vehicle := firstVehicle["Vehicle"].(map[string]interface{})
	cvInfo := vehicle["CvInformation"].(map[string]interface{})
	internalVIN := cvInfo["internalVin"].(string)

	return internalVIN, nil
}

// displayAllStatus displays all status information
func displayAllStatus(vecBaseInfos, vehicleStatus, evStatus map[string]interface{}, jsonOutput bool) string {
	if jsonOutput {
		data := map[string]interface{}{
			"battery":  extractBatteryData(evStatus),
			"fuel":     extractFuelData(vehicleStatus),
			"location": extractLocationData(vehicleStatus),
			"tires":    extractTiresData(vehicleStatus),
			"doors":    extractDoorsData(vehicleStatus),
		}
		jsonBytes, _ := json.MarshalIndent(data, "", "  ")
		return string(jsonBytes)
	}

	// Get timestamp from EV status
	timestamp := getTimestamp(evStatus)

	output := "\nCX-90 GT PHEV (2025)\n"
	output += fmt.Sprintf("Last Updated: %s\n\n", timestamp)
	output += displayBatteryStatus(evStatus, false) + "\n"
	output += displayFuelStatus(vehicleStatus, false) + "\n"
	output += displayDoorsStatus(vehicleStatus, false) + "\n"
	output += displayTiresStatus(vehicleStatus, false) + "\n"

	return output
}

// displayBatteryStatus displays battery status
func displayBatteryStatus(evStatus map[string]interface{}, jsonOutput bool) string {
	batteryLevel, range_, pluggedIn, charging := extractBatteryInfo(evStatus)
	return formatBatteryStatus(batteryLevel, range_, pluggedIn, charging, jsonOutput)
}

// displayFuelStatus displays fuel status
func displayFuelStatus(vehicleStatus map[string]interface{}, jsonOutput bool) string {
	fuelLevel, range_ := extractFuelInfo(vehicleStatus)
	return formatFuelStatus(fuelLevel, range_, jsonOutput)
}

// displayLocationStatus displays location status
func displayLocationStatus(vehicleStatus map[string]interface{}, jsonOutput bool) string {
	lat, lon, timestamp := extractLocationInfo(vehicleStatus)
	return formatLocationStatus(lat, lon, timestamp, jsonOutput)
}

// displayTiresStatus displays tire pressure status
func displayTiresStatus(vehicleStatus map[string]interface{}, jsonOutput bool) string {
	fl, fr, rl, rr := extractTiresInfo(vehicleStatus)
	return formatTiresStatus(fl, fr, rl, rr, jsonOutput)
}

// displayDoorsStatus displays door lock status
func displayDoorsStatus(vehicleStatus map[string]interface{}, jsonOutput bool) string {
	allLocked := extractDoorsInfo(vehicleStatus)
	return formatDoorsStatus(allLocked, jsonOutput)
}

// extractBatteryInfo extracts battery information from EV status
func extractBatteryInfo(evStatus map[string]interface{}) (batteryLevel, range_ float64, pluggedIn, charging bool) {
	resultData := evStatus["resultData"].([]interface{})
	firstResult := resultData[0].(map[string]interface{})
	plusBInfo := firstResult["PlusBInformation"].(map[string]interface{})
	vehicleInfo := plusBInfo["VehicleInfo"].(map[string]interface{})
	chargeInfo := vehicleInfo["ChargeInfo"].(map[string]interface{})

	batteryLevel = chargeInfo["SmaphSOC"].(float64)
	range_ = chargeInfo["SmaphRemDrvDistKm"].(float64)
	pluggedIn = int(chargeInfo["ChargerConnectorFitting"].(float64)) == 1
	charging = int(chargeInfo["ChargeStatusSub"].(float64)) == 6

	return
}

// extractFuelInfo extracts fuel information from vehicle status
func extractFuelInfo(vehicleStatus map[string]interface{}) (fuelLevel, range_ float64) {
	remoteInfos := vehicleStatus["remoteInfos"].([]interface{})
	firstInfo := remoteInfos[0].(map[string]interface{})
	residualFuel := firstInfo["ResidualFuel"].(map[string]interface{})

	fuelLevel = residualFuel["FuelSegementDActl"].(float64)
	range_ = residualFuel["RemDrvDistDActlKm"].(float64)

	return
}

// extractLocationInfo extracts location information from vehicle status
func extractLocationInfo(vehicleStatus map[string]interface{}) (lat, lon float64, timestamp string) {
	remoteInfos := vehicleStatus["remoteInfos"].([]interface{})
	firstInfo := remoteInfos[0].(map[string]interface{})
	positionInfo := firstInfo["PositionInfo"].(map[string]interface{})

	lat = positionInfo["Latitude"].(float64)
	lon = positionInfo["Longitude"].(float64)
	timestamp = positionInfo["AcquisitionDatetime"].(string)

	return
}

// extractTiresInfo extracts tire pressure information from vehicle status
func extractTiresInfo(vehicleStatus map[string]interface{}) (fl, fr, rl, rr float64) {
	remoteInfos := vehicleStatus["remoteInfos"].([]interface{})
	firstInfo := remoteInfos[0].(map[string]interface{})
	tpmsInfo := firstInfo["TPMSInformation"].(map[string]interface{})

	fl = tpmsInfo["FLTPrsDispPsi"].(float64)
	fr = tpmsInfo["FRTPrsDispPsi"].(float64)
	rl = tpmsInfo["RLTPrsDispPsi"].(float64)
	rr = tpmsInfo["RRTPrsDispPsi"].(float64)

	return
}

// extractDoorsInfo extracts door lock information from vehicle status
func extractDoorsInfo(vehicleStatus map[string]interface{}) bool {
	alertInfos := vehicleStatus["alertInfos"].([]interface{})
	firstAlert := alertInfos[0].(map[string]interface{})
	door := firstAlert["Door"].(map[string]interface{})

	// Check if all doors are locked (0 = locked)
	allLocked := door["DrStatDrv"].(float64) == 0 &&
		door["DrStatPsngr"].(float64) == 0 &&
		door["DrStatRl"].(float64) == 0 &&
		door["DrStatRr"].(float64) == 0 &&
		door["DrStatTrnkLg"].(float64) == 0

	return allLocked
}

// extractBatteryData extracts battery data for JSON output
func extractBatteryData(evStatus map[string]interface{}) map[string]interface{} {
	batteryLevel, range_, pluggedIn, charging := extractBatteryInfo(evStatus)
	return map[string]interface{}{
		"battery_level": batteryLevel,
		"range_km":      range_,
		"plugged_in":    pluggedIn,
		"charging":      charging,
	}
}

// extractFuelData extracts fuel data for JSON output
func extractFuelData(vehicleStatus map[string]interface{}) map[string]interface{} {
	fuelLevel, range_ := extractFuelInfo(vehicleStatus)
	return map[string]interface{}{
		"fuel_level": fuelLevel,
		"range_km":   range_,
	}
}

// extractLocationData extracts location data for JSON output
func extractLocationData(vehicleStatus map[string]interface{}) map[string]interface{} {
	lat, lon, timestamp := extractLocationInfo(vehicleStatus)
	return map[string]interface{}{
		"latitude":  lat,
		"longitude": lon,
		"timestamp": timestamp,
	}
}

// extractTiresData extracts tire data for JSON output
func extractTiresData(vehicleStatus map[string]interface{}) map[string]interface{} {
	fl, fr, rl, rr := extractTiresInfo(vehicleStatus)
	return map[string]interface{}{
		"front_left_psi":  fl,
		"front_right_psi": fr,
		"rear_left_psi":   rl,
		"rear_right_psi":  rr,
	}
}

// extractDoorsData extracts door data for JSON output
func extractDoorsData(vehicleStatus map[string]interface{}) map[string]interface{} {
	allLocked := extractDoorsInfo(vehicleStatus)
	return map[string]interface{}{
		"all_locked": allLocked,
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
		data := map[string]interface{}{
			"latitude":  lat,
			"longitude": lon,
			"timestamp": timestamp,
		}
		jsonBytes, _ := json.MarshalIndent(data, "", "  ")
		return string(jsonBytes)
	}

	return fmt.Sprintf("LOCATION: %.6f, %.6f (updated: %s)", lat, lon, formatTimestamp(timestamp))
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

// getTimestamp extracts and formats timestamp from EV status
func getTimestamp(evStatus map[string]interface{}) string {
	resultData := evStatus["resultData"].([]interface{})
	firstResult := resultData[0].(map[string]interface{})
	timestamp := firstResult["OccurrenceDate"].(string)
	return formatTimestamp(timestamp)
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

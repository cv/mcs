package cli

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cv/mcs/internal/api"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

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

// toJSON converts a map to formatted JSON string
func toJSON(data map[string]any) (string, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// buildBatteryStatusFlags builds the status flags for battery display
func buildBatteryStatusFlags(batteryInfo api.BatteryInfo) []string {
	var flags []string

	if batteryInfo.PluggedIn {
		if batteryInfo.Charging {
			// Show charging time estimates
			timeStr := formatChargeTime(batteryInfo.ChargeTimeACMin, batteryInfo.ChargeTimeQBCMin)
			if timeStr != "" {
				flags = append(flags, "charging, "+timeStr)
			} else {
				flags = append(flags, "charging")
			}
		} else {
			flags = append(flags, "plugged in, not charging")
		}
	}

	// Add heater status
	if batteryInfo.HeaterOn {
		if batteryInfo.HeaterAuto {
			flags = append(flags, "battery heater on, auto enabled")
		} else {
			flags = append(flags, "battery heater on")
		}
	} else if batteryInfo.HeaterAuto {
		flags = append(flags, "battery heater auto enabled")
	}

	return flags
}

// formatBatteryStatus formats battery status for display
func formatBatteryStatus(batteryInfo api.BatteryInfo, jsonOutput bool) (string, error) {
	if jsonOutput {
		return toJSON(batteryInfoToMap(batteryInfo))
	}

	// Create progress bar and format percentage/range
	progressBar := ProgressBar(batteryInfo.BatteryLevel, 10)
	status := fmt.Sprintf("BATTERY: %s (%.1f km range)", progressBar, batteryInfo.RangeKm)

	// Build status flags
	flags := buildBatteryStatusFlags(batteryInfo)

	if len(flags) > 0 {
		status += fmt.Sprintf(" [%s]", strings.Join(flags, ", "))
	}

	return status, nil
}

// formatFuelStatus formats fuel status for display
func formatFuelStatus(fuelInfo api.FuelInfo, jsonOutput bool) (string, error) {
	if jsonOutput {
		return toJSON(fuelInfoToMap(fuelInfo))
	}

	progressBar := ProgressBar(fuelInfo.FuelLevel, 10)
	return fmt.Sprintf("FUEL: %s (%.1f km range)", progressBar, fuelInfo.RangeKm), nil
}

// formatBatteryStatusCompact formats battery status without range (for combined view)
func formatBatteryStatusCompact(batteryInfo api.BatteryInfo) string {
	progressBar := ProgressBar(batteryInfo.BatteryLevel, 10)
	status := "BATTERY: " + progressBar

	// Build status flags
	flags := buildBatteryStatusFlags(batteryInfo)

	if len(flags) > 0 {
		status += fmt.Sprintf(" [%s]", strings.Join(flags, ", "))
	}

	return status
}

// formatFuelStatusWithRange formats fuel status with range display for PHEVs
// For PHEVs: RemDrvDistDActlKm (fuel API) = total range, SmaphRemDrvDistKm (EV API) = fuel-only range
// EV range = total - fuel-only
func formatFuelStatusWithRange(fuelInfo api.FuelInfo, batteryInfo api.BatteryInfo) string {
	progressBar := ProgressBar(fuelInfo.FuelLevel, 10)
	// Calculate EV range as difference between total and fuel-only
	// batteryInfo.RangeKm represents the fuel-only range for PHEVs
	evRange := fuelInfo.RangeKm - batteryInfo.RangeKm
	if evRange > 0.5 { // Only show EV range if meaningful (> 0.5 km)
		return fmt.Sprintf("FUEL: %s (%.0f km EV + %.0f km fuel = %.0f km total)",
			progressBar, evRange, batteryInfo.RangeKm, fuelInfo.RangeKm)
	}
	return fmt.Sprintf("FUEL: %s (%.1f km range)", progressBar, fuelInfo.RangeKm)
}

// formatLocationStatus formats location status for display
func formatLocationStatus(locationInfo api.LocationInfo, jsonOutput bool) (string, error) {
	mapsURL := fmt.Sprintf("https://maps.google.com/?q=%f,%f", locationInfo.Latitude, locationInfo.Longitude)
	if jsonOutput {
		return toJSON(locationInfoToMap(locationInfo))
	}

	return fmt.Sprintf("LOCATION: %.6f, %.6f\n  %s", locationInfo.Latitude, locationInfo.Longitude, mapsURL), nil
}

// formatTiresStatus formats tire status for display
func formatTiresStatus(tireInfo api.TireInfo, jsonOutput bool) (string, error) {
	if jsonOutput {
		return toJSON(tireInfoToMap(tireInfo))
	}

	// Color code each tire pressure based on deviation from recommended (36 PSI for Mazda CX-90)
	target := defaultTargetPressurePSI
	fl := ColorPressure(tireInfo.FrontLeftPsi, target)
	fr := ColorPressure(tireInfo.FrontRightPsi, target)
	rl := ColorPressure(tireInfo.RearLeftPsi, target)
	rr := ColorPressure(tireInfo.RearRightPsi, target)

	return fmt.Sprintf("TIRES: FL:%s FR:%s RL:%s RR:%s PSI", fl, fr, rl, rr), nil
}

// doorPosition describes a single door position for status checking
type doorPosition struct {
	name     string
	isOpen   bool
	isLocked bool
	hasLock  bool // trunk/hood/fuel lid don't have locks
}

// formatDoorsStatus formats door status for display
func formatDoorsStatus(doorStatus api.DoorStatus, jsonOutput bool) (string, error) {
	if jsonOutput {
		return toJSON(doorStatusToMap(doorStatus))
	}

	// If all locked and closed, show simple message
	if doorStatus.AllLocked {
		return "DOORS: " + Green("All locked"), nil
	}

	// Define all door positions to check
	doors := []doorPosition{
		{"Driver", doorStatus.DriverOpen, doorStatus.DriverLocked, true},
		{"Passenger", doorStatus.PassengerOpen, doorStatus.PassengerLocked, true},
		{"Rear left", doorStatus.RearLeftOpen, doorStatus.RearLeftLocked, true},
		{"Rear right", doorStatus.RearRightOpen, doorStatus.RearRightLocked, true},
		{"Trunk", doorStatus.TrunkOpen, false, false},
		{"Hood", doorStatus.HoodOpen, false, false},
		{"Fuel lid", doorStatus.FuelLidOpen, false, false},
	}

	// Build a list of issues
	var issues []string

	for _, door := range doors {
		// Check unlocked doors (closed but not locked)
		if door.hasLock && !door.isLocked && !door.isOpen {
			issues = append(issues, Yellow(door.name+" unlocked"))
		}

		// Check open doors/trunk/hood/fuel lid
		if door.isOpen {
			issues = append(issues, Red(door.name+" open"))
		}
	}

	if len(issues) == 0 {
		return "DOORS: Status unknown", nil
	}

	return "DOORS: " + strings.Join(issues, ", "), nil
}

// formatOdometerStatus formats odometer status for display
func formatOdometerStatus(odometerInfo api.OdometerInfo, jsonOutput bool) (string, error) {
	if jsonOutput {
		return toJSON(odometerInfoToMap(odometerInfo))
	}

	return fmt.Sprintf("ODOMETER: %s km", formatThousands(odometerInfo.OdometerKm)), nil
}

// formatHvacStatus formats HVAC status for display
func formatHvacStatus(hvacInfo api.HVACInfo, jsonOutput bool) (string, error) {
	if jsonOutput {
		return toJSON(hvacInfoToMap(hvacInfo))
	}

	var status string
	if hvacInfo.HVACOn {
		// Show current temp → target temp when HVAC is on and temps differ
		if hvacInfo.TargetTempC > 0 && hvacInfo.TargetTempC != hvacInfo.InteriorTempC {
			status = fmt.Sprintf("CLIMATE: On, %.0f°C → %.0f°C", hvacInfo.InteriorTempC, hvacInfo.TargetTempC)
		} else {
			status = fmt.Sprintf("CLIMATE: On, %.0f°C", hvacInfo.InteriorTempC)
		}
	} else {
		status = fmt.Sprintf("CLIMATE: Off, %.0f°C", hvacInfo.InteriorTempC)
	}

	// Build defroster status
	var defrosters []string
	if hvacInfo.FrontDefroster {
		defrosters = append(defrosters, "front")
	}
	if hvacInfo.RearDefroster {
		defrosters = append(defrosters, "rear")
	}

	if len(defrosters) == 2 {
		status += " (front and rear defrosters on)"
	} else if len(defrosters) == 1 {
		status += fmt.Sprintf(" (%s defroster on)", defrosters[0])
	}

	return status, nil
}

// formatRelativeTime returns a human-friendly relative time string
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	// Handle future times (shouldn't happen, but be safe)
	if diff < 0 {
		return "just now"
	}

	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := hours / 24

	switch {
	case seconds < 60:
		return fmt.Sprintf("%d sec ago", seconds)
	case minutes < 60:
		return fmt.Sprintf("%d min ago", minutes)
	case hours < 24:
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	default:
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}
}

// formatTimestamp converts timestamp from API format to readable format with relative time
func formatTimestamp(timestamp string) string {
	// API returns timestamp in format: YYYYMMDDHHmmss
	// Convert to: YYYY-MM-DD HH:mm:ss (X ago)
	if len(timestamp) != 14 {
		return timestamp
	}

	t, err := time.Parse("20060102150405", timestamp)
	if err != nil {
		return timestamp
	}

	return fmt.Sprintf("%s (%s)", t.Format("2006-01-02 15:04:05"), formatRelativeTime(t))
}

// formatThousands formats a float with comma separators for thousands
func formatThousands(value float64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%.1f", value)
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
func formatWindowsStatus(windowsInfo api.WindowStatus, jsonOutput bool) (string, error) {
	if jsonOutput {
		return toJSON(windowStatusToMap(windowsInfo))
	}

	// If all windows are closed, show simple message
	if windowsInfo.DriverPosition == api.WindowClosed && windowsInfo.PassengerPosition == api.WindowClosed &&
		windowsInfo.RearLeftPosition == api.WindowClosed && windowsInfo.RearRightPosition == api.WindowClosed {
		return "WINDOWS: " + Green("All closed"), nil
	}

	// Otherwise, build a list of open windows with percentages
	var openWindows []string

	if windowsInfo.DriverPosition > api.WindowClosed {
		openWindows = append(openWindows, Yellow(fmt.Sprintf("Driver %.0f%%", windowsInfo.DriverPosition)))
	}
	if windowsInfo.PassengerPosition > api.WindowClosed {
		openWindows = append(openWindows, Yellow(fmt.Sprintf("Passenger %.0f%%", windowsInfo.PassengerPosition)))
	}
	if windowsInfo.RearLeftPosition > api.WindowClosed {
		openWindows = append(openWindows, Yellow(fmt.Sprintf("Rear left %.0f%%", windowsInfo.RearLeftPosition)))
	}
	if windowsInfo.RearRightPosition > api.WindowClosed {
		openWindows = append(openWindows, Yellow(fmt.Sprintf("Rear right %.0f%%", windowsInfo.RearRightPosition)))
	}

	if len(openWindows) == 0 {
		return "WINDOWS: " + Green("All closed"), nil
	}

	return "WINDOWS: " + strings.Join(openWindows, ", "), nil
}

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
func toJSON(data map[string]interface{}) (string, error) {
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
				flags = append(flags, fmt.Sprintf("charging, %s", timeStr))
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
		data := map[string]interface{}{
			"battery_level": batteryInfo.BatteryLevel,
			"range_km":      batteryInfo.RangeKm,
			"plugged_in":    batteryInfo.PluggedIn,
			"charging":      batteryInfo.Charging,
			"heater_on":     batteryInfo.HeaterOn,
			"heater_auto":   batteryInfo.HeaterAuto,
		}
		if batteryInfo.Charging {
			data["charge_time_ac_minutes"] = batteryInfo.ChargeTimeACMin
			data["charge_time_qbc_minutes"] = batteryInfo.ChargeTimeQBCMin
		}
		return toJSON(data)
	}

	status := fmt.Sprintf("BATTERY: %.0f%% (%.1f km range)", batteryInfo.BatteryLevel, batteryInfo.RangeKm)

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
		return toJSON(map[string]interface{}{
			"fuel_level": fuelInfo.FuelLevel,
			"range_km":   fuelInfo.RangeKm,
		})
	}

	return fmt.Sprintf("FUEL: %.0f%% (%.1f km range)", fuelInfo.FuelLevel, fuelInfo.RangeKm), nil
}

// formatBatteryStatusCompact formats battery status without range (for combined view)
func formatBatteryStatusCompact(batteryInfo api.BatteryInfo) string {
	status := fmt.Sprintf("BATTERY: %.0f%%", batteryInfo.BatteryLevel)

	// Build status flags
	flags := buildBatteryStatusFlags(batteryInfo)

	if len(flags) > 0 {
		status += fmt.Sprintf(" [%s]", strings.Join(flags, ", "))
	}

	return status
}

// formatFuelStatusWithRange formats fuel status with combined range display for PHEVs
func formatFuelStatusWithRange(fuelInfo api.FuelInfo, evRange float64) string {
	// If both ranges are the same (common for PHEVs), show as total range
	if fuelInfo.RangeKm == evRange {
		return fmt.Sprintf("FUEL: %.0f%% (%.1f km total range)", fuelInfo.FuelLevel, fuelInfo.RangeKm)
	}
	// If different, show both EV and total (fuelRange is usually total for PHEVs)
	return fmt.Sprintf("FUEL: %.0f%% (%.1f km EV / %.1f km total range)", fuelInfo.FuelLevel, evRange, fuelInfo.RangeKm)
}

// formatLocationStatus formats location status for display
func formatLocationStatus(locationInfo api.LocationInfo, jsonOutput bool) (string, error) {
	mapsURL := fmt.Sprintf("https://maps.google.com/?q=%f,%f", locationInfo.Latitude, locationInfo.Longitude)
	if jsonOutput {
		return toJSON(map[string]interface{}{
			"latitude":  locationInfo.Latitude,
			"longitude": locationInfo.Longitude,
			"timestamp": locationInfo.Timestamp,
			"maps_url":  mapsURL,
		})
	}

	return fmt.Sprintf("LOCATION: %.6f, %.6f\n  %s", locationInfo.Latitude, locationInfo.Longitude, mapsURL), nil
}

// formatTiresStatus formats tire status for display
func formatTiresStatus(tireInfo api.TireInfo, jsonOutput bool) (string, error) {
	if jsonOutput {
		return toJSON(map[string]interface{}{
			"front_left_psi":  tireInfo.FrontLeftPsi,
			"front_right_psi": tireInfo.FrontRightPsi,
			"rear_left_psi":   tireInfo.RearLeftPsi,
			"rear_right_psi":  tireInfo.RearRightPsi,
		})
	}

	return fmt.Sprintf("TIRES: FL:%.1f FR:%.1f RL:%.1f RR:%.1f PSI",
		tireInfo.FrontLeftPsi, tireInfo.FrontRightPsi, tireInfo.RearLeftPsi, tireInfo.RearRightPsi), nil
}

// formatDoorsStatus formats door status for display
func formatDoorsStatus(doorStatus api.DoorStatus, jsonOutput bool) (string, error) {
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
		return "DOORS: All locked", nil
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
		return "DOORS: Status unknown", nil
	}

	return fmt.Sprintf("DOORS: %s", strings.Join(issues, ", ")), nil
}

// formatOdometerStatus formats odometer status for display
func formatOdometerStatus(odometerInfo api.OdometerInfo, jsonOutput bool) (string, error) {
	if jsonOutput {
		return toJSON(map[string]interface{}{
			"odometer_km": odometerInfo.OdometerKm,
		})
	}

	return fmt.Sprintf("ODOMETER: %s km", formatThousands(odometerInfo.OdometerKm)), nil
}

// formatHvacStatus formats HVAC status for display
func formatHvacStatus(hvacInfo api.HVACInfo, jsonOutput bool) (string, error) {
	if jsonOutput {
		return toJSON(map[string]interface{}{
			"hvac_on":                hvacInfo.HVACOn,
			"front_defroster":        hvacInfo.FrontDefroster,
			"rear_defroster":         hvacInfo.RearDefroster,
			"interior_temperature_c": hvacInfo.InteriorTempC,
			"target_temperature_c":   hvacInfo.TargetTempC,
		})
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
		return toJSON(map[string]interface{}{
			"driver_position":     windowsInfo.DriverPosition,
			"passenger_position":  windowsInfo.PassengerPosition,
			"rear_left_position":  windowsInfo.RearLeftPosition,
			"rear_right_position": windowsInfo.RearRightPosition,
		})
	}

	// If all windows are closed, show simple message
	if windowsInfo.DriverPosition == api.WindowClosed && windowsInfo.PassengerPosition == api.WindowClosed &&
		windowsInfo.RearLeftPosition == api.WindowClosed && windowsInfo.RearRightPosition == api.WindowClosed {
		return "WINDOWS: All closed", nil
	}

	// Otherwise, build a list of open windows with percentages
	var openWindows []string

	if windowsInfo.DriverPosition > api.WindowClosed {
		openWindows = append(openWindows, fmt.Sprintf("Driver %.0f%%", windowsInfo.DriverPosition))
	}
	if windowsInfo.PassengerPosition > api.WindowClosed {
		openWindows = append(openWindows, fmt.Sprintf("Passenger %.0f%%", windowsInfo.PassengerPosition))
	}
	if windowsInfo.RearLeftPosition > api.WindowClosed {
		openWindows = append(openWindows, fmt.Sprintf("Rear left %.0f%%", windowsInfo.RearLeftPosition))
	}
	if windowsInfo.RearRightPosition > api.WindowClosed {
		openWindows = append(openWindows, fmt.Sprintf("Rear right %.0f%%", windowsInfo.RearRightPosition))
	}

	if len(openWindows) == 0 {
		return "WINDOWS: All closed", nil
	}

	return fmt.Sprintf("WINDOWS: %s", strings.Join(openWindows, ", ")), nil
}

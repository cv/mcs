package cli

import (
	"fmt"

	"github.com/cv/mcs/internal/api"
)

// extractVehicleInfoData extracts vehicle info for JSON output
func extractVehicleInfoData(vehicleInfo VehicleInfo) map[string]interface{} {
	return map[string]interface{}{
		"vin":        vehicleInfo.VIN,
		"nickname":   vehicleInfo.Nickname,
		"model_name": vehicleInfo.ModelName,
		"model_year": vehicleInfo.ModelYear,
	}
}

// batteryInfoToMap converts BatteryInfo to a map for JSON output
func batteryInfoToMap(batteryInfo api.BatteryInfo) map[string]interface{} {
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
	return data
}

// extractBatteryData extracts battery data for JSON output
func extractBatteryData(evStatus *api.EVVehicleStatusResponse) map[string]interface{} {
	batteryInfo, _ := evStatus.GetBatteryInfo()
	return batteryInfoToMap(batteryInfo)
}

// fuelInfoToMap converts FuelInfo to a map for JSON output
func fuelInfoToMap(fuelInfo api.FuelInfo) map[string]interface{} {
	return map[string]interface{}{
		"fuel_level": fuelInfo.FuelLevel,
		"range_km":   fuelInfo.RangeKm,
	}
}

// extractFuelData extracts fuel data for JSON output
func extractFuelData(vehicleStatus *api.VehicleStatusResponse) map[string]interface{} {
	fuelInfo, _ := vehicleStatus.GetFuelInfo()
	return fuelInfoToMap(fuelInfo)
}

// locationInfoToMap converts LocationInfo to a map for JSON output
func locationInfoToMap(locationInfo api.LocationInfo) map[string]interface{} {
	mapsURL := fmt.Sprintf("https://maps.google.com/?q=%f,%f", locationInfo.Latitude, locationInfo.Longitude)
	return map[string]interface{}{
		"latitude":  locationInfo.Latitude,
		"longitude": locationInfo.Longitude,
		"timestamp": locationInfo.Timestamp,
		"maps_url":  mapsURL,
	}
}

// extractLocationData extracts location data for JSON output
func extractLocationData(vehicleStatus *api.VehicleStatusResponse) map[string]interface{} {
	locationInfo, _ := vehicleStatus.GetLocationInfo()
	return locationInfoToMap(locationInfo)
}

// tireInfoToMap converts TireInfo to a map for JSON output
func tireInfoToMap(tireInfo api.TireInfo) map[string]interface{} {
	return map[string]interface{}{
		"front_left_psi":  tireInfo.FrontLeftPsi,
		"front_right_psi": tireInfo.FrontRightPsi,
		"rear_left_psi":   tireInfo.RearLeftPsi,
		"rear_right_psi":  tireInfo.RearRightPsi,
	}
}

// extractTiresData extracts tire data for JSON output
func extractTiresData(vehicleStatus *api.VehicleStatusResponse) map[string]interface{} {
	tireInfo, _ := vehicleStatus.GetTiresInfo()
	return tireInfoToMap(tireInfo)
}

// doorStatusToMap converts DoorStatus to a map for JSON output
func doorStatusToMap(doorStatus api.DoorStatus) map[string]interface{} {
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

// extractDoorsData extracts door data for JSON output
func extractDoorsData(vehicleStatus *api.VehicleStatusResponse) map[string]interface{} {
	doorStatus, _ := vehicleStatus.GetDoorsInfo()
	return doorStatusToMap(doorStatus)
}

// odometerInfoToMap converts OdometerInfo to a map for JSON output
func odometerInfoToMap(odometerInfo api.OdometerInfo) map[string]interface{} {
	return map[string]interface{}{
		"odometer_km": odometerInfo.OdometerKm,
	}
}

// extractOdometerData extracts odometer data for JSON output
func extractOdometerData(vehicleStatus *api.VehicleStatusResponse) map[string]interface{} {
	odometerInfo, _ := vehicleStatus.GetOdometerInfo()
	return odometerInfoToMap(odometerInfo)
}

// hvacInfoToMap converts HVACInfo to a map for JSON output
func hvacInfoToMap(hvacInfo api.HVACInfo) map[string]interface{} {
	return map[string]interface{}{
		"hvac_on":                hvacInfo.HVACOn,
		"front_defroster":        hvacInfo.FrontDefroster,
		"rear_defroster":         hvacInfo.RearDefroster,
		"interior_temperature_c": hvacInfo.InteriorTempC,
		"target_temperature_c":   hvacInfo.TargetTempC,
	}
}

// extractHvacData extracts HVAC data for JSON output
func extractHvacData(evStatus *api.EVVehicleStatusResponse) map[string]interface{} {
	hvacInfo, _ := evStatus.GetHvacInfo()
	return hvacInfoToMap(hvacInfo)
}

// windowStatusToMap converts WindowStatus to a map for JSON output
func windowStatusToMap(windowsInfo api.WindowStatus) map[string]interface{} {
	return map[string]interface{}{
		"driver_position":     windowsInfo.DriverPosition,
		"passenger_position":  windowsInfo.PassengerPosition,
		"rear_left_position":  windowsInfo.RearLeftPosition,
		"rear_right_position": windowsInfo.RearRightPosition,
	}
}

// extractWindowsData extracts window data for JSON output
func extractWindowsData(vehicleStatus *api.VehicleStatusResponse) map[string]interface{} {
	windowsInfo, _ := vehicleStatus.GetWindowsInfo()
	return windowStatusToMap(windowsInfo)
}

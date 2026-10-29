package mqttd

import (
	"math"
	"strings"
	"time"
)

// Timestamps as Home Assistant reads them (the same text the Python poller wrote).
const (
	isoSecond = "2006-01-02T15:04:05+00:00"
	isoMicro  = "2006-01-02T15:04:05.000000+00:00"
)

func obj(v any) map[string]any {
	m, _ := v.(map[string]any)

	return m
}

// num is a finite JSON number within [lo, hi], else nil.
func num(v any, lo, hi float64) any {
	f, ok := v.(float64)
	if !ok || math.IsNaN(f) || math.IsInf(f, 0) || f < lo || f > hi {
		return nil
	}

	return f
}

func flag(v any) any {
	b, ok := v.(bool)
	if !ok {
		return nil
	}

	return b
}

// bit reads the health report's 0/1 flags.
func bit(v any) any {
	f, ok := v.(float64)
	switch {
	case !ok:
		return nil
	case f == 0:
		return false
	case f == 1:
		return true
	}

	return nil
}

func not(v any) any {
	if b, ok := v.(bool); ok {
		return !b
	}

	return nil
}

func isTrue(v any) bool { b, _ := v.(bool); return b }

// reportTime turns the API's YYYYMMDDhhmmss (UTC) into an ISO time, else nil.
func reportTime(v any) any {
	s, ok := v.(string)
	if !ok || len(s) != 14 || strings.Trim(s, "0123456789") != "" {
		return nil
	}
	t, err := time.Parse("20060102150405", s)
	if err != nil {
		return nil
	}

	return t.UTC().Format(isoSecond)
}

// Presence is home / not_home, or nil; the coordinates never leave this function.
func Presence(location map[string]any, home *Home) any {
	if home == nil || location == nil {
		return nil
	}
	lat, lon := num(location["latitude"], -90, 90), num(location["longitude"], -180, 180)
	if lat == nil || lon == nil || (lat.(float64) == 0 && lon.(float64) == 0) {
		return nil
	}
	rad := math.Pi / 180
	p1, p2 := lat.(float64)*rad, home.Lat*rad
	a := math.Pow(math.Sin((p2-p1)/2), 2) + math.Cos(p1)*math.Cos(p2)*math.Pow(math.Sin((home.Lon-lon.(float64))*rad/2), 2)
	if 6371000*2*math.Asin(math.Sqrt(a)) <= home.RadiusM {
		return "home"
	}

	return "not_home"
}

// Doors are the four doors, in the order the API names them.
var Doors = []string{"driver", "passenger", "rear_left", "rear_right"}

// Snapshot turns mcs status --json data into the published state: an allowlist,
// with no VIN, coordinates, nickname, device or account ids.
func Snapshot(data map[string]any, now time.Time, home *Home) map[string]any {
	dates := obj(data["reported_at"])
	bat, doors := obj(data["battery"]), obj(data["doors"])
	fuel, odo, tires := obj(data["fuel"]), obj(data["odometer"]), obj(data["tires"])
	cl, win := obj(data["climate"]), obj(data["windows"])
	locked := flag(doors["all_locked"])
	var opened any
	openKeys := []string{"driver_open", "passenger_open", "rear_left_open", "rear_right_open", "trunk_open", "hood_open"}
	all, anyOpen := true, false
	for _, k := range openKeys {
		b, ok := doors[k].(bool)
		all, anyOpen = all && ok, anyOpen || b
	}
	if all {
		opened = anyOpen
	}
	charging := flag(bat["charging"])
	var chargeTime any
	if isTrue(charging) {
		chargeTime = num(bat["charge_time_ac_minutes"], 0, 10000)
	}
	last := "unknown"
	if b, ok := locked.(bool); ok {
		last = map[bool]string{true: "locked", false: "unlocked"}[b]
	}
	s := map[string]any{
		"polled_at":           now.UTC().Format(isoMicro),
		"ev_reported_at":      reportTime(dates["ev"]),
		"doors_reported_at":   reportTime(dates["doors"]),
		"vehicle_reported_at": reportTime(dates["vehicle"]),
		// EV report
		"battery": num(bat["battery_level"], 0, 100), "charging": charging, "plugged_in": flag(bat["plugged_in"]),
		"range_km": num(bat["range_km"], 0, 3000), "charge_time_min": chargeTime,
		"battery_heater": flag(bat["heater_on"]), "battery_heater_auto": flag(bat["heater_auto"]),
		"climate_on": flag(cl["hvac_on"]), "front_defroster_on": flag(cl["front_defroster"]),
		"rear_defroster_on": flag(cl["rear_defroster"]), "cabin_temperature": num(cl["interior_temperature_c"], -60, 90),
		// vehicle report
		"fuel": num(fuel["fuel_level"], 0, 100), "odometer_km": num(odo["odometer_km"], 0, 3e6),
		"tire_front_left": num(tires["front_left_psi"], 0, 100), "tire_front_right": num(tires["front_right_psi"], 0, 100),
		"tire_rear_left": num(tires["rear_left_psi"], 0, 100), "tire_rear_right": num(tires["rear_right_psi"], 0, 100),
		// doors report
		"unlocked": not(locked), "doors_open": opened, "last_reported_locks": last,
		"trunk_open": flag(doors["trunk_open"]), "hood_open": flag(doors["hood_open"]),
		"fuel_lid_open": flag(doors["fuel_lid_open"]), "hazards": flag(data["hazards"]),
		"presence": Presence(obj(data["location"]), home),
	}
	var windows []any
	for _, d := range Doors {
		s[d+"_door_open"] = flag(doors[d+"_open"])
		s[d+"_door_unlocked"] = not(flag(doors[d+"_locked"]))
		var w any
		if p := num(win[d+"_position"], 0, 100); p != nil {
			w = p.(float64) > 0
		}
		s[d+"_window_open"] = w
		windows = append(windows, w)
	}
	s["windows_open"] = anyOf(windows)

	return s
}

// anyOf is nil when any value is unknown, else whether any is true.
func anyOf(vs []any) any {
	out := false
	for _, v := range vs {
		b, ok := v.(bool)
		if !ok {
			return nil
		}
		out = out || b
	}

	return out
}

// notSet is the health report's "no value" (e.g. no maintenance interval set).
const notSet = 65535

// Lamps maps the health report's lamp warnings to names.
var Lamps = []struct{ Key, Name string }{
	{"WngHeadLamp", "headlight"}, {"WngTailLamp", "tail light"}, {"WngBackLamp", "reversing light"},
	{"WngBreakLamp", "brake light"}, {"WngTurnLamp", "turn signal"}, {"WngSmallLamp", "parking light"},
	{"WngRearFogLamp", "rear fog light"},
}

// Health turns the raw health report into oil life, maintenance and warnings.
func Health(raw map[string]any) map[string]any {
	var info map[string]any
	if infos, ok := raw["remoteInfos"].([]any); ok && len(infos) > 0 {
		info = obj(infos[0])
	}
	oil, reg, tp := obj(info["OilMntInformation"]), obj(info["RegularMntInformation"]), obj(info["TPMSInformation"])
	var maint any
	if num(reg["MntSetDistKm"], 0, notSet-1) != nil {
		maint = num(reg["RemRegDistKm"], 0, notSet-1)
		if maint == nil {
			if m := num(reg["RemRegDistMile"], 0, notSet-1); m != nil {
				maint = math.Round(m.(float64) * 1.609344)
			}
		}
	}
	var lamps []string
	lampKnown := false
	for _, l := range Lamps {
		b := bit(info[l.Key])
		lampKnown = lampKnown || b != nil
		if isTrue(b) {
			lamps = append(lamps, l.Name)
		}
	}
	var bulbOut any
	if lampKnown {
		bulbOut = len(lamps) > 0
	}
	bulbs := strings.Join(lamps, ", ")
	if bulbs == "" {
		bulbs = "none"
	}
	var tireLow any
	if low := bit(info["WngTyrePressureLow"]); low != nil {
		t := isTrue(low)
		for _, k := range []string{"FLTyrePressWarn", "FRTyrePressWarn", "RLTyrePressWarn", "RRTyrePressWarn"} {
			t = t || isTrue(bit(tp[k]))
		}
		tireLow = t
	}
	var tpms any
	if a, b := bit(info["WngTpmsStatus"]), bit(tp["TPMSSystemFlt"]); a != nil || b != nil {
		tpms = isTrue(a) || isTrue(b)
	}

	return map[string]any{
		"health_reported_at": reportTime(info["OccurrenceDate"]),
		"oil_remaining_km":   num(oil["RemOilDistK"], 0, notSet-1), "maintenance_remaining_km": maint,
		"oil_low": bit(info["WngOilShortage"]), "oil_overfilled": bit(info["WngOilAmountExceed"]),
		"tire_pressure_low": tireLow, "tpms_fault": tpms, "bulb_out": bulbOut, "bulbs_out": bulbs,
	}
}

// HVACSetting reads the remote climate setting (temperature in °C, defrosters).
func HVACSetting(raw map[string]any) map[string]any {
	st := obj(raw["hvacSettings"])
	t := num(st["Temperature"], -50, 120)
	if f, ok := t.(float64); ok && st["TemperatureType"] == float64(2) {
		t = math.Round((f-32)*5/9*2) / 2
	}

	return map[string]any{"climate_temperature": t, "front_defroster": bit(st["FrontDefroster"]), "rear_defroster": bit(st["RearDefogger"])}
}

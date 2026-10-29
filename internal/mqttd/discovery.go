package mqttd

import (
	"maps"
	"strconv"
)

// Report time keys: which report a value comes from.
const (
	srcEV      = "ev_reported_at"
	srcVehicle = "vehicle_reported_at"
	srcDoors   = "doors_reported_at"
	srcHealth  = "health_reported_at"
)

type m = map[string]any

// spec is a read-only entity on the state topic. A fresh-only one goes
// unavailable once its report is older than --stale; the others keep the last
// reported value, with its time as the attribute reported_at.
type spec struct {
	domain, key, name, source string
	freshOnly                 bool
	extra                     m
}

func unit(class, u string) m {
	return m{"device_class": class, "unit_of_measurement": u, "state_class": "measurement"}
}

func with(a m, k string, v any) m {
	a[k] = v

	return a
}

func specs() []spec {
	p0 := "suggested_display_precision"
	out := []spec{
		{"sensor", "battery", "Battery", srcEV, false, unit("battery", "%")},
		{"binary_sensor", "charging", "Charging", srcEV, false, m{"device_class": "battery_charging"}},
		{"binary_sensor", "plugged_in", "Plugged in", srcEV, false, m{"device_class": "plug"}},
		{"sensor", "range_km", "Range (total)", srcEV, false, with(unit("distance", "km"), p0, 0)},
		{"sensor", "charge_time_min", "Charge time left", srcEV, false, m{p0: 0, "device_class": "duration", "unit_of_measurement": "min"}},
		{"binary_sensor", "battery_heater", "Battery heater", srcEV, false, m{"device_class": "heat"}},
		{"binary_sensor", "battery_heater_auto", "Battery heater auto", srcEV, false, m{"icon": "mdi:thermometer-auto"}},
		{"sensor", "cabin_temperature", "Cabin temperature", srcEV, false, unit("temperature", "°C")},
		{"binary_sensor", "climate_on", "Climate running", srcEV, true, m{"icon": "mdi:fan"}},
		{"binary_sensor", "front_defroster_on", "Front defroster running", srcEV, true, m{"icon": "mdi:car-defrost-front"}},
		{"binary_sensor", "rear_defroster_on", "Rear defroster running", srcEV, true, m{"icon": "mdi:car-defrost-rear"}},
		{"sensor", "fuel", "Fuel", srcVehicle, false, m{"icon": "mdi:gas-station", "unit_of_measurement": "%", "state_class": "measurement"}},
		{"sensor", "odometer_km", "Odometer", srcVehicle, false, m{p0: 0, "device_class": "distance", "unit_of_measurement": "km", "state_class": "total_increasing"}},
	}
	for _, t := range []struct{ key, name string }{{"front_left", "Tire front left"}, {"front_right", "Tire front right"}, {"rear_left", "Tire rear left"}, {"rear_right", "Tire rear right"}} {
		out = append(out, spec{"sensor", "tire_" + t.key, t.name, srcVehicle, false, with(unit("pressure", "psi"), p0, 0)})
	}
	out = append(out,
		spec{"binary_sensor", "unlocked", "Doors unlocked", srcDoors, true, m{"device_class": "lock"}},
		spec{"binary_sensor", "doors_open", "Door open", srcDoors, true, m{"device_class": "door"}},
		spec{"sensor", "last_reported_locks", "Last reported locks", "", false, m{"icon": "mdi:lock-clock"}},
		spec{"binary_sensor", "trunk_open", "Trunk open", srcDoors, true, m{"device_class": "opening", "icon": "mdi:car-back"}},
		spec{"binary_sensor", "hood_open", "Hood open", srcDoors, true, m{"device_class": "opening"}},
		spec{"binary_sensor", "fuel_lid_open", "Fuel lid open", srcDoors, true, m{"device_class": "opening", "icon": "mdi:gas-station"}},
		spec{"binary_sensor", "windows_open", "Window open", srcDoors, true, m{"device_class": "window"}},
		spec{"binary_sensor", "hazards", "Hazard lights", srcDoors, true, m{"icon": "mdi:car-light-alert"}},
		spec{"sensor", "oil_remaining_km", "Oil change in", srcHealth, false, m{p0: 0, "device_class": "distance", "unit_of_measurement": "km", "icon": "mdi:oil"}},
		spec{"sensor", "maintenance_remaining_km", "Maintenance in", srcHealth, false, m{p0: 0, "device_class": "distance", "unit_of_measurement": "km", "icon": "mdi:car-wrench"}},
		spec{"binary_sensor", "oil_low", "Oil low", srcHealth, false, m{"device_class": "problem"}},
		spec{"binary_sensor", "oil_overfilled", "Oil overfilled", srcHealth, false, m{"device_class": "problem"}},
		spec{"binary_sensor", "tire_pressure_low", "Tire pressure low", srcHealth, false, m{"device_class": "problem"}},
		spec{"binary_sensor", "tpms_fault", "Tire pressure system fault", srcHealth, false, m{"device_class": "problem"}},
		spec{"binary_sensor", "bulb_out", "Bulb out", srcHealth, false, m{"device_class": "problem"}},
		spec{"sensor", srcEV, "EV report time", "", false, m{"device_class": "timestamp"}},
		spec{"sensor", srcDoors, "Door report time", "", false, m{"device_class": "timestamp"}},
		spec{"sensor", srcVehicle, "Vehicle report time", "", false, m{"device_class": "timestamp"}},
		spec{"sensor", srcHealth, "Health report time", "", false, m{"device_class": "timestamp", "entity_category": "diagnostic"}},
		spec{"sensor", "polled_at", "Last successful poll", "", false, m{"device_class": "timestamp", "entity_category": "diagnostic"}},
	)
	for _, d := range Doors() {
		n := map[string]string{"driver": "Driver", "passenger": "Passenger", "rear_left": "Rear left", "rear_right": "Rear right"}[d]
		out = append(out,
			spec{"binary_sensor", d + "_door_open", n + " door open", srcDoors, true, m{"device_class": "door"}},
			spec{"binary_sensor", d + "_door_unlocked", n + " door unlocked", srcDoors, true, m{"device_class": "lock"}},
			spec{"binary_sensor", d + "_window_open", n + " window open", srcDoors, true, m{"device_class": "window"}})
	}

	return out
}

func freshness(field string, stale int) string {
	return "{{ 'online' if value_json." + field + " is not none and -300 <= as_timestamp(now()) - as_timestamp(value_json." + field +
		", 0) < " + strconv.Itoa(stale) + " else 'offline' }}"
}

func boolTemplate(key string) string {
	return "{{ 'unknown' if value_json." + key + " is none else ('ON' if value_json." + key + " else 'OFF') }}"
}

// discovery collects one car's discovery configs, by topic.
type discovery struct {
	c               *Config
	id, node, state string
	device          m
	avail           []m // the daemon's status
	stale, expiry   int // seconds
	out             map[string]m
}

func newDiscovery(c *Config, model string) *discovery {
	id := c.ID()
	node := "mcs_" + id
	name := c.DeviceName
	if name == "" {
		name = "Mazda " + model
	}

	return &discovery{
		c: c, id: id, node: node, state: c.Prefix + "/state",
		device: m{"identifiers": []string{node}, "name": name, "manufacturer": "Mazda", "model": model},
		avail:  []m{{"topic": c.Prefix + "/status"}},
		stale:  int(c.Stale.Seconds()),
		expiry: int(3 * c.Poll.Seconds()),
		out:    map[string]m{},
	}
}

// Discovery returns every discovery config topic with its payload. A nil
// payload clears that entity (a control that is not enabled).
func Discovery(c *Config, model string) map[string]m {
	d := newDiscovery(c, model)
	for _, s := range specs() {
		d.add(s.domain, s.key, d.specConfig(s))
	}
	d.add("device_tracker", "presence", m{
		"name": "Presence", "default_entity_id": "device_tracker." + d.id, "icon": "mdi:car", "state_topic": d.state,
		"value_template": "{{ value_json.presence if value_json.presence is not none else 'unknown' }}",
		"payload_home":   "home", "payload_not_home": "not_home", "source_type": "gps",
		"availability_mode": "all", "availability": d.freshAvailability(srcDoors),
	})
	d.add("binary_sensor", "connection", m{
		"name": "Cloud connection", "default_entity_id": "binary_sensor." + d.id + "_cloud_connection", "device_class": "connectivity",
		"entity_category": "diagnostic", "state_topic": c.Prefix + "/status", "payload_on": "online", "payload_off": "offline", "expire_after": d.expiry,
	})
	d.addControls()

	return d.out
}

// add sets a topic's config, with the ids, device and origin (nil clears it).
func (d *discovery) add(domain, key string, cfg m) {
	if cfg != nil {
		cfg["unique_id"] = d.node + "_" + key
		cfg["device"] = d.device
		cfg["origin"] = m{"name": "mcs"}
	}
	d.out[d.c.Discovery+"/"+domain+"/"+d.node+"/"+key+"/config"] = cfg
}

// freshAvailability is available while the daemon is online and source's
// report is fresh.
func (d *discovery) freshAvailability(source string) []m {
	return append(append([]m{}, d.avail...), m{"topic": d.state, "value_template": freshness(source, d.stale)})
}

// specConfig is a read-only entity's config.
func (d *discovery) specConfig(s spec) m {
	cfg := m{"name": s.name, "default_entity_id": s.domain + "." + d.id + "_" + s.key, "state_topic": d.state}
	maps.Copy(cfg, s.extra)
	if s.domain == "binary_sensor" {
		cfg["payload_on"], cfg["payload_off"], cfg["value_template"] = payloadOn, payloadOff, boolTemplate(s.key)
	} else {
		cfg["value_template"] = "{{ value_json." + s.key + " if value_json." + s.key + " is not none else 'None' }}"
	}
	switch {
	case s.key == "polled_at":
		cfg["json_attributes_topic"] = d.state // the allowlisted snapshot
	case s.key == "bulb_out":
		cfg["json_attributes_topic"] = d.state
		cfg["json_attributes_template"] = "{{ {'reported_at': value_json.health_reported_at, 'bulbs': value_json.bulbs_out} | tojson }}"
	case s.source != "":
		cfg["json_attributes_topic"] = d.state
		cfg["json_attributes_template"] = "{{ {'reported_at': value_json." + s.source + "} | tojson }}"
	}
	if s.freshOnly {
		cfg["availability_mode"], cfg["expire_after"] = "all", d.expiry
		cfg["availability"] = d.freshAvailability(s.source)
	}

	return cfg
}

// control adds a control's entity if its group is enabled, else clears it.
func (d *discovery) control(group, domain, key string, cfg m) {
	if !d.c.Has(group) {
		cfg = nil
	}
	d.add(domain, key, cfg)
}

// addControls adds the entities of the enabled groups and clears the others.
func (d *discovery) addControls() {
	c, id := d.c, d.id
	d.control(GroupRefresh, "button", "refresh", m{
		"name": "Ask car for a fresh report", "default_entity_id": "button." + id + "_fresh_report", "icon": "mdi:car-wireless",
		"command_topic": c.Prefix + "/refresh/set", "payload_press": payloadPress, "availability": d.avail,
	})
	d.control(GroupRefresh, "sensor", "refresh_status", m{
		"name": "Fresh report request", "default_entity_id": "sensor." + id + "_fresh_report_request", "icon": "mdi:car-clock",
		"state_topic": c.Prefix + "/refresh_status",
	})
	for _, b := range Buttons() {
		d.control(b.Group, "button", b.Name, m{
			"name": b.Label, "default_entity_id": "button." + id + "_" + b.Name, "icon": b.Icon,
			"command_topic": c.Prefix + "/cmd/" + b.Name + "/set", "payload_press": payloadPress, "availability": d.avail,
		})
	}
	d.control(GroupClimate, "number", SettingTemperature, m{
		"name": "Climate temperature", "default_entity_id": "number." + id + "_climate_temperature", "icon": "mdi:thermometer",
		"device_class": "temperature", "unit_of_measurement": "°C", "min": TempMin, "max": TempMax, "step": 0.5, "mode": "box",
		"command_topic": c.Prefix + "/cmd/climate_temperature/set", "state_topic": d.state,
		"value_template": "{{ value_json.climate_temperature if value_json.climate_temperature is not none else 'None' }}", "availability": d.avail,
	})
	for _, s := range []struct{ key, name, icon string }{{SettingFront, "Front defroster", "mdi:car-defrost-front"}, {SettingRear, "Rear defroster", "mdi:car-defrost-rear"}} {
		d.control(GroupClimate, "switch", s.key, m{
			"name": s.name, "default_entity_id": "switch." + id + "_" + s.key, "icon": s.icon,
			"command_topic": c.Prefix + "/cmd/" + s.key + "/set", "payload_on": payloadOn, "payload_off": payloadOff,
			"state_topic": d.state, "value_template": boolTemplate(s.key), "availability": d.avail,
		})
	}
	status := m{
		"name": "Last command", "default_entity_id": "sensor." + id + "_last_command", "icon": "mdi:remote",
		"state_topic": c.Prefix + "/command_status",
	}
	if !c.HasCommands() {
		status = nil
	}
	d.add("sensor", "command_status", status)
}

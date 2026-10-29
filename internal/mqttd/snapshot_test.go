package mqttd

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const statusJSON = `{"reported_at": {"doors": "20260101000000", "ev": "20260201000000", "vehicle": "20260101120000"},
 "vehicle": {"vin": "private", "nickname": "private"},
 "location": {"latitude": 45.5, "longitude": -73.6, "maps_url": "private"},
 "battery": {"battery_level": 100, "charging": false, "plugged_in": true, "range_km": 290, "heater_on": false, "heater_auto": false},
 "climate": {"hvac_on": false, "front_defroster": false, "rear_defroster": false, "interior_temperature_c": 16.5, "target_temperature_c": 0},
 "fuel": {"fuel_level": 46, "range_km": 290}, "odometer": {"odometer_km": 14655.4},
 "tires": {"front_left_psi": 33, "front_right_psi": 33, "rear_left_psi": 36, "rear_right_psi": 36},
 "doors": {"all_locked": false, "driver_open": false, "passenger_open": false, "rear_left_open": false, "rear_right_open": true,
           "trunk_open": false, "hood_open": false, "fuel_lid_open": false,
           "driver_locked": true, "passenger_locked": false, "rear_left_locked": true, "rear_right_locked": true},
 "windows": {"driver_position": 0, "passenger_position": 0, "rear_left_position": 20, "rear_right_position": 0},
 "hazards": false}`

const healthJSON = `{"resultCode": "200S00", "remoteInfos": [{
 "OccurrenceDate": "20261028205949", "OilMntInformation": {"RemOilDistK": 7000, "RemOilDistMile": 4600},
 "RegularMntInformation": {"MntSetDistKm": 65535, "MntSetDistMile": 40721, "RemRegDistMile": 40960},
 "TPMSInformation": {"FLTyrePressWarn": 0, "FRTyrePressWarn": 0, "RLTyrePressWarn": 0, "RRTyrePressWarn": 0, "TPMSSystemFlt": 0},
 "WngBackLamp": 0, "WngBreakLamp": 0, "WngHeadLamp": 0, "WngOilAmountExceed": 0, "WngOilShortage": 0,
 "WngRearFogLamp": 0, "WngSmallLamp": 0, "WngTailLamp": 0, "WngTpmsStatus": 0, "WngTurnLamp": 0, "WngTyrePressureLow": 0}]}`

func decode(t *testing.T, s string) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(s), &out))

	return out
}

func TestSnapshotAllowlist(t *testing.T) {
	t.Parallel()
	d := Snapshot(decode(t, statusJSON), time.Unix(1793210000, 0), &Home{45.5, -73.6, 200})
	text, err := json.Marshal(d)
	require.NoError(t, err)
	for _, secret := range []string{"private", "45.5", "-73.6", "latitude", "vin"} {
		assert.NotContains(t, string(text), secret)
	}
	assert.Equal(t, "home", d["presence"])
	assert.Equal(t, "unlocked", d["last_reported_locks"])
	assert.Equal(t, true, d["doors_open"])
	assert.Equal(t, true, d["rear_right_door_open"])
	assert.Equal(t, true, d["passenger_door_unlocked"])
	assert.Equal(t, false, d["driver_door_unlocked"])
	assert.Equal(t, true, d["rear_left_window_open"])
	assert.Equal(t, true, d["windows_open"])
	assert.Equal(t, []any{46.0, 14655.4, 290.0, 33.0}, []any{d["fuel"], d["odometer_km"], d["range_km"], d["tire_front_left"]})
	assert.Equal(t, 16.5, d["cabin_temperature"])
	assert.Nil(t, d["charge_time_min"]) // only while charging
	assert.Equal(t, "2026-02-01T00:00:00+00:00", d["ev_reported_at"])
	assert.Equal(t, "2026-10-28T17:53:20.000000+00:00", d["polled_at"])
}

func TestPresence(t *testing.T) {
	t.Parallel()
	home := &Home{45.5, -73.6, 200}
	assert.Equal(t, "not_home", Presence(map[string]any{"latitude": 45.6, "longitude": -73.6}, home))
	assert.Nil(t, Presence(map[string]any{"latitude": 0.0, "longitude": 0.0}, home))
	assert.Nil(t, Presence(map[string]any{"latitude": 45.5, "longitude": -73.6}, nil))
}

func TestMissingIsUnknown(t *testing.T) {
	t.Parallel()
	d := Snapshot(map[string]any{}, time.Unix(0, 0), nil)
	for k, v := range d {
		if k != "polled_at" && k != "last_reported_locks" {
			assert.Nil(t, v, k)
		}
	}
	assert.Equal(t, "unknown", d["last_reported_locks"])
	assert.Nil(t, Snapshot(map[string]any{"battery": map[string]any{"battery_level": 101.0, "charging": "false"}}, time.Unix(0, 0), nil)["battery"])
	assert.Nil(t, reportTime("not a date"))
	assert.Nil(t, reportTime("20269999999999"))
}

func TestHealth(t *testing.T) {
	t.Parallel()
	h := Health(decode(t, healthJSON))
	assert.Equal(t, 7000.0, h["oil_remaining_km"])
	assert.Nil(t, h["maintenance_remaining_km"]) // 65535: no interval set
	for _, k := range []string{"oil_low", "oil_overfilled", "tire_pressure_low", "tpms_fault", "bulb_out"} {
		assert.Equal(t, false, h[k], k)
	}
	assert.Equal(t, "none", h["bulbs_out"])
	assert.Equal(t, "2026-10-28T20:59:49+00:00", h["health_reported_at"])

	bad := strings.NewReplacer(`"WngHeadLamp": 0`, `"WngHeadLamp": 1`, `"WngTurnLamp": 0`, `"WngTurnLamp": 1`,
		`"RLTyrePressWarn": 0`, `"RLTyrePressWarn": 1`,
		`{"MntSetDistKm": 65535, "MntSetDistMile": 40721, "RemRegDistMile": 40960}`, `{"MntSetDistKm": 10000, "RemRegDistKm": 1234}`).Replace(healthJSON)
	h = Health(decode(t, bad))
	assert.Equal(t, true, h["bulb_out"])
	assert.Equal(t, "headlight, turn signal", h["bulbs_out"])
	assert.Equal(t, true, h["tire_pressure_low"])
	assert.Equal(t, 1234.0, h["maintenance_remaining_km"])
	empty := Health(nil)
	assert.Nil(t, empty["oil_remaining_km"])
	assert.Nil(t, empty["bulb_out"])
}

func TestHVACSetting(t *testing.T) {
	t.Parallel()
	assert.Equal(t, map[string]any{"climate_temperature": 17.0, "front_defroster": false, "rear_defroster": true},
		HVACSetting(decode(t, `{"hvacSettings": {"Temperature": 17, "TemperatureType": 1, "FrontDefroster": 0, "RearDefogger": 1}}`)))
	assert.Equal(t, 21.0, HVACSetting(decode(t, `{"hvacSettings": {"Temperature": 70, "TemperatureType": 2}}`))["climate_temperature"])
	assert.Nil(t, HVACSetting(nil)["climate_temperature"])
}

func testConfig(controls ...string) *Config {
	c := Defaults()
	c.Prefix, c.Broker, c.Controls = "mazda/cx90", "tcp://127.0.0.1:1883", controls
	if err := c.Validate(); err != nil {
		panic(err)
	}

	return &c
}

func TestConfig(t *testing.T) {
	t.Parallel()
	c := Defaults()
	require.ErrorIs(t, c.Validate(), ErrConfig) // no prefix
	require.ErrorContains(t, c.Validate(), "prefix is required")
	c.Prefix = "mazda/cx90"
	require.ErrorContains(t, c.Validate(), "broker is required")
	c.DryRun = true
	require.NoError(t, c.Validate())
	for _, p := range []string{"mazda/#", "/mazda", "mazda//cx90", "a/+/b"} {
		c.Prefix = p
		require.ErrorIs(t, c.Validate(), ErrConfig, p)
	}
	c.Prefix, c.Controls = "mazda/cx90", []string{"lock", "selfdestruct"}
	require.ErrorContains(t, c.Validate(), "selfdestruct")
	c.Controls = []string{"all"}
	require.NoError(t, c.Validate())
	assert.Equal(t, Groups, c.Controls)
	c.Poll = time.Minute
	require.ErrorContains(t, c.Validate(), "at least 5m")
	assert.Equal(t, "cx90", testConfig().ID())
}

func TestDiscovery(t *testing.T) {
	t.Parallel()
	all := Discovery(testConfig("all"), "CX-90")
	commands := map[string]bool{}
	for topic, c := range all {
		assert.Contains(t, topic, "/mcs_cx90/")
		require.NotNil(t, c, topic)
		for k, v := range c {
			if strings.HasPrefix(k, "command") {
				s, ok := v.(string)
				require.True(t, ok)
				commands[s] = true
				assert.True(t, strings.HasPrefix(s, "mazda/cx90/cmd/") || s == "mazda/cx90/refresh/set", s)
			}
		}
	}
	assert.Len(t, commands, 14) // 10 buttons, refresh, temperature, 2 defrosters
	for _, key := range []string{"unlocked", "doors_open", "windows_open", "hazards", "climate_on", "driver_door_unlocked"} {
		c := all["homeassistant/binary_sensor/mcs_cx90/"+key+"/config"]
		assert.Equal(t, "all", c["availability_mode"], key)
		av, ok := c["availability"].([]m)
		require.True(t, ok)
		assert.Contains(t, av[1]["value_template"], "1800")
	}
	bat := all["homeassistant/sensor/mcs_cx90/battery/config"]
	assert.NotContains(t, bat, "availability")
	assert.Contains(t, bat["json_attributes_template"], "ev_reported_at")
	assert.Equal(t, "mcs_cx90_battery", bat["unique_id"])
	assert.Equal(t, "sensor.cx90_battery", bat["default_entity_id"])

	// Read-only: the same topics, every control's config cleared.
	ro := Discovery(testConfig(), "CX-90")
	assert.Len(t, ro, len(all))
	for topic, c := range ro {
		if c != nil {
			for k := range c {
				assert.False(t, strings.HasPrefix(k, "command"), topic)
			}
		}
	}
	assert.Nil(t, ro["homeassistant/button/mcs_cx90/unlock/config"])
	assert.Nil(t, ro["homeassistant/sensor/mcs_cx90/command_status/config"])
	lockOnly := Discovery(testConfig("lock"), "CX-90")
	assert.NotNil(t, lockOnly["homeassistant/button/mcs_cx90/lock/config"])
	assert.Nil(t, lockOnly["homeassistant/button/mcs_cx90/unlock/config"])
}

func TestParseCommand(t *testing.T) {
	t.Parallel()
	c := testConfig("all")
	cmd, ok := ParseCommand(c, "mazda/cx90/cmd/lock/set", []byte("PRESS"))
	assert.True(t, ok)
	assert.Equal(t, Command{Name: "lock"}, cmd)
	for _, bad := range [][2]string{{"mazda/cx90/cmd/lock/set", "press"}, {"mazda/cx90/cmd/selfdestruct/set", "PRESS"},
		{"mazda/cx90/cmd/climate_temperature/set", "40"}, {"mazda/cx90/cmd/climate_temperature/set", "nan"},
		{"mazda/cx90/refresh/set", "PRESS"}, {"other/cx90/cmd/lock/set", "PRESS"}} {
		_, ok := ParseCommand(c, bad[0], []byte(bad[1]))
		assert.False(t, ok, bad)
	}
	cmd, _ = ParseCommand(c, "mazda/cx90/cmd/climate_temperature/set", []byte("21.3"))
	assert.Equal(t, Command{Name: SettingTemperature, Value: 21.5}, cmd)
	cmd, _ = ParseCommand(c, "mazda/cx90/cmd/rear_defroster/set", []byte("ON"))
	assert.Equal(t, Command{Name: SettingRear, Value: true}, cmd)
	_, ok = ParseCommand(testConfig("lock"), "mazda/cx90/cmd/unlock/set", []byte("PRESS"))
	assert.False(t, ok, "a group that is not enabled")
}

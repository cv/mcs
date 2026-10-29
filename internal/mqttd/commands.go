package mqttd

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Button is a command without a value.
type Button struct {
	Name, Group, Label, Icon string
}

// Button commands, as named in <prefix>/cmd/<name>/set.
const (
	CmdLock        = "lock"
	CmdUnlock      = "unlock"
	CmdEngineStart = "engine_start"
	CmdEngineStop  = "engine_stop"
	CmdChargeStart = "charge_start"
	CmdChargeStop  = "charge_stop"
	CmdClimateOn   = "climate_on"
	CmdClimateOff  = "climate_off"
	CmdLightsOn    = "lights_on"
	CmdLightsOff   = "lights_off"
)

// Payloads a button press and a switch send.
const (
	payloadPress = "PRESS"
	payloadOn    = "ON"
	payloadOff   = "OFF"
)

// Buttons lists the commands HA gets a button for.
func Buttons() []Button {
	return []Button{
		{CmdLock, GroupLock, "Lock", "mdi:car-door-lock"},
		{CmdUnlock, GroupUnlock, "Unlock", "mdi:car-door"},
		{CmdEngineStart, GroupEngine, "Start engine", "mdi:engine"},
		{CmdEngineStop, GroupEngine, "Stop engine", "mdi:engine-off"},
		{CmdChargeStart, GroupCharge, "Start charging", "mdi:ev-station"},
		{CmdChargeStop, GroupCharge, "Stop charging", "mdi:ev-plug-type1"},
		{CmdClimateOn, GroupClimate, "Climate on", "mdi:fan"},
		{CmdClimateOff, GroupClimate, "Climate off", "mdi:fan-off"},
		{CmdLightsOn, GroupLights, "Hazard lights on", "mdi:car-light-alert"},
		{CmdLightsOff, GroupLights, "Hazard lights off", "mdi:car-light-dimmed"},
	}
}

// The climate setting: a temperature and two defrosters, always sent together.
const (
	SettingTemperature = "climate_temperature"
	SettingFront       = "front_defroster"
	SettingRear        = "rear_defroster"
	TempMin, TempMax   = 16.0, 28.0
)

// HasCommands reports whether any command besides the fresh report is enabled.
func (c *Config) HasCommands() bool {
	for _, g := range c.Controls {
		if g != GroupRefresh {
			return true
		}
	}

	return false
}

// Command is a parsed command message.
type Command struct {
	Name  string
	Value any // float64 for the temperature, bool for a defroster, nil for a button
}

// ParseCommand reads <prefix>/cmd/<name>/set; anything else, or a command that
// is not enabled, gives false.
func ParseCommand(c *Config, topic string, payload []byte) (Command, bool) {
	name, ok := commandName(c.Prefix, topic)
	if !ok {
		return Command{}, false
	}
	text := strings.TrimSpace(string(payload))
	if b, ok := findButton(name); ok {
		return Command{Name: name}, c.Has(b.Group) && text == payloadPress
	}
	if !c.Has(GroupClimate) {
		return Command{}, false
	}

	return parseSetting(name, text)
}

// commandName is <name> in <prefix>/cmd/<name>/set.
func commandName(prefix, topic string) (string, bool) {
	name, ok := strings.CutPrefix(topic, prefix+"/cmd/")
	if !ok {
		return "", false
	}

	return strings.CutSuffix(name, "/set")
}

func findButton(name string) (Button, bool) {
	buttons := Buttons()
	if i := slices.IndexFunc(buttons, func(b Button) bool { return b.Name == name }); i >= 0 {
		return buttons[i], true
	}

	return Button{}, false
}

// parseSetting reads a climate setting command: a temperature (rounded to
// 0.5 °C) or a defroster's ON / OFF.
func parseSetting(name, text string) (Command, bool) {
	switch name {
	case SettingTemperature:
		t, err := strconv.ParseFloat(text, 64)
		if err != nil || math.IsNaN(t) || t < TempMin || t > TempMax {
			return Command{}, false
		}

		return Command{Name: name, Value: math.Round(t*2) / 2}, true
	case SettingFront, SettingRear:
		if text != payloadOn && text != payloadOff {
			return Command{}, false
		}

		return Command{Name: name, Value: text == payloadOn}, true
	}

	return Command{}, false
}

// IsSetting reports whether a command changes the climate setting.
func IsSetting(name string) bool {
	return name == SettingTemperature || name == SettingFront || name == SettingRear
}

// Limits remembers when each command last went out, across restarts.
type Limits struct {
	path string
	last map[string]time.Time
}

// LoadLimits reads the file (a missing or bad one starts empty); an empty path
// keeps them in memory only.
func LoadLimits(path string) *Limits {
	l := &Limits{path: path, last: map[string]time.Time{}}
	if path == "" {
		return l
	}
	if b, err := os.ReadFile(path); err == nil {
		var raw map[string]float64
		if json.Unmarshal(b, &raw) == nil {
			for k, v := range raw {
				l.last[k] = time.Unix(0, int64(v*1e9))
			}
		}
	}

	return l
}

// Wait is how long until key may go out again (0: now).
func (l *Limits) Wait(key string, now time.Time, every time.Duration) time.Duration {
	t, ok := l.last[key]
	if !ok || now.Sub(t) >= every {
		return 0
	}

	return every - now.Sub(t)
}

// Mark records key as sent now and saves the file (atomically).
func (l *Limits) Mark(key string, now time.Time) {
	l.last[key] = now
	if l.path == "" {
		return
	}
	raw := map[string]float64{}
	for k, v := range l.last {
		raw[k] = float64(v.UnixNano()) / 1e9
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return // kept in memory only, as when the file cannot be written
	}
	_ = os.MkdirAll(filepath.Dir(l.path), 0o700)
	tmp := l.path + ".new"
	if os.WriteFile(tmp, b, 0o600) == nil {
		_ = os.Rename(tmp, l.path)
	}
}

// minutes rounds a wait up to whole minutes, for the status line.
func minutes(d time.Duration) int { return int(math.Ceil(d.Minutes())) }

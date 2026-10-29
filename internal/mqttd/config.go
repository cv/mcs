// Package mqttd publishes a vehicle's status to MQTT, with Home Assistant
// discovery, and carries out the commands it is allowed to (mcs mqtt).
package mqttd

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

// ErrConfig marks a configuration problem (exit status 2: not worth restarting).
var ErrConfig = errors.New("invalid mqtt configuration")

// ErrAuthLost marks a lost Mazda session (exit status 3: renew it by hand; a
// restart would only compete with the phone for its session).
var ErrAuthLost = errors.New("the Mazda session is lost; renew it with mcs status on a computer and copy the cache")

// Control groups, as named in --controls.
const (
	GroupLock    = "lock"
	GroupUnlock  = "unlock"
	GroupEngine  = "engine"
	GroupCharge  = "charge"
	GroupClimate = "climate"
	GroupLights  = "lights"
	GroupRefresh = "refresh"
)

// allControls enables every control group in --controls.
const allControls = "all"

// Groups lists every control group.
func Groups() []string {
	return []string{GroupLock, GroupUnlock, GroupEngine, GroupCharge, GroupClimate, GroupLights, GroupRefresh}
}

// Home is the place presence is measured against.
type Home struct {
	Lat, Lon, RadiusM float64
}

// Config is everything mcs mqtt needs besides the Mazda login.
type Config struct {
	Broker     string // tcp://host:1883, ssl://host:8883 (also mqtt://, mqtts://, tls://)
	Username   string
	Password   string
	ClientID   string
	Prefix     string // topic root, e.g. mazda/cx90; required
	Discovery  string // Home Assistant discovery prefix, or "off"
	DeviceName string

	Poll, Health, Stale                              time.Duration
	CommandInterval, EngineInterval, RefreshInterval time.Duration

	Controls []string // groups; empty: read-only
	Home     *Home
	StateDir string
	Once     bool
	DryRun   bool
}

// Defaults returns a Config with every optional setting filled in. Prefix and
// Broker have no default.
func Defaults() Config {
	return Config{
		Discovery: "homeassistant",
		Poll:      5 * time.Minute, Health: 30 * time.Minute, Stale: 30 * time.Minute,
		CommandInterval: time.Minute, EngineInterval: 10 * time.Minute, RefreshInterval: 30 * time.Minute,
	}
}

var level = regexp.MustCompile(`^[^/#+\x00]+$`)

// Validate checks the configuration (and normalizes Controls); every error
// wraps ErrConfig.
func (c *Config) Validate() error {
	for _, check := range []func() error{
		c.validatePrefix, c.validateBroker, c.validateDiscovery, c.validateIntervals, c.normalizeControls, c.validateHome,
	} {
		if err := check(); err != nil {
			return err
		}
	}

	return nil
}

// NormalizeControls expands "all", drops repeats and checks the names.
func NormalizeControls(names []string) ([]string, error) {
	var out []string
	for _, n := range names {
		n = strings.ToLower(strings.TrimSpace(n))
		switch {
		case n == "":
		case n == allControls:
			return Groups(), nil
		case slices.Contains(Groups(), n):
			if !slices.Contains(out, n) {
				out = append(out, n)
			}
		default:
			return nil, fmt.Errorf("unknown control %q (one of %s, all)", n, strings.Join(Groups(), ", "))
		}
	}

	return out, nil
}

// Has reports whether a control group is enabled.
func (c *Config) Has(group string) bool { return slices.Contains(c.Controls, group) }

// ID is the last level of the prefix, made safe for entity ids (cx90 for mazda/cx90).
func (c *Config) ID() string {
	last := c.Prefix[strings.LastIndex(c.Prefix, "/")+1:]
	id := strings.Trim(regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(strings.ToLower(last), "_"), "_")
	if id == "" {
		return "car"
	}

	return id
}

func (c *Config) validatePrefix() error {
	if c.Prefix == "" {
		return invalid("prefix is required (mqtt.prefix in the config, MCS_MQTT_PREFIX or --prefix), e.g. mazda/cx90")
	}
	for l := range strings.SplitSeq(c.Prefix, "/") {
		if !level.MatchString(l) {
			return invalid("prefix %q: every level must be non-empty, without / + #", c.Prefix)
		}
	}

	return nil
}

// validateBroker requires a broker (a dry run needs none) with a known scheme.
func (c *Config) validateBroker() error {
	if c.Broker == "" && !c.DryRun {
		return invalid("broker is required (mqtt.broker, MCS_MQTT_BROKER or --broker), e.g. tcp://127.0.0.1:1883")
	}
	if c.Broker != "" {
		scheme, _, ok := strings.Cut(c.Broker, "://")
		if !ok || !slices.Contains([]string{"tcp", "mqtt", "ssl", "tls", "mqtts"}, scheme) {
			return invalid("broker %q: use tcp://, mqtt://, ssl://, tls:// or mqtts://", c.Broker)
		}
	}

	return nil
}

func (c *Config) validateDiscovery() error {
	if c.Discovery == "" || (c.Discovery != "off" && !level.MatchString(c.Discovery)) {
		return invalid("discovery %q: one topic level, or off", c.Discovery)
	}

	return nil
}

// validateIntervals enforces the minimum intervals (they protect the account).
func (c *Config) validateIntervals() error {
	switch {
	case c.Poll < 5*time.Minute:
		return invalid("poll %s: at least 5m", c.Poll)
	case c.Health < c.Poll:
		return invalid("health %s: at least the poll interval", c.Health)
	case c.Stale < 2*c.Poll:
		return invalid("stale %s: at least twice the poll interval", c.Stale)
	case c.CommandInterval < 30*time.Second:
		return invalid("command interval %s: at least 30s", c.CommandInterval)
	case c.EngineInterval < c.CommandInterval:
		return invalid("engine interval %s: at least the command interval", c.EngineInterval)
	case c.RefreshInterval < 10*time.Minute:
		return invalid("refresh interval %s: at least 10m", c.RefreshInterval)
	}

	return nil
}

// normalizeControls replaces Controls with NormalizeControls' result.
func (c *Config) normalizeControls() error {
	groups, err := NormalizeControls(c.Controls)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrConfig, err)
	}
	c.Controls = groups

	return nil
}

func (c *Config) validateHome() error {
	if c.Home != nil && (c.Home.Lat < -90 || c.Home.Lat > 90 || c.Home.Lon < -180 || c.Home.Lon > 180 || c.Home.RadiusM <= 0) {
		return invalid("home: latitude -90..90, longitude -180..180, radius above 0")
	}

	return nil
}

// invalid is a configuration error (it wraps ErrConfig).
func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrConfig, fmt.Sprintf(format, args...))
}

// ParseHome reads "LAT,LON" (the radius comes separately, in meters).
func ParseHome(s string, radiusM float64) (*Home, error) {
	latS, lonS, ok := strings.Cut(s, ",")
	if !ok {
		return nil, fmt.Errorf("%w: home %q: use LAT,LON", ErrConfig, s)
	}
	lat, err1 := strconv.ParseFloat(strings.TrimSpace(latS), 64)
	lon, err2 := strconv.ParseFloat(strings.TrimSpace(lonS), 64)
	if err1 != nil || err2 != nil {
		return nil, fmt.Errorf("%w: home %q: use LAT,LON", ErrConfig, s)
	}

	return &Home{Lat: lat, Lon: lon, RadiusM: radiusM}, nil
}

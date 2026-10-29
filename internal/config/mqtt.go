package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// MQTT is the [mqtt] section of config.toml (env MCS_MQTT_<KEY> overrides,
// e.g. MCS_MQTT_PREFIX). Unset values are zero; mcs mqtt fills its defaults.
type MQTT struct {
	Broker, Username, Password, PasswordFile, ClientID string
	Prefix, Discovery, DeviceName, Home, StateDir      string
	HomeRadiusM                                        float64
	Poll, Health, Stale                                time.Duration
	CommandInterval, EngineInterval, RefreshInterval   time.Duration
	Controls                                           []string
}

// LoadMQTT reads the [mqtt] section from the config file (the --config one, or
// ~/.config/mcs/config.toml) and the environment. The password comes from
// MCS_MQTT_PASSWORD or password_file, never from the config itself.
func LoadMQTT(configPath string) (*MQTT, error) {
	v, err := readMQTTConfig(configPath)
	if err != nil {
		return nil, err
	}
	get := func(k string) string { return v.GetString("mqtt." + k) }
	m := &MQTT{
		Broker: get("broker"), Username: get("username"), Password: os.Getenv("MCS_MQTT_PASSWORD"),
		PasswordFile: get("password_file"), ClientID: get("client_id"), Prefix: get("prefix"),
		Discovery: get("discovery"), DeviceName: get("device_name"), Home: get("home"), StateDir: get("state_dir"),
		HomeRadiusM: v.GetFloat64("mqtt.home_radius_m"),
		Controls:    mqttControls(v),
	}
	for _, d := range []struct {
		key string
		to  *time.Duration
	}{{"poll", &m.Poll}, {"health", &m.Health}, {"stale", &m.Stale}, {"command_interval", &m.CommandInterval},
		{"engine_interval", &m.EngineInterval}, {"refresh_interval", &m.RefreshInterval}} {
		if *d.to, err = mqttDuration(v, d.key); err != nil {
			return nil, err
		}
	}
	if m.Password == "" && m.PasswordFile != "" {
		b, err := os.ReadFile(m.PasswordFile)
		if err != nil {
			return nil, fmt.Errorf("mqtt.password_file: %w", err)
		}
		m.Password = strings.TrimRight(string(b), "\r\n")
	}

	return m, nil
}

// readMQTTConfig reads the config file (a missing one is fine) and binds the
// MCS_MQTT_<KEY> variables.
func readMQTTConfig(configPath string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("toml")
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		v.SetConfigName("config")
		v.AddConfigPath(filepath.Join(home, ".config", "mcs"))
	}
	if err := v.ReadInConfig(); err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}
	v.SetEnvPrefix("MCS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	keys := []string{"broker", "username", "password", "password_file", "client_id", "prefix", "discovery", "device_name",
		"home", "home_radius_m", "state_dir", "poll", "health", "stale", "command_interval", "engine_interval",
		"refresh_interval", "controls"}
	for _, k := range keys {
		_ = v.BindEnv("mqtt." + k)
	}

	return v, nil
}

// mqttControls reads mqtt.controls: a list, or a comma-separated string (as
// the environment gives it); nil when unset.
func mqttControls(v *viper.Viper) []string {
	c := v.Get("mqtt.controls")
	if c == nil {
		return nil
	}
	if s, ok := c.(string); ok {
		return strings.Split(s, ",")
	}

	return v.GetStringSlice("mqtt.controls")
}

// mqttDuration reads mqtt.<k> as a duration (0 when unset).
func mqttDuration(v *viper.Viper, k string) (time.Duration, error) {
	s := v.GetString("mqtt." + k)
	if s == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("mqtt.%s: %w", k, err)
	}

	return d, nil
}

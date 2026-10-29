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
	get := func(k string) string { return v.GetString("mqtt." + k) }
	dur := func(k string) (time.Duration, error) {
		if get(k) == "" {
			return 0, nil
		}
		d, err := time.ParseDuration(get(k))
		if err != nil {
			return 0, fmt.Errorf("mqtt.%s: %w", k, err)
		}

		return d, nil
	}
	m := &MQTT{
		Broker: get("broker"), Username: get("username"), Password: os.Getenv("MCS_MQTT_PASSWORD"),
		PasswordFile: get("password_file"), ClientID: get("client_id"), Prefix: get("prefix"),
		Discovery: get("discovery"), DeviceName: get("device_name"), Home: get("home"), StateDir: get("state_dir"),
		HomeRadiusM: v.GetFloat64("mqtt.home_radius_m"),
	}
	if c := v.Get("mqtt.controls"); c != nil {
		if s, ok := c.(string); ok {
			m.Controls = strings.Split(s, ",")
		} else {
			m.Controls = v.GetStringSlice("mqtt.controls")
		}
	}
	for _, d := range []struct {
		key string
		to  *time.Duration
	}{{"poll", &m.Poll}, {"health", &m.Health}, {"stale", &m.Stale}, {"command_interval", &m.CommandInterval},
		{"engine_interval", &m.EngineInterval}, {"refresh_interval", &m.RefreshInterval}} {
		var err error
		if *d.to, err = dur(d.key); err != nil {
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

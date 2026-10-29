package cli

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/cv/mcs/internal/api"
	"github.com/cv/mcs/internal/config"
	"github.com/cv/mcs/internal/mqttd"
	"github.com/spf13/cobra"
)

// exitError carries an exit status out of Execute.
type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string { return e.err.Error() }
func (e *exitError) Unwrap() error { return e.err }

// ExitCode is the process's exit status.
func (e *exitError) ExitCode() int { return e.code }

// mqttFlags are the command line's values; only the ones given override the config.
type mqttFlags struct {
	broker, username, passwordFile, clientID, prefix, discovery, deviceName, home, stateDir, logFormat string
	homeRadius                                                                                         float64
	poll, health, stale, commandInterval, engineInterval, refreshInterval                              time.Duration
	controls                                                                                           []string
	once, dryRun                                                                                       bool
}

// NewMQTTCmd creates the mqtt command.
func NewMQTTCmd() *cobra.Command {
	var f mqttFlags
	cmd := &cobra.Command{
		Use:   "mqtt",
		Short: "Publish the vehicle's status to MQTT (with Home Assistant discovery)",
		Long: `Publish the vehicle's status to MQTT, with Home Assistant discovery, and
carry out the commands enabled with --controls (none by default). Runs until
stopped. It never signs in with a password nor attaches a new session (the
cached session is used, as with MCS_REFRESH_ONLY=1), so the phone stays signed in.

Settings come from the [mqtt] section of the config file, MCS_MQTT_<KEY>
variables, then flags. The prefix and the broker have no default.

Exit status: 0 stopped, 1 error (restart it), 2 bad configuration, 3 the Mazda
session is lost (renew it with mcs status on a computer; do not restart).`,
		Example: `  # One poll, printed instead of published
  mcs mqtt --prefix mazda/cx90 --once --dry-run

  # Read-only, next to the broker
  mcs mqtt --broker tcp://127.0.0.1:1883 --username mazda --password-file /etc/mcs/mqtt.pw --prefix mazda/cx90

  # Everything
  mcs mqtt --prefix mazda/cx90 --broker tcp://127.0.0.1:1883 --controls all --home 45.50,-73.57`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runMQTT(cmd, &f)
		},
		SilenceUsage: true,
	}
	fl := cmd.Flags()
	fl.StringVar(&f.broker, "broker", "", "broker URL: tcp://host:1883, ssl://host:8883")
	fl.StringVar(&f.username, "username", "", "MQTT login")
	fl.StringVar(&f.passwordFile, "password-file", "", "file holding the MQTT password (or MCS_MQTT_PASSWORD)")
	fl.StringVar(&f.clientID, "client-id", "", "MQTT client id (default mcs-<last prefix level>)")
	fl.StringVar(&f.prefix, "prefix", "", "topic root, e.g. mazda/cx90 (required)")
	fl.StringVar(&f.discovery, "discovery", "", "Home Assistant discovery prefix, or off (default homeassistant)")
	fl.StringVar(&f.deviceName, "device-name", "", "Home Assistant device name (default Mazda <model>)")
	fl.DurationVar(&f.poll, "poll", 0, "cloud status reads (default 5m, at least 5m)")
	fl.DurationVar(&f.health, "health", 0, "health report and climate setting reads (default 30m)")
	fl.DurationVar(&f.stale, "stale", 0, "locks, doors, windows... unavailable after this (default 30m)")
	fl.StringSliceVar(&f.controls, "controls", nil, "commands to accept: lock, unlock, engine, charge, climate, lights, refresh, all")
	fl.DurationVar(&f.commandInterval, "command-interval", 0, "each command at most once per (default 1m)")
	fl.DurationVar(&f.engineInterval, "engine-interval", 0, "engine start at most once per (default 10m)")
	fl.DurationVar(&f.refreshInterval, "refresh-interval", 0, "fresh report at most once per (default 30m)")
	fl.StringVar(&f.home, "home", "", "LAT,LON for presence (home / not_home); coordinates are never published")
	fl.Float64Var(&f.homeRadius, "home-radius", 0, "presence radius in meters (default 200)")
	fl.StringVar(&f.stateDir, "state-dir", "", "where rate limits survive restarts (default ~/.local/state/mcs)")
	fl.BoolVar(&f.once, "once", false, "one poll, then exit")
	fl.BoolVar(&f.dryRun, "dry-run", false, "print what would be published; no broker, no commands")
	fl.StringVar(&f.logFormat, "log-format", "text", "text or json")

	return cmd
}

// mqttConfig merges config file, environment and flags.
func mqttConfig(cmd *cobra.Command, f *mqttFlags) (*mqttd.Config, error) {
	path := ""
	if cliCfg := ConfigFromContext(cmd.Context()); cliCfg != nil {
		path = cliCfg.ConfigFile
	}
	m, err := config.LoadMQTT(path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", mqttd.ErrConfig, err)
	}
	c := mqttd.Defaults()
	changed := cmd.Flags().Changed
	applyMQTTStrings(&c, f, m, changed)
	if c.Password, err = mqttPassword(f, m, changed); err != nil {
		return nil, err
	}
	applyMQTTDurations(&c, f, m, changed)
	c.Controls = m.Controls
	if changed("controls") {
		c.Controls = f.controls
	}
	if c.Home, err = mqttHome(f, m, changed); err != nil {
		return nil, err
	}
	if c.StateDir == "" {
		if dir, err := os.UserHomeDir(); err == nil {
			c.StateDir = filepath.Join(dir, ".local", "state", "mcs")
		}
	}
	c.Once, c.DryRun = f.once, f.dryRun
	if c.DryRun {
		c.Controls = nil // a dry run never moves the car
		c.Once = true
	}

	return &c, c.Validate()
}

// applyMQTTStrings sets each string setting from its flag if given, else from
// the config; an empty value keeps the default.
func applyMQTTStrings(c *mqttd.Config, f *mqttFlags, m *config.MQTT, changed func(string) bool) {
	pick := func(flagName, flagVal, cfgVal string) string {
		if changed(flagName) {
			return flagVal
		}

		return cfgVal
	}
	c.Broker = cmp.Or(pick("broker", f.broker, m.Broker), c.Broker)
	c.Username = cmp.Or(pick("username", f.username, m.Username), c.Username)
	c.ClientID = cmp.Or(pick("client-id", f.clientID, m.ClientID), c.ClientID)
	c.Prefix = cmp.Or(pick("prefix", f.prefix, m.Prefix), c.Prefix)
	c.Discovery = cmp.Or(pick("discovery", f.discovery, m.Discovery), c.Discovery)
	c.DeviceName = cmp.Or(pick("device-name", f.deviceName, m.DeviceName), c.DeviceName)
	c.StateDir = cmp.Or(pick("state-dir", f.stateDir, m.StateDir), c.StateDir)
}

// mqttPassword is the --password-file's password if given, else the config's.
func mqttPassword(f *mqttFlags, m *config.MQTT, changed func(string) bool) (string, error) {
	if !changed("password-file") {
		return m.Password, nil
	}
	b, err := os.ReadFile(f.passwordFile)
	if err != nil {
		return "", fmt.Errorf("%w: --password-file: %w", mqttd.ErrConfig, err)
	}

	return strings.TrimRight(string(b), "\r\n"), nil
}

// applyMQTTDurations sets each interval from its flag if given (even 0), else
// from the config unless unset there.
func applyMQTTDurations(c *mqttd.Config, f *mqttFlags, m *config.MQTT, changed func(string) bool) {
	for _, d := range []struct {
		flag     string
		dst      *time.Duration
		flagV, v time.Duration
	}{{"poll", &c.Poll, f.poll, m.Poll}, {"health", &c.Health, f.health, m.Health}, {"stale", &c.Stale, f.stale, m.Stale},
		{"command-interval", &c.CommandInterval, f.commandInterval, m.CommandInterval},
		{"engine-interval", &c.EngineInterval, f.engineInterval, m.EngineInterval},
		{"refresh-interval", &c.RefreshInterval, f.refreshInterval, m.RefreshInterval}} {
		switch {
		case changed(d.flag):
			*d.dst = d.flagV
		case d.v != 0:
			*d.dst = d.v
		}
	}
}

// mqttHome is the home for presence (nil without one); the radius defaults to
// 200 m.
func mqttHome(f *mqttFlags, m *config.MQTT, changed func(string) bool) (*mqttd.Home, error) {
	home, radius := m.Home, m.HomeRadiusM
	if changed("home") {
		home = f.home
	}
	if changed("home-radius") {
		radius = f.homeRadius
	}
	if home == "" {
		return nil, nil
	}
	if radius == 0 {
		radius = 200
	}

	return mqttd.ParseHome(home, radius)
}

func runMQTT(cmd *cobra.Command, f *mqttFlags) error {
	cfg, err := mqttConfig(cmd, f)
	if err != nil {
		return &exitError{2, err}
	}
	var handler slog.Handler = slog.NewTextHandler(cmd.ErrOrStderr(), nil)
	if f.logFormat == "json" {
		handler = slog.NewJSONHandler(cmd.ErrOrStderr(), nil)
	}
	log := slog.New(handler)
	ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// Never a password login nor a new session: either could sign the phone out.
	_ = os.Setenv("MCS_REFRESH_ONLY", "1")
	car, err := newMQTTVehicle(ctx)
	if err != nil {
		if mqttd.AuthLostText(err) {
			return &exitError{3, fmt.Errorf("%w: %w", mqttd.ErrAuthLost, err)}
		}

		return err
	}
	var broker mqttd.Broker = mqttd.NewPahoBroker(cfg)
	if cfg.DryRun {
		broker = mqttd.PrintBroker{Out: cmd.OutOrStdout()}
	}
	err = mqttd.New(cfg, car, broker, log).Run(ctx)
	if errors.Is(err, mqttd.ErrAuthLost) {
		return &exitError{3, err}
	}

	return err
}

// mqttVehicle is the car through the API client; the session cache is saved
// after every call (tokens rotate).
type mqttVehicle struct {
	ctx    context.Context // carries the CLI config (cache file)
	client *api.Client
	info   VehicleInfo
}

func newMQTTVehicle(ctx context.Context) (*mqttVehicle, error) {
	client, info, err := setupVehicleClient(ctx)
	if err != nil {
		return nil, err
	}

	return &mqttVehicle{ctx: ctx, client: client, info: info}, nil
}

// Model is the model's first word, e.g. CX-90.
func (v *mqttVehicle) Model() string {
	if f := strings.Fields(v.info.ModelName); len(f) > 0 {
		return f[0]
	}

	return "vehicle"
}

// Status reads the EV and vehicle reports as mcs status --json does; with
// refresh it first asks the car for a fresh EV report.
func (v *mqttVehicle) Status(ctx context.Context, refresh bool) (map[string]any, bool, error) {
	defer saveClientCache(v.ctx, v.client)
	ev, err := v.client.GetEVVehicleStatus(ctx, v.vin())
	if err != nil {
		return nil, false, err
	}
	answered := false
	if refresh {
		if ev, answered, err = v.refreshEV(ctx, ev); err != nil {
			return nil, false, err
		}
	}
	vs, err := v.client.GetVehicleStatus(ctx, v.vin())
	if err != nil {
		return nil, false, err
	}
	text, err := displayAllStatusJSON(vs, ev, v.info)
	if err != nil {
		return nil, false, err
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		return nil, false, err
	}

	return data, answered, nil
}

func (v *mqttVehicle) Health(ctx context.Context) (map[string]any, error) {
	defer saveClientCache(v.ctx, v.client)

	return v.client.GetHealthReport(ctx, v.vin())
}

func (v *mqttVehicle) HVAC(ctx context.Context) (map[string]any, error) {
	defer saveClientCache(v.ctx, v.client)

	return v.client.GetHVACSetting(ctx, v.vin())
}

// Do sends a button command, or the climate setting.
func (v *mqttVehicle) Do(ctx context.Context, name string, s mqttd.Setting) error {
	defer saveClientCache(v.ctx, v.client)
	c, vin := v.client, v.vin()
	if send, ok := buttonCalls(c)[name]; ok {
		return send(ctx, vin)
	}
	if mqttd.IsSetting(name) {
		return c.SetHVACSetting(ctx, vin, s.Temperature, api.Celsius, s.Front, s.Rear)
	}

	return fmt.Errorf("unknown command %q", name)
}

func (v *mqttVehicle) vin() string { return string(v.info.InternalVIN) }

// refreshEV asks the car for a fresh EV report and waits up to 2 minutes for
// it; it reports whether one arrived (else ev stays).
func (v *mqttVehicle) refreshEV(ctx context.Context, ev *api.EVVehicleStatusResponse) (*api.EVVehicleStatusResponse, bool, error) {
	before, _ := ev.GetOccurrenceDate()
	if err := v.client.RefreshVehicleStatus(ctx, v.vin()); err != nil {
		return nil, false, err
	}
	for deadline := time.Now().Add(2 * time.Minute); time.Now().Before(deadline); {
		select {
		case <-ctx.Done():
			return nil, false, ctx.Err()
		case <-time.After(20 * time.Second):
		}
		if next, err := v.client.GetEVVehicleStatus(ctx, v.vin()); err == nil {
			if after, _ := next.GetOccurrenceDate(); after != before {
				return next, true, nil
			}
		}
	}

	return ev, false, nil
}

// buttonCalls maps each button command to its API call.
func buttonCalls(c *api.Client) map[string]func(context.Context, string) error {
	return map[string]func(context.Context, string) error{
		mqttd.CmdLock:        c.DoorLock,
		mqttd.CmdUnlock:      c.DoorUnlock,
		mqttd.CmdEngineStart: c.EngineStart,
		mqttd.CmdEngineStop:  c.EngineStop,
		mqttd.CmdChargeStart: c.ChargeStart,
		mqttd.CmdChargeStop:  c.ChargeStop,
		mqttd.CmdClimateOn:   c.HVACOn,
		mqttd.CmdClimateOff:  c.HVACOff,
		mqttd.CmdLightsOn:    c.LightsOn,
		mqttd.CmdLightsOff:   c.LightsOff,
	}
}

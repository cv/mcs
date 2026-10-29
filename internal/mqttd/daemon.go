package mqttd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cv/mcs/internal/api"
)

// Setting is the remote climate setting a command sends.
type Setting struct {
	Temperature float64
	Front, Rear bool
}

// Vehicle is the car as the daemon uses it (the CLI implements it with the API client).
type Vehicle interface {
	// Status is mcs status --json's data; with refresh it first asks the car for
	// a fresh EV report and reports whether one arrived.
	Status(ctx context.Context, refresh bool) (data map[string]any, answered bool, err error)
	Health(ctx context.Context) (map[string]any, error)
	HVAC(ctx context.Context) (map[string]any, error)
	// Do sends a button command, or the climate setting (name is a setting).
	Do(ctx context.Context, name string, s Setting) error
	Model() string
}

// Broker is the MQTT connection. onConnect runs after every (re)connect;
// onMessage gets every message on the subscribed topics.
type Broker interface {
	Start(onConnect func(), onMessage func(topic string, payload []byte, retained bool)) error
	Publish(topic, payload string, retain bool) error
	Subscribe(topics ...string) error
	Stop()
}

// Daemon is mcs mqtt: one poll loop; commands run in it, one at a time, so only
// one API call is ever in flight (the session's tokens rotate).
type Daemon struct {
	cfg    *Config
	car    Vehicle
	broker Broker
	log    *slog.Logger
	now    func() time.Time
	limits *Limits

	connected chan struct{}
	wake      chan struct{}
	commands  chan Command
	refresh   atomic.Bool

	data, health, hvac map[string]any // hvac: parsed (HVACSetting's keys)
	healthAt           time.Time
}

// New makes a daemon; Run starts it.
func New(cfg *Config, car Vehicle, broker Broker, log *slog.Logger) *Daemon {
	limits := ""
	if cfg.StateDir != "" {
		limits = filepath.Join(cfg.StateDir, "mqtt-limits.json")
	}

	return &Daemon{
		cfg: cfg, car: car, broker: broker, log: log, now: time.Now, limits: LoadLimits(limits),
		connected: make(chan struct{}, 1), wake: make(chan struct{}, 1), commands: make(chan Command, 8),
	}
}

func (d *Daemon) topic(sub string) string { return d.cfg.Prefix + "/" + sub }

func signal(ch chan struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

// onMessage runs on the broker's goroutine: it only queues. A retained message
// never acts (it would wake or move the car on every reconnect).
func (d *Daemon) onMessage(topic string, payload []byte, retained bool) {
	if retained {
		return
	}
	if topic == d.topic("refresh/set") {
		if d.cfg.Has(GroupRefresh) && string(payload) == "PRESS" {
			d.refresh.Store(true)
			signal(d.wake)
		}

		return
	}
	if cmd, ok := ParseCommand(d.cfg, topic, payload); ok {
		select {
		case d.commands <- cmd:
			signal(d.wake)
		default:
		}
	}
}

// Run polls until ctx ends (nil), the session is lost (ErrAuthLost) or, with
// Once, after one poll.
func (d *Daemon) Run(ctx context.Context) error {
	if err := d.broker.Start(func() { signal(d.connected) }, d.onMessage); err != nil {
		return err
	}
	defer d.broker.Stop()
	// The first poll waits for the broker, so its state is not lost.
	select {
	case <-ctx.Done():
		return nil
	case <-d.connected:
		d.announce()
	}
	timer := time.NewTimer(0)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			_ = d.broker.Publish(d.topic("status"), "offline", true)

			return nil
		case <-d.connected:
			d.announce()

			continue
		case <-d.wake:
		case <-timer.C:
		}
		err := d.cycle(ctx)
		if errors.Is(err, ErrAuthLost) {
			return err
		}
		if d.cfg.Once {
			return err
		}
		timer.Reset(d.cfg.Poll)
	}
}

// announce runs after every connect: status, discovery, subscriptions, and the
// last state again.
func (d *Daemon) announce() {
	_ = d.broker.Publish(d.topic("status"), "offline", true)
	if d.cfg.Discovery != "off" {
		for topic, cfg := range Discovery(d.cfg, d.car.Model()) {
			payload := ""
			if cfg != nil {
				b, _ := json.Marshal(cfg)
				payload = string(b)
			}
			if err := d.broker.Publish(topic, payload, true); err != nil {
				d.log.Warn("discovery publish failed", "err", err)

				return
			}
		}
	}
	var topics []string
	if d.cfg.Has(GroupRefresh) {
		topics = append(topics, d.topic("refresh/set"))
	}
	if d.cfg.HasCommands() {
		topics = append(topics, d.topic("cmd/+/set"))
	}
	if len(topics) > 0 {
		if err := d.broker.Subscribe(topics...); err != nil {
			d.log.Warn("subscribe failed", "err", err)
		}
	}
	if d.data != nil {
		d.publishState()
	}
	d.log.Info("connected", "broker", d.cfg.Broker, "controls", strings.Join(d.cfg.Controls, ","))
}

func (d *Daemon) publishState() {
	b, _ := json.Marshal(d.data)
	if d.broker.Publish(d.topic("state"), string(b), true) == nil {
		_ = d.broker.Publish(d.topic("error"), "", true)
		_ = d.broker.Publish(d.topic("status"), "online", true)
	}
}

// AuthLostText reports whether an API error means the session is gone.
func AuthLostText(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	for _, k := range []string{"explicit sign-in", "refresh failed", "multiple devices", "another device", "sign-in refused"} {
		if strings.Contains(s, k) {
			return true
		}
	}

	return false
}

func (d *Daemon) authLost() error {
	_ = d.broker.Publish(d.topic("status"), "offline", true)
	_ = d.broker.Publish(d.topic("error"), "authentication required; polling paused", true)
	d.log.Error("Mazda session unavailable; stopping (renew the cache, then restart)")

	return ErrAuthLost
}

func stamp(t time.Time) string { return t.Local().Format("15:04") }

// cycle runs queued commands, then one poll.
func (d *Daemon) cycle(ctx context.Context) error {
	ran := false
	for done := false; !done; {
		select {
		case cmd := <-d.commands:
			_ = d.broker.Publish(d.topic("command_status"), strings.ReplaceAll(cmd.Name, "_", " ")+": sending", true)
			line, err := d.run(ctx, cmd)
			if AuthLostText(err) {
				return d.authLost()
			}
			_ = d.broker.Publish(d.topic("command_status"), line, true)
			ran = true
		default:
			done = true
		}
	}
	refresh := d.refresh.Swap(false)
	if refresh {
		if wait := d.limits.Wait("refresh", d.now(), d.cfg.RefreshInterval); wait > 0 {
			refresh = false
			_ = d.broker.Publish(d.topic("refresh_status"), fmt.Sprintf("not sent: one request per %d min (try again in %d min)",
				minutes(d.cfg.RefreshInterval), minutes(wait)), true)
		} else {
			d.limits.Mark("refresh", d.now())
			_ = d.broker.Publish(d.topic("refresh_status"), "waiting for the car (up to 2 min)", true)
			d.log.Info("fresh report requested from the car")
		}
	}
	data, answered, err := d.car.Status(ctx, refresh)
	if err == nil {
		d.data = Snapshot(data, d.now(), d.cfg.Home)
		if d.data[srcEV] == nil && d.data[srcDoors] == nil {
			err = errors.New("no dated vehicle reports")
		}
	}
	if err != nil {
		if AuthLostText(err) {
			return d.authLost()
		}
		d.log.Warn("poll failed", "err", errorKind(err))
		_ = d.broker.Publish(d.topic("status"), "offline", true)
		if refresh {
			_ = d.broker.Publish(d.topic("refresh_status"), "request failed", true)
		}
		d.data = nil

		return err
	}
	if ran || d.now().Sub(d.healthAt) >= d.cfg.Health {
		if err := d.readHealth(ctx); err != nil {
			return err
		}
	}
	maps.Copy(d.data, Health(d.health))
	if d.hvac == nil {
		maps.Copy(d.data, HVACSetting(nil))
	} else {
		maps.Copy(d.data, d.hvac)
	}
	d.publishState()
	if refresh {
		msg := "car did not answer in time (" + stamp(d.now()) + ")"
		if answered {
			msg = "car answered at " + stamp(d.now()) + " (battery and charging; locks update when the car is used)"
		}
		_ = d.broker.Publish(d.topic("refresh_status"), msg, true)
	}
	d.log.Info("poll", "ev", d.data[srcEV], "vehicle", d.data[srcVehicle], "doors", d.data[srcDoors])

	return nil
}

// readHealth reads the health report and the climate setting; a failure keeps
// the last ones.
func (d *Daemon) readHealth(ctx context.Context) error {
	for _, r := range []struct {
		name string
		read func(context.Context) (map[string]any, error)
	}{{"health", d.car.Health}, {"hvac", d.car.HVAC}} {
		raw, err := r.read(ctx)
		if AuthLostText(err) {
			return d.authLost()
		}
		if err != nil {
			d.log.Warn(r.name+" read failed", "err", errorKind(err))

			continue
		}
		if r.name == "hvac" {
			d.hvac = HVACSetting(raw)
		} else {
			d.health = raw
		}
	}
	d.healthAt = d.now()

	return nil
}

// errorKind is what gets logged of an error: its result code, if any, never
// the rest (it may hold ids).
func errorKind(err error) string {
	if rc, ok := errors.AsType[*api.ResultCodeError](err); ok {
		return "result code " + rc.ResultCode
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}

	return fmt.Sprintf("%T", errors.Unwrap(err))
}

// run sends one command; the result is its status line.
func (d *Daemon) run(ctx context.Context, cmd Command) (string, error) {
	label := strings.ReplaceAll(cmd.Name, "_", " ")
	key, every := cmd.Name, d.cfg.CommandInterval
	if IsSetting(cmd.Name) {
		key = "climate_setting"
	}
	if cmd.Name == "engine_start" {
		every = d.cfg.EngineInterval
	}
	now := d.now()
	if wait := d.limits.Wait(key, now, every); wait > 0 {
		return fmt.Sprintf("%s: not sent, once per %d min (try again in %d min)", label, minutes(every), minutes(wait)), nil
	}
	var s Setting
	if IsSetting(cmd.Name) {
		t, ok := d.hvac[SettingTemperature].(float64)
		if !ok {
			return label + ": not sent, the climate setting is not known yet", nil
		}
		s = Setting{Temperature: t, Front: isTrue(d.hvac[SettingFront]), Rear: isTrue(d.hvac[SettingRear])}
		switch cmd.Name {
		case SettingTemperature:
			s.Temperature = cmd.Value.(float64)
		case SettingFront:
			s.Front = cmd.Value.(bool)
		case SettingRear:
			s.Rear = cmd.Value.(bool)
		}
	}
	d.limits.Mark(key, now)
	d.log.Info("command sent", "command", cmd.Name)
	if err := d.car.Do(ctx, cmd.Name, s); err != nil {
		if AuthLostText(err) {
			return "", err
		}
		kind := errorKind(err)
		d.log.Warn("command failed", "command", cmd.Name, "err", kind)

		return fmt.Sprintf("%s: failed (%s, %s)", label, stamp(now), kind), nil
	}
	if IsSetting(cmd.Name) {
		d.hvac = map[string]any{SettingTemperature: s.Temperature, SettingFront: s.Front, SettingRear: s.Rear}
	}
	d.log.Info("command accepted", "command", cmd.Name)
	// Mazda accepts and passes the command on; the car acts seconds later
	// (MyMazda notifies the phone). The door report does not update from it.
	return label + ": accepted " + stamp(now), nil
}

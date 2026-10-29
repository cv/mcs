package mqttd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cv/mcs/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCar struct {
	mu       sync.Mutex
	calls    []string
	status   string
	statusEr error
	doErr    error
}

func (f *fakeCar) record(s string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, s)
}

func (f *fakeCar) Status(_ context.Context, refresh bool) (map[string]any, bool, error) {
	f.record(map[bool]string{false: "status", true: "status+refresh"}[refresh])
	if f.statusEr != nil {
		return nil, false, f.statusEr
	}
	var d map[string]any

	return d, refresh, json.Unmarshal([]byte(f.status), &d)
}

func (f *fakeCar) Health(context.Context) (map[string]any, error) {
	f.record("health")
	var d map[string]any

	return d, json.Unmarshal([]byte(healthJSON), &d)
}

func (f *fakeCar) HVAC(context.Context) (map[string]any, error) {
	f.record("hvac")

	return map[string]any{"hvacSettings": map[string]any{"Temperature": 17.0, "TemperatureType": 1.0}}, nil
}

func (f *fakeCar) Do(_ context.Context, name string, s Setting) error {
	if IsSetting(name) {
		f.record(name + ":" + strings.TrimRight(strings.TrimRight(jsonNum(s.Temperature), "0"), "."))
	} else {
		f.record(name)
	}

	return f.doErr
}

func jsonNum(v float64) string { b, _ := json.Marshal(v); return string(b) }

func (f *fakeCar) Model() string { return "CX-90" }

type fakeBroker struct {
	mu  sync.Mutex
	pub map[string]string
	log []string
}

func (b *fakeBroker) Start(onConnect func(), _ func(string, []byte, bool)) error {
	onConnect()

	return nil
}

func (b *fakeBroker) Publish(topic, payload string, _ bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.pub == nil {
		b.pub = map[string]string{}
	}
	b.pub[topic] = payload
	b.log = append(b.log, topic+"="+payload)

	return nil
}

func (b *fakeBroker) Subscribe(...string) error { return nil }
func (b *fakeBroker) Stop()                     {}

func (b *fakeBroker) get(topic string) string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.pub[topic]
}

func newTest(t *testing.T, car *fakeCar, controls ...string) (*Daemon, *fakeBroker) {
	t.Helper()
	c := testConfig(controls...)
	c.Once = true
	c.StateDir = t.TempDir()
	b := &fakeBroker{}

	return New(c, car, b, slog.New(slog.NewTextHandler(io.Discard, nil))), b
}

func TestPollPublishesState(t *testing.T) {
	t.Parallel()
	car := &fakeCar{status: statusJSON}
	d, b := newTest(t, car)
	require.NoError(t, d.Run(t.Context()))
	assert.Equal(t, []string{"status", "health", "hvac"}, car.calls)
	assert.Equal(t, "online", b.get("mazda/cx90/status"))
	var state map[string]any
	require.NoError(t, json.Unmarshal([]byte(b.get("mazda/cx90/state")), &state))
	assert.Equal(t, 7000.0, state["oil_remaining_km"])
	assert.Equal(t, 17.0, state["climate_temperature"])
	assert.NotContains(t, b.get("mazda/cx90/state"), "latitude")
	assert.Empty(t, b.get("homeassistant/button/mcs_cx90/unlock/config"), "read-only clears the controls")
	assert.NotEmpty(t, b.get("homeassistant/sensor/mcs_cx90/battery/config"))
}

func TestRetainedNeverActs(t *testing.T) {
	t.Parallel()
	d, _ := newTest(t, &fakeCar{status: statusJSON}, "all")
	for _, topic := range []string{"mazda/cx90/refresh/set", "mazda/cx90/cmd/unlock/set", "mazda/cx90/cmd/engine_start/set"} {
		d.onMessage(topic, []byte("PRESS"), true)
	}
	assert.Empty(t, d.commands)
	assert.False(t, d.refresh.Load())
	d.onMessage("mazda/cx90/cmd/unlock/set", []byte("PRESS"), false)
	assert.Equal(t, Command{Name: "unlock"}, <-d.commands)
}

func TestCommandBeforePollAndRateLimit(t *testing.T) {
	t.Parallel()
	car := &fakeCar{status: statusJSON}
	d, b := newTest(t, car, "all")
	d.onMessage("mazda/cx90/cmd/lock/set", []byte("PRESS"), false)
	d.onMessage("mazda/cx90/cmd/lock/set", []byte("PRESS"), false)
	require.NoError(t, d.Run(t.Context()))
	assert.Equal(t, []string{"lock", "status", "health", "hvac"}, car.calls)
	assert.Contains(t, b.get("mazda/cx90/command_status"), "lock: not sent, once per 1 min")
	assert.Contains(t, strings.Join(b.log, "\n"), "command_status=lock: accepted")

	// The limit survives a restart.
	d2 := New(d.cfg, car, b, d.log)
	assert.Positive(t, d2.limits.Wait("lock", time.Now(), time.Minute))
	assert.FileExists(t, filepath.Join(d.cfg.StateDir, "mqtt-limits.json"))
}

func TestSettingUsesTheKnownSetting(t *testing.T) {
	t.Parallel()
	car := &fakeCar{status: statusJSON}
	d, b := newTest(t, car, "climate")
	d.onMessage("mazda/cx90/cmd/rear_defroster/set", []byte("ON"), false)
	require.NoError(t, d.Run(t.Context()))
	assert.Contains(t, b.get("mazda/cx90/command_status"), "not known yet") // no setting read yet

	d.cfg.Once = true
	d.onMessage("mazda/cx90/cmd/climate_temperature/set", []byte("21"), false)
	require.NoError(t, d.Run(t.Context()))
	assert.Contains(t, car.calls, "climate_temperature:21")
	var state map[string]any
	require.NoError(t, json.Unmarshal([]byte(b.get("mazda/cx90/state")), &state))
	assert.Equal(t, 17.0, state["climate_temperature"], "the setting read after the command wins")
}

func TestFailureShowsOnlyTheCode(t *testing.T) {
	t.Parallel()
	car := &fakeCar{status: statusJSON, doErr: api.NewResultCodeError("920E00", "lock doors")}
	d, b := newTest(t, car, "lock")
	d.onMessage("mazda/cx90/cmd/lock/set", []byte("PRESS"), false)
	require.NoError(t, d.Run(t.Context()))
	assert.Contains(t, b.get("mazda/cx90/command_status"), "result code 920E00")
}

func TestAuthLossStops(t *testing.T) {
	t.Parallel()
	car := &fakeCar{statusEr: errors.New("refresh failed; explicit sign-in required: 400")}
	d, b := newTest(t, car)
	d.cfg.Once = false
	require.ErrorIs(t, d.Run(t.Context()), ErrAuthLost)
	assert.Equal(t, "offline", b.get("mazda/cx90/status"))
	assert.Equal(t, "authentication required; polling paused", b.get("mazda/cx90/error"))
}

func TestEmptyReportNotOnline(t *testing.T) {
	t.Parallel()
	d, b := newTest(t, &fakeCar{status: `{}`})
	require.Error(t, d.Run(t.Context()))
	assert.Equal(t, "offline", b.get("mazda/cx90/status"))
}

func TestRefreshRateLimit(t *testing.T) {
	t.Parallel()
	car := &fakeCar{status: statusJSON}
	d, b := newTest(t, car, "refresh")
	d.onMessage("mazda/cx90/refresh/set", []byte("PRESS"), false)
	require.NoError(t, d.Run(t.Context()))
	assert.Equal(t, "status+refresh", car.calls[0])
	assert.Contains(t, b.get("mazda/cx90/refresh_status"), "car answered")
	d.onMessage("mazda/cx90/refresh/set", []byte("PRESS"), false)
	require.NoError(t, d.Run(t.Context()))
	assert.Contains(t, b.get("mazda/cx90/refresh_status"), "not sent: one request per 30 min")
}

func TestStopsOnContext(t *testing.T) {
	t.Parallel()
	d, b := newTest(t, &fakeCar{status: statusJSON})
	d.cfg.Once = false
	ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
	defer cancel()
	require.NoError(t, d.Run(ctx))
	assert.Equal(t, "offline", b.get("mazda/cx90/status"))
}

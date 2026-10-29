package mqttd

import (
	"errors"
	"fmt"
	"io"
	"time"

	paho "github.com/eclipse/paho.mqtt.golang"
)

// PahoBroker is the real MQTT connection: clean session (a command sent while
// the daemon is away is never run later), retained status will, reconnecting.
type PahoBroker struct {
	cfg    *Config
	client paho.Client
}

// NewPahoBroker connects to cfg.Broker once Start is called.
func NewPahoBroker(cfg *Config) *PahoBroker { return &PahoBroker{cfg: cfg} }

const publishTimeout = 15 * time.Second

// Start connects in the background; onConnect runs after each connect.
func (b *PahoBroker) Start(onConnect func(), onMessage func(string, []byte, bool)) error {
	id := b.cfg.ClientID
	if id == "" {
		id = "mcs-" + b.cfg.ID()
	}
	opts := paho.NewClientOptions().AddBroker(b.cfg.Broker).SetClientID(id).
		SetUsername(b.cfg.Username).SetPassword(b.cfg.Password).
		SetCleanSession(true).SetAutoReconnect(true).SetConnectRetry(true).
		SetConnectRetryInterval(5*time.Second).SetMaxReconnectInterval(time.Minute).
		SetKeepAlive(60*time.Second).SetOrderMatters(false).
		SetWill(b.cfg.Prefix+"/status", "offline", 1, true).
		SetOnConnectHandler(func(paho.Client) { onConnect() }).
		SetDefaultPublishHandler(func(_ paho.Client, m paho.Message) { onMessage(m.Topic(), m.Payload(), m.Retained()) })
	b.client = paho.NewClient(opts)
	b.client.Connect() // with ConnectRetry it completes on the first successful connect

	return nil
}

// Publish sends at QoS 1 and waits for the broker's answer.
func (b *PahoBroker) Publish(topic, payload string, retain bool) error {
	t := b.client.Publish(topic, 1, retain, payload)
	if !t.WaitTimeout(publishTimeout) {
		return errors.New("mqtt publish unconfirmed")
	}

	return t.Error()
}

// Subscribe subscribes at QoS 1; messages go to Start's onMessage.
func (b *PahoBroker) Subscribe(topics ...string) error {
	filters := map[string]byte{}
	for _, t := range topics {
		filters[t] = 1
	}
	t := b.client.SubscribeMultiple(filters, nil)
	if !t.WaitTimeout(publishTimeout) {
		return errors.New("mqtt subscribe unconfirmed")
	}

	return t.Error()
}

// Stop disconnects (after up to a second for pending messages).
func (b *PahoBroker) Stop() {
	if b.client != nil {
		b.client.Disconnect(1000)
	}
}

// PrintBroker is --dry-run: every publish is printed, nothing is received.
type PrintBroker struct{ Out io.Writer }

// Start "connects" at once.
func (p PrintBroker) Start(onConnect func(), _ func(string, []byte, bool)) error {
	onConnect()

	return nil
}

// Publish prints the topic and payload (an empty retained payload clears a topic).
func (p PrintBroker) Publish(topic, payload string, retain bool) error {
	mark := ""
	if retain {
		mark = " (retained)"
	}
	if payload == "" {
		payload = "<clear>"
	}
	_, err := fmt.Fprintf(p.Out, "%s%s  %s\n", topic, mark, payload)

	return err
}

// Subscribe does nothing in a dry run.
func (p PrintBroker) Subscribe(...string) error { return nil }

// Stop does nothing in a dry run.
func (p PrintBroker) Stop() {}

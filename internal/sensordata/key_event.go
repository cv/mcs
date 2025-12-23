package sensordata

import (
	"fmt"
	mathrand "math/rand"
	"strings"
	"time"
)

// KeyEvent represents a key event
type KeyEvent struct {
	time             int
	idCharCodeSum    int
	longerThanBefore bool
}

// ToString converts KeyEvent to string format
func (k *KeyEvent) ToString() string {
	if k.longerThanBefore {
		return fmt.Sprintf("2,%d,%d,1;", k.time, k.idCharCodeSum)
	}
	return fmt.Sprintf("2,%d,%d;", k.time, k.idCharCodeSum)
}

// KeyEventList manages a list of key events
type KeyEventList struct {
	keyEvents []*KeyEvent
}

// NewKeyEventList creates a new KeyEventList
func NewKeyEventList() *KeyEventList {
	return &KeyEventList{
		keyEvents: []*KeyEvent{},
	}
}

// Randomize generates random key events
func (k *KeyEventList) Randomize(sensorCollectionStartTimestamp time.Time) {
	k.keyEvents = []*KeyEvent{}

	if mathrand.Intn(20) > 0 {
		return
	}

	nowTimestamp := time.Now().UTC()
	timeSinceSensorCollectionStart := int(nowTimestamp.Sub(sensorCollectionStartTimestamp).Milliseconds())

	if timeSinceSensorCollectionStart < 10000 {
		return
	}

	eventCount := mathrand.Intn(3) + 2
	idCharCodeSum := mathrand.Intn(2) + 517

	for i := 0; i < eventCount; i++ {
		var eventTime int
		if i == 0 {
			eventTime = mathrand.Intn(3000) + 5000
		} else {
			eventTime = mathrand.Intn(40) + 10
		}

		k.keyEvents = append(k.keyEvents, &KeyEvent{
			time:             eventTime,
			idCharCodeSum:    idCharCodeSum,
			longerThanBefore: mathrand.Intn(2) == 0,
		})
	}
}

// ToString converts KeyEventList to string format
func (k *KeyEventList) ToString() string {
	var sb strings.Builder
	for _, event := range k.keyEvents {
		sb.WriteString(event.ToString())
	}
	return sb.String()
}

// GetSum calculates the sum of event values
func (k *KeyEventList) GetSum() int {
	sum := 0
	for _, event := range k.keyEvents {
		sum += event.idCharCodeSum
		sum += event.time
		sum += 2
	}
	return sum
}

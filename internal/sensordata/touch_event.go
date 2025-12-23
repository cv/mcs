package sensordata

import (
	"fmt"
	mathrand "math/rand"
	"strings"
	"time"
)

// TouchEvent represents a touch event
type TouchEvent struct {
	eventType    int
	time         int
	pointerCount int
	toolType     int
}

// ToString converts TouchEvent to string format
func (t *TouchEvent) ToString() string {
	return fmt.Sprintf("%d,%d,0,0,%d,1,%d,-1;", t.eventType, t.time, t.pointerCount, t.toolType)
}

// TouchEventList manages a list of touch events
type TouchEventList struct {
	touchEvents []*TouchEvent
}

// NewTouchEventList creates a new TouchEventList
func NewTouchEventList() *TouchEventList {
	return &TouchEventList{
		touchEvents: []*TouchEvent{},
	}
}

// Randomize generates random touch events
func (t *TouchEventList) Randomize(sensorCollectionStartTimestamp time.Time) {
	t.touchEvents = []*TouchEvent{}

	nowTimestamp := time.Now().UTC()
	timeSinceSensorCollectionStart := int(nowTimestamp.Sub(sensorCollectionStartTimestamp).Milliseconds())

	switch {
	case timeSinceSensorCollectionStart < 3000:
		return
	case timeSinceSensorCollectionStart < 5000:
		downTime := timeSinceSensorCollectionStart - mathrand.Intn(1000) - 1000
		t.addTouchSequence(downTime)
	case timeSinceSensorCollectionStart < 10000:
		for i := 0; i < 2; i++ {
			timestampOffset := 0
			if i == 1 {
				timestampOffset = 5000
			}
			downTime := mathrand.Intn(900) + 100 + timestampOffset
			t.addTouchSequence(downTime)
		}
	default:
		for i := 0; i < 3; i++ {
			timestampOffset := 0
			if i == 0 {
				timestampOffset = timeSinceSensorCollectionStart - 9000
			} else {
				timestampOffset = mathrand.Intn(1000) + 2000
			}
			downTime := mathrand.Intn(900) + 100 + timestampOffset
			t.addTouchSequence(downTime)
		}
	}
}

// addTouchSequence adds a complete touch sequence (down, moves, up) starting at the given time
func (t *TouchEventList) addTouchSequence(downTime int) {
	// down event
	t.touchEvents = append(t.touchEvents, &TouchEvent{
		eventType:    2,
		time:         downTime,
		pointerCount: 1,
		toolType:     1,
	})

	// move events (2-8 events)
	numMoveEvents := mathrand.Intn(7) + 2
	for i := 0; i < numMoveEvents; i++ {
		t.touchEvents = append(t.touchEvents, &TouchEvent{
			eventType:    1,
			time:         mathrand.Intn(47) + 3,
			pointerCount: 1,
			toolType:     1,
		})
	}

	// up event
	t.touchEvents = append(t.touchEvents, &TouchEvent{
		eventType:    3,
		time:         mathrand.Intn(97) + 3,
		pointerCount: 1,
		toolType:     1,
	})
}

// ToString converts TouchEventList to string format
func (t *TouchEventList) ToString() string {
	var sb strings.Builder
	for _, event := range t.touchEvents {
		sb.WriteString(event.ToString())
	}
	return sb.String()
}

// GetSum calculates the sum of event types and times
func (t *TouchEventList) GetSum() int {
	sum := 0
	for _, event := range t.touchEvents {
		sum += event.eventType
		sum += event.time
	}
	return sum
}

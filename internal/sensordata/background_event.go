package sensordata

import (
	"fmt"
	mathrand "math/rand"
	"time"
)

// BackgroundEvent represents a background event
type BackgroundEvent struct {
	eventType int
	timestamp int64
}

// ToString converts BackgroundEvent to string format
func (b *BackgroundEvent) ToString() string {
	return fmt.Sprintf("%d,%d", b.eventType, b.timestamp)
}

// BackgroundEventList manages a list of background events
type BackgroundEventList struct {
	backgroundEvents []*BackgroundEvent
}

// NewBackgroundEventList creates a new BackgroundEventList
func NewBackgroundEventList() *BackgroundEventList {
	return &BackgroundEventList{
		backgroundEvents: []*BackgroundEvent{},
	}
}

// Randomize generates random background events
func (b *BackgroundEventList) Randomize(sensorCollectionStartTimestamp time.Time) {
	b.backgroundEvents = []*BackgroundEvent{}

	if mathrand.Intn(10) > 0 {
		return
	}

	nowTimestamp := time.Now().UTC()
	timeSinceSensorCollectionStart := int(nowTimestamp.Sub(sensorCollectionStartTimestamp).Milliseconds())

	if timeSinceSensorCollectionStart < 10000 {
		return
	}

	pausedTimestamp := timestampToMillis(sensorCollectionStartTimestamp) + int64(mathrand.Intn(3700)+800)
	resumedTimestamp := pausedTimestamp + int64(mathrand.Intn(3000)+2000)

	b.backgroundEvents = append(b.backgroundEvents, &BackgroundEvent{
		eventType: 2,
		timestamp: pausedTimestamp,
	})

	b.backgroundEvents = append(b.backgroundEvents, &BackgroundEvent{
		eventType: 3,
		timestamp: resumedTimestamp,
	})
}

// ToString converts BackgroundEventList to string format
func (b *BackgroundEventList) ToString() string {
	result := ""
	for _, event := range b.backgroundEvents {
		result += event.ToString()
	}
	return result
}

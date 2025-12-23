package sensordata

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTouchEvent_ToString(t *testing.T) {
	event := &TouchEvent{
		eventType:    2,
		time:         1000,
		pointerCount: 1,
		toolType:     1,
	}

	result := event.ToString()
	expected := "2,1000,0,0,1,1,1,-1;"

	assert.Equalf(t, expected, result, "TouchEvent.ToString() = %q, want %q")
}

func TestNewTouchEventList(t *testing.T) {
	list := NewTouchEventList()

	require.NotNil(t, list, "Expected non-nil TouchEventList")

	assert.Lenf(t, list.touchEvents, 0, "Expected empty touchEvents, got %d", len(list.touchEvents))
}

func TestTouchEventList_Randomize_ShortDuration(t *testing.T) {
	list := NewTouchEventList()
	// Create a timestamp that's very recent (less than 3 seconds ago)
	recentTimestamp := time.Now().UTC()

	list.Randomize(recentTimestamp)

	// With duration < 3000ms, should have no events
	assert.Lenf(t, list.touchEvents, 0, "Expected 0 events for short duration, got %d", len(list.touchEvents))
}

func TestTouchEventList_Randomize_MediumDuration(t *testing.T) {
	list := NewTouchEventList()
	// Create a timestamp ~4 seconds ago (between 3000 and 5000ms)
	timestamp := time.Now().UTC().Add(-4 * time.Second)

	list.Randomize(timestamp)

	// Should have events: 1 down + 2-8 move + 1 up = at least 4 events
	assert.GreaterOrEqualf(t, len(list.touchEvents), 4, "Expected at least 4 events for medium duration, got %d", len(list.touchEvents))

	// First event should be down (type 2)
	assert.Equalf(t, 2, list.touchEvents[0].eventType, "Expected first event type 2 (down), got %d", list.touchEvents[0].eventType)

	// Last event should be up (type 3)
	assert.Equalf(t, 3, list.touchEvents[len(list.touchEvents)-1].eventType, "Expected last event type 3 (up), got %d", list.touchEvents[len(list.touchEvents)-1].eventType)
}

func TestTouchEventList_Randomize_LongDuration(t *testing.T) {
	list := NewTouchEventList()
	// Create a timestamp ~12 seconds ago (>= 10000ms)
	timestamp := time.Now().UTC().Add(-12 * time.Second)

	list.Randomize(timestamp)

	// Should have multiple touch sequences (3 sets)
	// Each set has: 1 down + 2-8 move + 1 up = at least 4 events per set
	// 3 sets = at least 12 events
	assert.GreaterOrEqualf(t, len(list.touchEvents), 12, "Expected at least 12 events for long duration, got %d", len(list.touchEvents))
}

func TestTouchEventList_ToString(t *testing.T) {
	list := NewTouchEventList()
	list.touchEvents = []*TouchEvent{
		{eventType: 2, time: 100, pointerCount: 1, toolType: 1},
		{eventType: 1, time: 50, pointerCount: 1, toolType: 1},
		{eventType: 3, time: 25, pointerCount: 1, toolType: 1},
	}

	result := list.ToString()

	// Should contain all events
	assert.True(t, strings.Contains(result, "2,100"), "Expected ToString to contain first event")
	assert.True(t, strings.Contains(result, "1,50"), "Expected ToString to contain second event")
	assert.True(t, strings.Contains(result, "3,25"), "Expected ToString to contain third event")

	// Should end with semicolon
	assert.True(t, strings.HasSuffix(result, ";"), "Expected ToString to end with semicolon")
}

func TestTouchEventList_GetSum(t *testing.T) {
	list := NewTouchEventList()
	list.touchEvents = []*TouchEvent{
		{eventType: 2, time: 100, pointerCount: 1, toolType: 1},
		{eventType: 1, time: 50, pointerCount: 1, toolType: 1},
		{eventType: 3, time: 25, pointerCount: 1, toolType: 1},
	}

	sum := list.GetSum()
	expected := (2 + 100) + (1 + 50) + (3 + 25)

	assert.Equalf(t, expected, sum, "GetSum() = %d, want %d")
}

func TestTouchEventList_GetSum_Empty(t *testing.T) {
	list := NewTouchEventList()
	sum := list.GetSum()

	assert.Equalf(t, 0, sum, "GetSum() for empty list = %d, want 0", sum)
}

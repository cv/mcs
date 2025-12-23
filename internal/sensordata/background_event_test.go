package sensordata

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackgroundEvent_ToString(t *testing.T) {
	event := &BackgroundEvent{
		eventType: 2,
		timestamp: 1701446400000,
	}

	result := event.ToString()
	expected := "2,1701446400000"

	assert.Equal(t, expected, result)
}

func TestNewBackgroundEventList(t *testing.T) {
	list := NewBackgroundEventList()

	require.NotNil(t, list, "Expected non-nil BackgroundEventList")

	assert.Emptyf(t, list.backgroundEvents, "Expected empty backgroundEvents, got %d", len(list.backgroundEvents))
}

func TestBackgroundEventList_Randomize_ShortDuration(t *testing.T) {
	list := NewBackgroundEventList()
	// Create a timestamp that's less than 10 seconds ago
	timestamp := time.Now().UTC().Add(-5 * time.Second)

	list.Randomize(timestamp)

	// With duration < 10000ms, should have no events
	assert.Emptyf(t, list.backgroundEvents, "Expected 0 events for short duration, got %d", len(list.backgroundEvents))
}

func TestBackgroundEventList_ToString(t *testing.T) {
	list := NewBackgroundEventList()
	list.backgroundEvents = []*BackgroundEvent{
		{eventType: 2, timestamp: 1000000},
		{eventType: 3, timestamp: 1005000},
	}

	result := list.ToString()
	expected := "2,10000003,1005000"

	assert.Equal(t, expected, result)
}

func TestBackgroundEventList_ToString_Empty(t *testing.T) {
	list := NewBackgroundEventList()
	result := list.ToString()

	assert.Emptyf(t, result, "ToString() for empty list = %q, want empty string", result)
}

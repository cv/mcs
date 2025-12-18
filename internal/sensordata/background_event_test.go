package sensordata

import (
	"testing"
	"time"
)

func TestBackgroundEvent_ToString(t *testing.T) {
	event := &BackgroundEvent{
		eventType: 2,
		timestamp: 1701446400000,
	}

	result := event.ToString()
	expected := "2,1701446400000"

	if result != expected {
		t.Errorf("BackgroundEvent.ToString() = %q, want %q", result, expected)
	}
}

func TestNewBackgroundEventList(t *testing.T) {
	list := NewBackgroundEventList()

	if list == nil {
		t.Fatal("Expected non-nil BackgroundEventList")
	}

	if len(list.backgroundEvents) != 0 {
		t.Errorf("Expected empty backgroundEvents, got %d", len(list.backgroundEvents))
	}
}

func TestBackgroundEventList_Randomize_ShortDuration(t *testing.T) {
	list := NewBackgroundEventList()
	// Create a timestamp that's less than 10 seconds ago
	timestamp := time.Now().UTC().Add(-5 * time.Second)

	list.Randomize(timestamp)

	// With duration < 10000ms, should have no events
	if len(list.backgroundEvents) != 0 {
		t.Errorf("Expected 0 events for short duration, got %d", len(list.backgroundEvents))
	}
}

func TestBackgroundEventList_ToString(t *testing.T) {
	list := NewBackgroundEventList()
	list.backgroundEvents = []*BackgroundEvent{
		{eventType: 2, timestamp: 1000000},
		{eventType: 3, timestamp: 1005000},
	}

	result := list.ToString()
	expected := "2,10000003,1005000"

	if result != expected {
		t.Errorf("ToString() = %q, want %q", result, expected)
	}
}

func TestBackgroundEventList_ToString_Empty(t *testing.T) {
	list := NewBackgroundEventList()
	result := list.ToString()

	if result != "" {
		t.Errorf("ToString() for empty list = %q, want empty string", result)
	}
}

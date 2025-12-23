package sensordata

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeyEvent_ToString(t *testing.T) {
	tests := []struct {
		name  string
		event *KeyEvent
		want  string
	}{
		{
			name: "not longer than before",
			event: &KeyEvent{
				time:             1000,
				idCharCodeSum:    517,
				longerThanBefore: false,
			},
			want: "2,1000,517;",
		},
		{
			name: "longer than before",
			event: &KeyEvent{
				time:             1500,
				idCharCodeSum:    518,
				longerThanBefore: true,
			},
			want: "2,1500,518,1;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.event.ToString()
			assert.Equalf(t, tt.want, got, "KeyEvent.ToString() = %q, want %q")
		})
	}
}

func TestNewKeyEventList(t *testing.T) {
	list := NewKeyEventList()

	require.NotNil(t, list, "Expected non-nil KeyEventList")

	assert.Lenf(t, list.keyEvents, 0, "Expected empty keyEvents, got %d", len(list.keyEvents))
}

func TestKeyEventList_Randomize_ShortDuration(t *testing.T) {
	list := NewKeyEventList()
	// Create a timestamp that's less than 10 seconds ago
	timestamp := time.Now().UTC().Add(-5 * time.Second)

	list.Randomize(timestamp)

	// With duration < 10000ms, should have no events
	assert.Lenf(t, list.keyEvents, 0, "Expected 0 events for short duration, got %d", len(list.keyEvents))
}

func TestKeyEventList_ToString(t *testing.T) {
	list := NewKeyEventList()
	list.keyEvents = []*KeyEvent{
		{time: 5000, idCharCodeSum: 517, longerThanBefore: false},
		{time: 30, idCharCodeSum: 517, longerThanBefore: true},
	}

	result := list.ToString()

	// Should contain both events
	assert.True(t, strings.Contains(result, "2,5000,517;"), "Expected ToString to contain first event")
	assert.True(t, strings.Contains(result, "2,30,517,1;"), "Expected ToString to contain second event")
}

func TestKeyEventList_ToString_Empty(t *testing.T) {
	list := NewKeyEventList()
	result := list.ToString()

	assert.Equalf(t, "", result, "ToString() for empty list = %q, want empty string", result)
}

func TestKeyEventList_GetSum(t *testing.T) {
	list := NewKeyEventList()
	list.keyEvents = []*KeyEvent{
		{time: 5000, idCharCodeSum: 517, longerThanBefore: false},
		{time: 30, idCharCodeSum: 518, longerThanBefore: true},
	}

	sum := list.GetSum()
	// Sum = (idCharCodeSum + time + 2) for each event
	expected := (517 + 5000 + 2) + (518 + 30 + 2)

	assert.Equalf(t, expected, sum, "GetSum() = %d, want %d")
}

func TestKeyEventList_GetSum_Empty(t *testing.T) {
	list := NewKeyEventList()
	sum := list.GetSum()

	assert.Equalf(t, 0, sum, "GetSum() for empty list = %d, want 0", sum)
}

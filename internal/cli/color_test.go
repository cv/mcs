package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProgressBar(t *testing.T) {
	// Disable colors for consistent test results
	oldColorEnabled := colorEnabled
	SetColorEnabled(false)
	defer SetColorEnabled(oldColorEnabled)

	tests := []struct {
		name     string
		percent  float64
		width    int
		expected string
	}{
		{
			name:     "50% with width 10",
			percent:  50,
			width:    10,
			expected: "[█████░░░░░] 50%",
		},
		{
			name:     "100% with width 10",
			percent:  100,
			width:    10,
			expected: "[██████████] 100%",
		},
		{
			name:     "0% with width 10",
			percent:  0,
			width:    10,
			expected: "[░░░░░░░░░░] 0%",
		},
		{
			name:     "75% with width 10",
			percent:  75,
			width:    10,
			expected: "[███████░░░] 75%",
		},
		{
			name:     "33% with width 10",
			percent:  33,
			width:    10,
			expected: "[███░░░░░░░] 33%",
		},
		{
			name:     "66% with width 10",
			percent:  66,
			width:    10,
			expected: "[██████░░░░] 66%",
		},
		{
			name:     "negative percent clamped to 0",
			percent:  -10,
			width:    10,
			expected: "[░░░░░░░░░░] 0%",
		},
		{
			name:     "over 100% clamped to 100",
			percent:  150,
			width:    10,
			expected: "[██████████] 100%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProgressBar(tt.percent, tt.width)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProgressBar_WithColors(t *testing.T) {
	// Enable colors for color testing
	oldColorEnabled := colorEnabled
	SetColorEnabled(true)
	defer SetColorEnabled(oldColorEnabled)

	tests := []struct {
		name    string
		percent float64
		width   int
	}{
		{name: "high battery (green)", percent: 80, width: 10},
		{name: "medium battery (yellow)", percent: 50, width: 10},
		{name: "low battery (red)", percent: 20, width: 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ProgressBar(tt.percent, tt.width)
			// Just verify it contains ANSI codes (starts with escape sequence)
			assert.Equal(t, byte('\033'), result[0], "Expected colored output to start with ANSI escape sequence")
		})
	}
}

func TestColorize(t *testing.T) {
	// Disable colors
	oldColorEnabled := colorEnabled
	SetColorEnabled(false)
	defer SetColorEnabled(oldColorEnabled)

	text := "test"
	assert.Equal(t, text, Red(text))
	assert.Equal(t, text, Green(text))
	assert.Equal(t, text, Yellow(text))
	assert.Equal(t, text, Bold(text))
}

func TestColorize_WithColors(t *testing.T) {
	// Enable colors
	oldColorEnabled := colorEnabled
	SetColorEnabled(true)
	defer SetColorEnabled(oldColorEnabled)

	text := "test"

	// Test Red
	result := Red(text)
	expected := "\033[31mtest\033[0m"
	assert.Equal(t, expected, result)

	// Test Green
	result = Green(text)
	expected = "\033[32mtest\033[0m"
	assert.Equal(t, expected, result)

	// Test Yellow
	result = Yellow(text)
	expected = "\033[33mtest\033[0m"
	assert.Equal(t, expected, result)

	// Test Bold
	result = Bold(text)
	expected = "\033[1mtest\033[0m"
	assert.Equal(t, expected, result)
}

func TestSetColorEnabled(t *testing.T) {
	oldColorEnabled := colorEnabled
	defer SetColorEnabled(oldColorEnabled)

	SetColorEnabled(true)
	assert.True(t, IsColorEnabled())

	SetColorEnabled(false)
	assert.False(t, IsColorEnabled())
}

func TestColorPressure(t *testing.T) {
	// Disable colors for consistent test results
	oldColorEnabled := colorEnabled
	SetColorEnabled(false)
	defer SetColorEnabled(oldColorEnabled)

	target := 36.0 // Mazda CX-90 recommended

	tests := []struct {
		name     string
		pressure float64
		expected string
	}{
		// Green: within ±3 PSI
		{"exact target", 36.0, "36.0"},
		{"slightly high", 38.0, "38.0"},
		{"slightly low", 34.0, "34.0"},
		{"at +3 boundary", 39.0, "39.0"},
		{"at -3 boundary", 33.0, "33.0"},
		// Yellow: 4-6 PSI off
		{"4 PSI high", 40.0, "40.0"},
		{"6 PSI low", 30.0, "30.0"},
		// Red: >6 PSI off
		{"very high", 45.0, "45.0"},
		{"very low", 25.0, "25.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColorPressure(tt.pressure, target)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestColorPressure_WithColors(t *testing.T) {
	// Enable colors for color testing
	oldColorEnabled := colorEnabled
	SetColorEnabled(true)
	defer SetColorEnabled(oldColorEnabled)

	target := 36.0

	tests := []struct {
		name          string
		pressure      float64
		expectedColor string // The ANSI color code
	}{
		{"green - exact", 36.0, colorGreen},
		{"green - +3", 39.0, colorGreen},
		{"yellow - +4", 40.0, colorYellow},
		{"yellow - -6", 30.0, colorYellow},
		{"red - +7", 43.0, colorRed},
		{"red - very low", 25.0, colorRed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColorPressure(tt.pressure, target)
			// Check it starts with expected color code
			assert.GreaterOrEqual(t, len(result), len(tt.expectedColor))
			assert.Equal(t, tt.expectedColor, result[:len(tt.expectedColor)])
		})
	}
}

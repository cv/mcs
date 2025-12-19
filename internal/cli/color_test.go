package cli

import (
	"testing"
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
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
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
			if result[0] != '\033' {
				t.Errorf("Expected colored output to start with ANSI escape sequence")
			}
		})
	}
}

func TestColorize(t *testing.T) {
	// Disable colors
	oldColorEnabled := colorEnabled
	SetColorEnabled(false)
	defer SetColorEnabled(oldColorEnabled)

	text := "test"
	if result := Red(text); result != text {
		t.Errorf("Expected uncolored text '%s', got '%s'", text, result)
	}
	if result := Green(text); result != text {
		t.Errorf("Expected uncolored text '%s', got '%s'", text, result)
	}
	if result := Yellow(text); result != text {
		t.Errorf("Expected uncolored text '%s', got '%s'", text, result)
	}
	if result := Bold(text); result != text {
		t.Errorf("Expected uncolored text '%s', got '%s'", text, result)
	}
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
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test Green
	result = Green(text)
	expected = "\033[32mtest\033[0m"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test Yellow
	result = Yellow(text)
	expected = "\033[33mtest\033[0m"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test Bold
	result = Bold(text)
	expected = "\033[1mtest\033[0m"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestSetColorEnabled(t *testing.T) {
	oldColorEnabled := colorEnabled
	defer SetColorEnabled(oldColorEnabled)

	SetColorEnabled(true)
	if !IsColorEnabled() {
		t.Error("Expected colors to be enabled")
	}

	SetColorEnabled(false)
	if IsColorEnabled() {
		t.Error("Expected colors to be disabled")
	}
}

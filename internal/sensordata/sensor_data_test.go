package sensordata

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestNewSensorDataBuilder(t *testing.T) {
	builder := NewSensorDataBuilder()

	if builder == nil {
		t.Fatal("Expected non-nil builder")
	}

	if builder.systemInfo == nil {
		t.Error("Expected systemInfo to be initialized")
	}

	if builder.touchEventList == nil {
		t.Error("Expected touchEventList to be initialized")
	}

	if builder.keyEventList == nil {
		t.Error("Expected keyEventList to be initialized")
	}

	if builder.backgroundEventList == nil {
		t.Error("Expected backgroundEventList to be initialized")
	}

	if builder.performanceTestResults == nil {
		t.Error("Expected performanceTestResults to be initialized")
	}

	if builder.deviceInfoTime < 3000 || builder.deviceInfoTime >= 8000 {
		t.Errorf("Expected deviceInfoTime between 3000 and 8000, got %d", builder.deviceInfoTime)
	}
}

func TestSensorDataBuilder_GenerateSensorData(t *testing.T) {
	builder := NewSensorDataBuilder()

	result, err := builder.GenerateSensorData()
	if err != nil {
		t.Fatalf("GenerateSensorData() error = %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty sensor data")
	}

	// Check format: should start with "1,a,"
	if !strings.HasPrefix(result, "1,a,") {
		t.Errorf("Expected result to start with '1,a,', got prefix: %s", result[:min(10, len(result))])
	}

	// Format is: 1,a,<aes_key>,<hmac_key>$<encrypted_data>$<timestamps>
	// Check that it ends with timestamps in format: $<num>,<num>,<num>
	if !regexp.MustCompile(`\$[0-9]+,[0-9]+,[0-9]+$`).MatchString(result) {
		t.Error("Expected timestamp format at end: $<num>,<num>,<num>")
	}
}

func TestCountSeparators(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty string", "", 0},
		{"no separators", "abc", 0},
		{"one separator", "a;b", 1},
		{"multiple separators", "a;b;c;d", 3},
		{"only separators", ";;;", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countSeparators(tt.input)
			if got != tt.want {
				t.Errorf("countSeparators(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestFeistelCipher(t *testing.T) {
	// Test that feistelCipher produces consistent results
	result1 := feistelCipher(100, 50, 12345)
	result2 := feistelCipher(100, 50, 12345)

	if result1 != result2 {
		t.Errorf("feistelCipher should be deterministic: got %d and %d", result1, result2)
	}

	// Test that different inputs produce different outputs
	result3 := feistelCipher(200, 50, 12345)
	if result1 == result3 {
		t.Error("feistelCipher should produce different outputs for different inputs")
	}
}

func TestTimestampToMillis(t *testing.T) {
	// Use a fixed time for testing
	testTime := time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC)
	expectedMillis := testTime.UnixMilli()

	result := timestampToMillis(testTime)
	if result != expectedMillis {
		t.Errorf("timestampToMillis() = %d, want %d", result, expectedMillis)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

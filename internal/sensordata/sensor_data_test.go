package sensordata

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSensorDataBuilder(t *testing.T) {
	builder := NewSensorDataBuilder()

	require.NotNil(t, builder, "Expected non-nil builder")

	assert.NotNil(t, builder.systemInfo)
	assert.NotNil(t, builder.touchEventList)
	assert.NotNil(t, builder.keyEventList)
	assert.NotNil(t, builder.backgroundEventList)
	assert.NotNil(t, builder.performanceTestResults)

	assert.GreaterOrEqual(t, builder.deviceInfoTime, 3000)
	assert.Less(t, builder.deviceInfoTime, 8000)

}

func TestSensorDataBuilder_GenerateSensorData(t *testing.T) {
	builder := NewSensorDataBuilder()

	result, err := builder.GenerateSensorData()
	require.NoError(t, err, "GenerateSensorData() error = %v")

	assert.NotEmpty(t, result, "Expected non-empty sensor data")

	// Check format: should start with "1,a,"
	assert.Truef(t, strings.HasPrefix(result, "1,a,"), "Expected result to start with '1,a,', got prefix: %s", result[:min(10, len(result))])

	// Format is: 1,a,<aes_key>,<hmac_key>$<encrypted_data>$<timestamps>
	// Check that it ends with timestamps in format: $<num>,<num>,<num>
	assert.Regexp(t, `\$[0-9]+,[0-9]+,[0-9]+$`, result, "expected timestamp format at end")
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
			assert.Equalf(t, tt.want, got, "countSeparators(%q) = %d, want %d", tt.input, got, tt.want)
		})
	}
}

func TestFeistelCipher(t *testing.T) {
	// Test that feistelCipher produces consistent results
	result1 := feistelCipher(100, 50, 12345)
	result2 := feistelCipher(100, 50, 12345)

	assert.Equal(t, result2, result1)

	// Test that different inputs produce different outputs
	result3 := feistelCipher(200, 50, 12345)
	assert.NotEqual(t, result3, result1, "feistelCipher should produce different outputs for different inputs")
}

func TestTimestampToMillis(t *testing.T) {
	// Use a fixed time for testing
	testTime := time.Date(2023, 12, 1, 12, 0, 0, 0, time.UTC)
	expectedMillis := testTime.UnixMilli()

	result := timestampToMillis(testTime)
	assert.Equal(t, expectedMillis, result)
}

package sensordata

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSystemInfo(t *testing.T) {
	si := NewSystemInfo()
	require.NotNil(t, si, "Expected non-nil SystemInfo")
}

func TestSystemInfo_Randomize(t *testing.T) {
	si := NewSystemInfo()
	si.Randomize()

	// Check that screen dimensions are set to valid values
	validScreenSizes := map[int]bool{1280: true, 1920: true, 2560: true}
	assert.Truef(t, validScreenSizes[si.screenHeight], "screenHeight %d not in valid set", si.screenHeight)

	// Check battery level is in reasonable range
	if si.batteryLevel < 10 || si.batteryLevel > 90 {
		t.Errorf("batteryLevel %d out of expected range [10, 90]", si.batteryLevel)
	}

	// Check required fields are set
	assert.Equalf(t, "en", si.language, "Expected language 'en', got '%s'", si.language)

	assert.Equalf(t, "com.interrait.mymazda", si.packageName, "Expected packageName 'com.interrait.mymazda', got '%s'", si.packageName)

	assert.Equalf(t, "Pixel 3a", si.buildModel, "Expected buildModel 'Pixel 3a', got '%s'", si.buildModel)

	// Check Android ID is set (16 hex chars)
	assert.Lenf(t, si.androidID, 16, "Expected androidID length 16, got %d", len(si.androidID))
}

func TestSystemInfo_ToString(t *testing.T) {
	si := NewSystemInfo()
	si.Randomize()

	result := si.ToString()

	// Check that result is non-empty
	assert.NotEqual(t, "", result, "Expected non-empty ToString result")

	// Check that it starts with "-1,uaend,-1,"
	assert.Truef(t, strings.HasPrefix(result, "-1,uaend,-1,"), "Expected ToString to start with '-1,uaend,-1,', got '%s'", result[:min(20, len(result))])

	// Check that it contains the package name
	assert.True(t, strings.Contains(result, "com.interrait.mymazda"), "Expected ToString to contain package name")
}

func TestSystemInfo_GetCharCodeSum(t *testing.T) {
	si := NewSystemInfo()
	si.Randomize()

	sum := si.GetCharCodeSum()

	// Sum should be positive and reasonably large given the string content
	if sum <= 0 {
		t.Errorf("Expected positive char code sum, got %d", sum)
	}

	// The sum should be consistent
	sum2 := si.GetCharCodeSum()
	assert.False(t, sum != sum2, "GetCharCodeSum should be deterministic")
}

func TestPercentEncode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"simple", "abc", "abc"},
		{"with space", "a b", "a%20b"},
		{"with comma", "a,b", "a%2Cb"},
		{"with percent", "a%b", "a%25b"},
		{"numbers", "123", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := percentEncode(tt.input)
			assert.Equalf(t, tt.want, got, "percentEncode(%q) = %q, want %q", tt.input, got, tt.want)
		})
	}
}

func TestSumCharCodes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"empty", "", 0},
		{"single char", "a", 97},
		{"hello", "ABC", 65 + 66 + 67},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sumCharCodes(tt.input)
			assert.Equalf(t, tt.want, got, "sumCharCodes(%q) = %d, want %d", tt.input, got, tt.want)
		})
	}
}

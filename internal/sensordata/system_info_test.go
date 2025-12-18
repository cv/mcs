package sensordata

import (
	"strings"
	"testing"
)

func TestNewSystemInfo(t *testing.T) {
	si := NewSystemInfo()
	if si == nil {
		t.Fatal("Expected non-nil SystemInfo")
	}
}

func TestSystemInfo_Randomize(t *testing.T) {
	si := NewSystemInfo()
	si.Randomize()

	// Check that screen dimensions are set to valid values
	validScreenSizes := map[int]bool{1280: true, 1920: true, 2560: true}
	if !validScreenSizes[si.screenHeight] {
		t.Errorf("screenHeight %d not in valid set", si.screenHeight)
	}

	// Check battery level is in reasonable range
	if si.batteryLevel < 10 || si.batteryLevel > 90 {
		t.Errorf("batteryLevel %d out of expected range [10, 90]", si.batteryLevel)
	}

	// Check required fields are set
	if si.language != "en" {
		t.Errorf("Expected language 'en', got '%s'", si.language)
	}

	if si.packageName != "com.interrait.mymazda" {
		t.Errorf("Expected packageName 'com.interrait.mymazda', got '%s'", si.packageName)
	}

	if si.buildModel != "Pixel 3a" {
		t.Errorf("Expected buildModel 'Pixel 3a', got '%s'", si.buildModel)
	}

	// Check Android ID is set (16 hex chars)
	if len(si.androidID) != 16 {
		t.Errorf("Expected androidID length 16, got %d", len(si.androidID))
	}
}

func TestSystemInfo_ToString(t *testing.T) {
	si := NewSystemInfo()
	si.Randomize()

	result := si.ToString()

	// Check that result is non-empty
	if result == "" {
		t.Error("Expected non-empty ToString result")
	}

	// Check that it starts with "-1,uaend,-1,"
	if !strings.HasPrefix(result, "-1,uaend,-1,") {
		t.Errorf("Expected ToString to start with '-1,uaend,-1,', got '%s'", result[:min(20, len(result))])
	}

	// Check that it contains the package name
	if !strings.Contains(result, "com.interrait.mymazda") {
		t.Error("Expected ToString to contain package name")
	}
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
	if sum != sum2 {
		t.Error("GetCharCodeSum should be deterministic")
	}
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
			if got != tt.want {
				t.Errorf("percentEncode(%q) = %q, want %q", tt.input, got, tt.want)
			}
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
			if got != tt.want {
				t.Errorf("sumCharCodes(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

package sensordata

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPerformanceTestResults(t *testing.T) {
	p := NewPerformanceTestResults()

	require.NotNil(t, p, "Expected non-nil PerformanceTestResults")
}

func TestPerformanceTestResults_Randomize(t *testing.T) {
	p := NewPerformanceTestResults()
	p.Randomize()

	// Check that modTestResult is set correctly
	assert.Equalf(t, 16, p.modTestResult, "Expected modTestResult = 16, got %d", p.modTestResult)

	// Check that floatTestResult is set correctly
	assert.Equalf(t, 59, p.floatTestResult, "Expected floatTestResult = 59, got %d", p.floatTestResult)

	// Check that loop test result is in expected range
	if p.loopTestResult < 8500 || p.loopTestResult > 16000 {
		t.Errorf("loopTestResult %d out of expected range [8500, 16000]", p.loopTestResult)
	}

	// Check that iterations are positive
	if p.modTestIters <= 0 {
		t.Errorf("Expected positive modTestIters, got %d", p.modTestIters)
	}
	if p.floatTestIters <= 0 {
		t.Errorf("Expected positive floatTestIters, got %d", p.floatTestIters)
	}
	if p.sqrtTestIters <= 0 {
		t.Errorf("Expected positive sqrtTestIters, got %d", p.sqrtTestIters)
	}
	if p.trigTestIters <= 0 {
		t.Errorf("Expected positive trigTestIters, got %d", p.trigTestIters)
	}
}

func TestPerformanceTestResults_ToString(t *testing.T) {
	p := &PerformanceTestResults{
		modTestResult:   16,
		modTestIters:    500,
		floatTestResult: 59,
		floatTestIters:  1000,
		sqrtTestResult:  100,
		sqrtTestIters:   800,
		trigTestResult:  75000,
		trigTestIters:   750,
		loopTestResult:  10000,
	}

	result := p.ToString()

	// Should be comma-separated values
	parts := strings.Split(result, ",")
	assert.Lenf(t, parts, 9, "Expected 9 comma-separated values, got %d", len(parts))

	expected := "16,500,59,1000,100,800,75000,750,10000"
	assert.Equalf(t, expected, result, "ToString() = %q, want %q")
}

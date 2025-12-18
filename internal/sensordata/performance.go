package sensordata

import (
	mathrand "math/rand"
	"strconv"
	"strings"
)

// PerformanceTestResults represents performance test results
type PerformanceTestResults struct {
	modTestResult    int
	modTestIters     int
	floatTestResult  int
	floatTestIters   int
	sqrtTestResult   int
	sqrtTestIters    int
	trigTestResult   int
	trigTestIters    int
	loopTestResult   int
}

// NewPerformanceTestResults creates a new PerformanceTestResults
func NewPerformanceTestResults() *PerformanceTestResults {
	return &PerformanceTestResults{}
}

// Randomize generates random performance test results
func (p *PerformanceTestResults) Randomize() {
	numIterations1 := (mathrand.Intn(250)*100 + 350*100) - 1
	p.modTestResult = 16
	p.modTestIters = numIterations1 / 100

	numIterations2 := (mathrand.Intn(1437)*100 + 563*100) - 1
	p.floatTestResult = 59
	p.floatTestIters = numIterations2 / 100

	numIterations3 := (mathrand.Intn(1500)*100 + 500*100) - 1
	p.sqrtTestResult = numIterations3 - 899
	p.sqrtTestIters = numIterations3 / 100

	numIterations4 := (mathrand.Intn(1000)*100 + 500*100) - 1
	p.trigTestResult = numIterations4
	p.trigTestIters = numIterations4 / 100

	p.loopTestResult = mathrand.Intn(7500) + 8500
}

// ToString converts PerformanceTestResults to string format
func (p *PerformanceTestResults) ToString() string {
	values := []string{
		strconv.Itoa(p.modTestResult),
		strconv.Itoa(p.modTestIters),
		strconv.Itoa(p.floatTestResult),
		strconv.Itoa(p.floatTestIters),
		strconv.Itoa(p.sqrtTestResult),
		strconv.Itoa(p.sqrtTestIters),
		strconv.Itoa(p.trigTestResult),
		strconv.Itoa(p.trigTestIters),
		strconv.Itoa(p.loopTestResult),
	}
	return strings.Join(values, ",")
}

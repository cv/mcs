package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBold   = "\033[1m"
)

// colorEnabled tracks whether color output is enabled
var colorEnabled = true

func init() {
	// Disable colors if NO_COLOR env var is set or output is not a TTY
	if os.Getenv("NO_COLOR") != "" {
		colorEnabled = false
	}
}

// SetColorEnabled sets whether color output is enabled
func SetColorEnabled(enabled bool) {
	colorEnabled = enabled
}

// IsColorEnabled returns whether color output is enabled
func IsColorEnabled() bool {
	return colorEnabled
}

// IsTTY checks if the given writer is a terminal
func IsTTY(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		// Check if file descriptor refers to a terminal
		// On Unix-like systems, we can check if it's a character device
		fileInfo, err := f.Stat()
		if err != nil {
			return false
		}
		// Check if it's a character device (terminal)
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// colorize wraps text in ANSI color codes if colors are enabled
func colorize(color, text string) string {
	if !colorEnabled {
		return text
	}
	return color + text + colorReset
}

// Red returns text in red
func Red(text string) string {
	return colorize(colorRed, text)
}

// Green returns text in green
func Green(text string) string {
	return colorize(colorGreen, text)
}

// Yellow returns text in yellow
func Yellow(text string) string {
	return colorize(colorYellow, text)
}

// Bold returns text in bold
func Bold(text string) string {
	return colorize(colorBold, text)
}

// Default recommended tire pressure (PSI) - Mazda CX-90 MHEV
const defaultTargetPressurePSI = 36.0

// ColorPressure returns a colored pressure string based on deviation from target
// Green: within ±3 PSI, Yellow: 4-6 PSI off, Red: >6 PSI off
func ColorPressure(pressure float64, targetPSI float64) string {
	text := fmt.Sprintf("%.1f", pressure)
	deviation := pressure - targetPSI
	if deviation < 0 {
		deviation = -deviation
	}

	switch {
	case deviation <= 3:
		return Green(text)
	case deviation <= 6:
		return Yellow(text)
	default:
		return Red(text)
	}
}

// ProgressBar creates a simple ASCII progress bar
// Example: [████████░░] 80%
func ProgressBar(percent float64, width int) string {
	if width <= 0 {
		width = 10
	}

	// Clamp percent to 0-100 range
	if percent < 0 {
		percent = 0
	} else if percent > 100 {
		percent = 100
	}

	filled := int((percent / 100.0) * float64(width))
	empty := width - filled

	// Build the bar
	bar := "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"

	// Add color based on level
	var coloredBar string
	switch {
	case percent >= 80:
		coloredBar = Green(bar)
	case percent >= 30:
		coloredBar = Yellow(bar)
	default:
		coloredBar = Red(bar)
	}

	return fmt.Sprintf("%s %.0f%%", coloredBar, percent)
}

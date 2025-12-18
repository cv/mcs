package cli

import (
	"testing"
)

// TestStartCommand tests the start command
func TestStartCommand(t *testing.T) {
	cmd := NewStartCmd()

	if cmd.Use != "start" {
		t.Errorf("Expected Use to be 'start', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}

// TestStartCommand_NoArgs tests start command with no arguments
func TestStartCommand_NoArgs(t *testing.T) {
	cmd := NewStartCmd()

	// Should accept no args
	if err := cmd.ValidateArgs([]string{}); err != nil {
		t.Errorf("Start command should accept no arguments: %v", err)
	}
}

// TestStopCommand tests the stop command
func TestStopCommand(t *testing.T) {
	cmd := NewStopCmd()

	if cmd.Use != "stop" {
		t.Errorf("Expected Use to be 'stop', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}

// TestStopCommand_NoArgs tests stop command with no arguments
func TestStopCommand_NoArgs(t *testing.T) {
	cmd := NewStopCmd()

	// Should accept no args
	if err := cmd.ValidateArgs([]string{}); err != nil {
		t.Errorf("Stop command should accept no arguments: %v", err)
	}
}

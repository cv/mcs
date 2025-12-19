package cli

import (
	"testing"
)

// TestStartCommand tests the start command
func TestStartCommand(t *testing.T) {
	cmd := NewStartCmd()
	assertCommandBasics(t, cmd, "start")
}

// TestStartCommand_NoArgs tests start command with no arguments
func TestStartCommand_NoArgs(t *testing.T) {
	cmd := NewStartCmd()
	assertNoArgsCommand(t, cmd)
}

// TestStopCommand tests the stop command
func TestStopCommand(t *testing.T) {
	cmd := NewStopCmd()
	assertCommandBasics(t, cmd, "stop")
}

// TestStopCommand_NoArgs tests stop command with no arguments
func TestStopCommand_NoArgs(t *testing.T) {
	cmd := NewStopCmd()
	assertNoArgsCommand(t, cmd)
}

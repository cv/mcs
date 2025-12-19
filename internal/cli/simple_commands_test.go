package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestSimpleCommands tests basic properties of simple control commands
func TestSimpleCommands(t *testing.T) {
	tests := []struct {
		name        string
		cmdFactory  func() *cobra.Command
		expectedUse string
	}{
		{"lock", NewLockCmd, "lock"},
		{"unlock", NewUnlockCmd, "unlock"},
		{"start", NewStartCmd, "start"},
		{"stop", NewStopCmd, "stop"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmdFactory()
			assertCommandBasics(t, cmd, tt.expectedUse)
		})
	}
}

// TestSimpleCommands_NoArgs tests that simple control commands reject arguments
func TestSimpleCommands_NoArgs(t *testing.T) {
	tests := []struct {
		name       string
		cmdFactory func() *cobra.Command
	}{
		{"lock", NewLockCmd},
		{"unlock", NewUnlockCmd},
		{"start", NewStartCmd},
		{"stop", NewStopCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.cmdFactory()
			assertNoArgsCommand(t, cmd)
		})
	}
}

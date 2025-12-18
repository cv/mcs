package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestChargeCommand tests the charge command
func TestChargeCommand(t *testing.T) {
	cmd := NewChargeCmd()

	if cmd.Use != "charge" {
		t.Errorf("Expected Use to be 'charge', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}

// TestChargeCommand_StartSubcommand tests charge start subcommand
func TestChargeCommand_StartSubcommand(t *testing.T) {
	cmd := NewChargeCmd()

	// Find start subcommand
	var startCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "start" {
			startCmd = subCmd
			break
		}
	}

	if startCmd == nil {
		t.Fatal("Expected start subcommand to exist")
	}

	if startCmd.Short == "" {
		t.Error("Expected start subcommand to have a description")
	}

	// Should accept no args
	if err := startCmd.ValidateArgs([]string{}); err != nil {
		t.Errorf("Start subcommand should accept no arguments: %v", err)
	}
}

// TestChargeCommand_StopSubcommand tests charge stop subcommand
func TestChargeCommand_StopSubcommand(t *testing.T) {
	cmd := NewChargeCmd()

	// Find stop subcommand
	var stopCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "stop" {
			stopCmd = subCmd
			break
		}
	}

	if stopCmd == nil {
		t.Fatal("Expected stop subcommand to exist")
	}

	if stopCmd.Short == "" {
		t.Error("Expected stop subcommand to have a description")
	}

	// Should accept no args
	if err := stopCmd.ValidateArgs([]string{}); err != nil {
		t.Errorf("Stop subcommand should accept no arguments: %v", err)
	}
}

package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestClimateCommand tests the climate command
func TestClimateCommand(t *testing.T) {
	cmd := NewClimateCmd()

	if cmd.Use != "climate" {
		t.Errorf("Expected Use to be 'climate', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}

// TestClimateCommand_OnSubcommand tests climate on subcommand
func TestClimateCommand_OnSubcommand(t *testing.T) {
	cmd := NewClimateCmd()

	// Find on subcommand
	var onCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "on" {
			onCmd = subCmd
			break
		}
	}

	if onCmd == nil {
		t.Fatal("Expected on subcommand to exist")
	}

	if onCmd.Short == "" {
		t.Error("Expected on subcommand to have a description")
	}

	// Should accept no args
	if err := onCmd.ValidateArgs([]string{}); err != nil {
		t.Errorf("On subcommand should accept no arguments: %v", err)
	}
}

// TestClimateCommand_OffSubcommand tests climate off subcommand
func TestClimateCommand_OffSubcommand(t *testing.T) {
	cmd := NewClimateCmd()

	// Find off subcommand
	var offCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "off" {
			offCmd = subCmd
			break
		}
	}

	if offCmd == nil {
		t.Fatal("Expected off subcommand to exist")
	}

	if offCmd.Short == "" {
		t.Error("Expected off subcommand to have a description")
	}

	// Should accept no args
	if err := offCmd.ValidateArgs([]string{}); err != nil {
		t.Errorf("Off subcommand should accept no arguments: %v", err)
	}
}

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

// TestClimateCommand_SetSubcommand tests climate set subcommand
func TestClimateCommand_SetSubcommand(t *testing.T) {
	cmd := NewClimateCmd()

	// Find set subcommand
	var setCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "set" {
			setCmd = subCmd
			break
		}
	}

	if setCmd == nil {
		t.Fatal("Expected set subcommand to exist")
	}

	if setCmd.Short == "" {
		t.Error("Expected set subcommand to have a description")
	}
}

// TestClimateCommand_SetSubcommand_Flags tests climate set subcommand flags
func TestClimateCommand_SetSubcommand_Flags(t *testing.T) {
	cmd := NewClimateCmd()

	var setCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "set" {
			setCmd = subCmd
			break
		}
	}

	if setCmd == nil {
		t.Fatal("Expected set subcommand to exist")
	}

	// Test that flags exist
	tempFlag := setCmd.Flags().Lookup("temp")
	if tempFlag == nil {
		t.Error("Expected --temp flag to exist")
	}

	unitFlag := setCmd.Flags().Lookup("unit")
	if unitFlag == nil {
		t.Error("Expected --unit flag to exist")
	}
	if unitFlag != nil && unitFlag.DefValue != "c" {
		t.Errorf("Expected --unit default to be 'c', got '%s'", unitFlag.DefValue)
	}

	frontDefrostFlag := setCmd.Flags().Lookup("front-defrost")
	if frontDefrostFlag == nil {
		t.Error("Expected --front-defrost flag to exist")
	}

	rearDefrostFlag := setCmd.Flags().Lookup("rear-defrost")
	if rearDefrostFlag == nil {
		t.Error("Expected --rear-defrost flag to exist")
	}
}

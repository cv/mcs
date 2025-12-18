package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

// TestRawCommand tests the raw command
func TestRawCommand(t *testing.T) {
	cmd := NewRawCmd()

	if cmd.Use != "raw" {
		t.Errorf("Expected Use to be 'raw', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}

// TestRawCommand_StatusSubcommand tests raw status subcommand
func TestRawCommand_StatusSubcommand(t *testing.T) {
	cmd := NewRawCmd()

	// Find status subcommand
	var statusCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "status" {
			statusCmd = subCmd
			break
		}
	}

	if statusCmd == nil {
		t.Fatal("Expected status subcommand to exist")
	}

	if statusCmd.Short == "" {
		t.Error("Expected status subcommand to have a description")
	}

	// Should accept no args
	if err := statusCmd.ValidateArgs([]string{}); err != nil {
		t.Errorf("Status subcommand should accept no arguments: %v", err)
	}
}

// TestRawCommand_EVSubcommand tests raw ev subcommand
func TestRawCommand_EVSubcommand(t *testing.T) {
	cmd := NewRawCmd()

	// Find ev subcommand
	var evCmd *cobra.Command
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == "ev" {
			evCmd = subCmd
			break
		}
	}

	if evCmd == nil {
		t.Fatal("Expected ev subcommand to exist")
	}

	if evCmd.Short == "" {
		t.Error("Expected ev subcommand to have a description")
	}

	// Should accept no args
	if err := evCmd.ValidateArgs([]string{}); err != nil {
		t.Errorf("EV subcommand should accept no arguments: %v", err)
	}
}

package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

// findSubcommand finds a subcommand by name in the given parent command.
// Returns nil if not found.
func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == name {
			return subCmd
		}
	}
	return nil
}

// assertCommandBasics tests that a command has the expected Use field
// and a non-empty Short description. This helper eliminates duplicated
// test patterns across command test files.
func assertCommandBasics(t *testing.T, cmd *cobra.Command, expectedUse string) {
	t.Helper()

	if cmd.Use != expectedUse {
		t.Errorf("Expected Use to be '%s', got '%s'", expectedUse, cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}

// assertNoArgsCommand tests that a command accepts no arguments.
func assertNoArgsCommand(t *testing.T, cmd *cobra.Command) {
	t.Helper()

	if err := cmd.ValidateArgs([]string{}); err != nil {
		t.Errorf("%s command should accept no arguments: %v", cmd.Use, err)
	}
}

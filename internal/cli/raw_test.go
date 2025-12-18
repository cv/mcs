package cli

import "testing"

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

// TestRawCommand_Subcommands tests raw subcommands
func TestRawCommand_Subcommands(t *testing.T) {
	subcommands := []string{"status", "ev"}

	for _, name := range subcommands {
		t.Run(name, func(t *testing.T) {
			cmd := NewRawCmd()
			subCmd := findSubcommand(cmd, name)

			if subCmd == nil {
				t.Fatalf("Expected %s subcommand to exist", name)
			}

			if subCmd.Short == "" {
				t.Errorf("Expected %s subcommand to have a description", name)
			}

			if err := subCmd.ValidateArgs([]string{}); err != nil {
				t.Errorf("%s subcommand should accept no arguments: %v", name, err)
			}
		})
	}
}

package cli

import "testing"

// TestChargeCommand tests the charge command
func TestChargeCommand(t *testing.T) {
	cmd := NewChargeCmd()
	assertCommandBasics(t, cmd, "charge")
}

// TestChargeCommand_Subcommands tests charge subcommands
func TestChargeCommand_Subcommands(t *testing.T) {
	subcommands := []string{"start", "stop"}

	for _, name := range subcommands {
		t.Run(name, func(t *testing.T) {
			cmd := NewChargeCmd()
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

package cli

import "testing"

// TestChargeCommand tests the charge command
func TestChargeCommand(t *testing.T) {
	cmd := NewChargeCmd()
	assertCommandBasics(t, cmd, "charge")
}

// TestChargeCommand_Subcommands tests charge subcommands
func TestChargeCommand_Subcommands(t *testing.T) {
	cmd := NewChargeCmd()
	assertSubcommandsExist(t, cmd, []string{"start", "stop"}, true)
}

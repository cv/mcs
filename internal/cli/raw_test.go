package cli

import "testing"

// TestRawCommand tests the raw command
func TestRawCommand(t *testing.T) {
	cmd := NewRawCmd()
	assertCommandBasics(t, cmd, "raw")
}

// TestRawCommand_Subcommands tests raw subcommands
func TestRawCommand_Subcommands(t *testing.T) {
	cmd := NewRawCmd()
	assertSubcommandsExist(t, cmd, []string{"status", "ev"}, true)
}

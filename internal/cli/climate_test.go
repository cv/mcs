package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestClimateCommand tests the climate command
func TestClimateCommand(t *testing.T) {
	cmd := NewClimateCmd()
	assertCommandBasics(t, cmd, "climate")
}

// TestClimateCommand_Subcommands tests climate subcommands
func TestClimateCommand_Subcommands(t *testing.T) {
	cmd := NewClimateCmd()

	// Test on/off subcommands (accept no args)
	assertSubcommandsExist(t, cmd, []string{"on", "off"}, true)

	// Test set subcommand (accepts flags, so no args validation)
	t.Run("set", func(t *testing.T) {
		assertSubcommandExists(t, cmd, "set", false)
	})
}

// TestClimateCommand_SetSubcommand_Flags tests climate set subcommand flags
func TestClimateCommand_SetSubcommand_Flags(t *testing.T) {
	cmd := NewClimateCmd()
	setCmd := findSubcommand(cmd, "set")

	require.NotNil(t, setCmd, "Expected set subcommand to exist")

	// Test that flags exist with proper defaults
	assertFlagExists(t, setCmd, FlagAssertion{Name: "temp"})
	assertFlagExists(t, setCmd, FlagAssertion{Name: "unit", DefaultValue: "c"})
	assertFlagExists(t, setCmd, FlagAssertion{Name: "front-defrost"})
	assertFlagExists(t, setCmd, FlagAssertion{Name: "rear-defrost"})
}

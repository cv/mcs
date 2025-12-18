package cli

import "testing"

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

// TestClimateCommand_Subcommands tests climate subcommands
func TestClimateCommand_Subcommands(t *testing.T) {
	subcommands := []string{"on", "off", "set"}

	for _, name := range subcommands {
		t.Run(name, func(t *testing.T) {
			cmd := NewClimateCmd()
			subCmd := findSubcommand(cmd, name)

			if subCmd == nil {
				t.Fatalf("Expected %s subcommand to exist", name)
			}

			if subCmd.Short == "" {
				t.Errorf("Expected %s subcommand to have a description", name)
			}
		})
	}
}

// TestClimateCommand_SetSubcommand_Flags tests climate set subcommand flags
func TestClimateCommand_SetSubcommand_Flags(t *testing.T) {
	cmd := NewClimateCmd()
	setCmd := findSubcommand(cmd, "set")

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

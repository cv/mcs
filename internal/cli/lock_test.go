package cli

import (
	"testing"
)

// TestLockCommand tests the lock command
func TestLockCommand(t *testing.T) {
	cmd := NewLockCmd()

	if cmd.Use != "lock" {
		t.Errorf("Expected Use to be 'lock', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}

// TestLockCommand_NoArgs tests lock command with no arguments
func TestLockCommand_NoArgs(t *testing.T) {
	cmd := NewLockCmd()

	// Should accept no args
	if err := cmd.ValidateArgs([]string{}); err != nil {
		t.Errorf("Lock command should accept no arguments: %v", err)
	}
}

// TestUnlockCommand tests the unlock command
func TestUnlockCommand(t *testing.T) {
	cmd := NewUnlockCmd()

	if cmd.Use != "unlock" {
		t.Errorf("Expected Use to be 'unlock', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}

// TestUnlockCommand_NoArgs tests unlock command with no arguments
func TestUnlockCommand_NoArgs(t *testing.T) {
	cmd := NewUnlockCmd()

	// Should accept no args
	if err := cmd.ValidateArgs([]string{}); err != nil {
		t.Errorf("Unlock command should accept no arguments: %v", err)
	}
}

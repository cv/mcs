package cli

import (
	"testing"
)

// TestLockCommand tests the lock command
func TestLockCommand(t *testing.T) {
	cmd := NewLockCmd()
	assertCommandBasics(t, cmd, "lock")
}

// TestLockCommand_NoArgs tests lock command with no arguments
func TestLockCommand_NoArgs(t *testing.T) {
	cmd := NewLockCmd()
	assertNoArgsCommand(t, cmd)
}

// TestUnlockCommand tests the unlock command
func TestUnlockCommand(t *testing.T) {
	cmd := NewUnlockCmd()
	assertCommandBasics(t, cmd, "unlock")
}

// TestUnlockCommand_NoArgs tests unlock command with no arguments
func TestUnlockCommand_NoArgs(t *testing.T) {
	cmd := NewUnlockCmd()
	assertNoArgsCommand(t, cmd)
}

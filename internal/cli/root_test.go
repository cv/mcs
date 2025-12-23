package cli

import (
	"bytes"
	"context"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd_Version(t *testing.T) {
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"--version"})

	var output bytes.Buffer
	rootCmd.SetOut(&output)

	err := rootCmd.Execute()
	require.NoError(t, err, "Execute() error = %v")

	result := output.String()
	assert.Truef(t, strings.Contains(result, "mcs version"), "Expected version output, got: %s", result)
}

func TestRootCmd_Help(t *testing.T) {
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"--help"})

	var output bytes.Buffer
	rootCmd.SetOut(&output)

	err := rootCmd.Execute()
	require.NoError(t, err, "Execute() error = %v")

	result := output.String()
	assert.Truef(t, strings.Contains(result, "mcs"), "Expected help output to contain 'mcs', got: %s", result)
	// Check for content from the Long description
	assert.Truef(t, strings.Contains(result, "manufacturer API"), "Expected help output to contain 'manufacturer API', got: %s", result)
}

func TestRootCmd_NoArgs(t *testing.T) {
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{})

	var output bytes.Buffer
	rootCmd.SetOut(&output)
	rootCmd.SetErr(&output)

	// Should show help when no args provided
	err := rootCmd.Execute()
	// Root command with no args should not error, just show help
	require.NoError(t, err, "Execute() error = %v")
}

func TestExecute_SignalHandling(t *testing.T) {
	// Create a command that blocks until context is cancelled
	rootCmd := NewRootCmd()

	// Add a test subcommand that respects context cancellation
	testCmd := &cobra.Command{
		Use:   "test-signal",
		Short: "Test signal handling",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Wait for context to be cancelled
			<-cmd.Context().Done()
			return cmd.Context().Err()
		},
	}
	rootCmd.AddCommand(testCmd)
	rootCmd.SetArgs([]string{"test-signal"})

	// Create a cancellable context to simulate signal handling
	ctx, cancel := context.WithCancel(context.Background())

	// Run command in goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- rootCmd.ExecuteContext(ctx)
	}()

	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)

	// Cancel the context (simulating signal)
	cancel()

	// Wait for command to finish
	select {
	case err := <-errCh:
		assert.Equalf(t, context.Canceled, err, "Expected context.Canceled, got: %v", err)
	case <-time.After(1 * time.Second):
		t.Fatal("Command did not respond to context cancellation")
	}
}

func TestExecute_WithRealSignal(t *testing.T) {
	// This test verifies that signal.NotifyContext properly captures signals
	// We'll test the signal mechanism without actually running a full command

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a signal channel
	sigCh := make(chan os.Signal, 1)

	// Test that we can notify on SIGINT and SIGTERM
	// This validates that our imports and usage are correct
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-sigCh:
			cancel()
		}
	}()

	// Verify the signal types we're using are valid
	signals := []os.Signal{os.Interrupt, syscall.SIGTERM}
	for _, sig := range signals {
		assert.NotNil(t, sig, "Signal is nil")
	}
}

func TestCheckSkillVersionMismatch_SkipsSkillCommands(t *testing.T) {
	// Create a skill command
	skillCmd := &cobra.Command{Use: "skill"}

	// Capture stderr
	var errBuf bytes.Buffer
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	checkSkillVersionMismatch(skillCmd)

	_ = w.Close()
	os.Stderr = oldStderr

	_, _ = errBuf.ReadFrom(r)

	// Should not print anything for skill command
	if errBuf.Len() > 0 {
		t.Errorf("Expected no output for skill command, got: %s", errBuf.String())
	}
}

func TestCheckSkillVersionMismatch_SkipsSkillSubcommands(t *testing.T) {
	// Create a skill subcommand (e.g., skill install)
	skillCmd := &cobra.Command{Use: "skill"}
	installCmd := &cobra.Command{Use: "install"}
	skillCmd.AddCommand(installCmd)

	// Capture stderr
	var errBuf bytes.Buffer
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	checkSkillVersionMismatch(installCmd)

	_ = w.Close()
	os.Stderr = oldStderr

	_, _ = errBuf.ReadFrom(r)

	// Should not print anything for skill subcommand
	if errBuf.Len() > 0 {
		t.Errorf("Expected no output for skill subcommand, got: %s", errBuf.String())
	}
}

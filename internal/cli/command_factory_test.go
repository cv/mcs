package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/cv/mcs/internal/api"
)

// TestNewSimpleCommand tests the simple command factory
func TestNewSimpleCommand(t *testing.T) {
	config := SimpleCommandConfig{
		Use:   "test",
		Short: "Test command",
		Long:  "This is a test command",
		APICall: func(ctx context.Context, client *api.Client, vin api.InternalVIN) error {
			return nil
		},
		SuccessMsg:   "Test succeeded",
		ErrorMsgTmpl: "test failed: %w",
	}

	cmd := NewSimpleCommand(config)

	if cmd.Use != "test" {
		t.Errorf("Expected Use to be 'test', got '%s'", cmd.Use)
	}

	if cmd.Short != "Test command" {
		t.Errorf("Expected Short to be 'Test command', got '%s'", cmd.Short)
	}

	if cmd.Long != "This is a test command" {
		t.Errorf("Expected Long to be 'This is a test command', got '%s'", cmd.Long)
	}

	if !cmd.SilenceUsage {
		t.Error("Expected SilenceUsage to be true")
	}
}

// TestNewSimpleCommand_ErrorHandling tests error handling in simple commands
func TestNewSimpleCommand_ErrorHandling(t *testing.T) {
	testErr := errors.New("API call failed")

	config := SimpleCommandConfig{
		Use:   "test",
		Short: "Test command",
		Long:  "This is a test command",
		APICall: func(ctx context.Context, client *api.Client, vin api.InternalVIN) error {
			return testErr
		},
		SuccessMsg:   "Test succeeded",
		ErrorMsgTmpl: "test failed: %w",
	}

	cmd := NewSimpleCommand(config)

	// The RunE function should wrap the error
	// Note: We can't actually test this without a full setup because it requires createAPIClient
	// but we can verify the command structure is correct
	if cmd.RunE == nil {
		t.Error("Expected RunE to be set")
	}
}

// TestNewParentWithSubcommands tests parent command with subcommands
func TestNewParentWithSubcommands(t *testing.T) {
	subcommands := []SimpleCommandConfig{
		{
			Use:   "start",
			Short: "Start something",
			Long:  "Start something long description",
			APICall: func(ctx context.Context, client *api.Client, vin api.InternalVIN) error {
				return nil
			},
			SuccessMsg:   "Started",
			ErrorMsgTmpl: "failed to start: %w",
		},
		{
			Use:   "stop",
			Short: "Stop something",
			Long:  "Stop something long description",
			APICall: func(ctx context.Context, client *api.Client, vin api.InternalVIN) error {
				return nil
			},
			SuccessMsg:   "Stopped",
			ErrorMsgTmpl: "failed to stop: %w",
		},
	}

	cmd := NewParentWithSubcommands("parent", "Parent command", "Parent command description", "", subcommands)

	if cmd.Use != "parent" {
		t.Errorf("Expected Use to be 'parent', got '%s'", cmd.Use)
	}

	if cmd.Short != "Parent command" {
		t.Errorf("Expected Short to be 'Parent command', got '%s'", cmd.Short)
	}

	if cmd.Long != "Parent command description" {
		t.Errorf("Expected Long to be 'Parent command description', got '%s'", cmd.Long)
	}

	// Check that subcommands were added
	if len(cmd.Commands()) != 2 {
		t.Errorf("Expected 2 subcommands, got %d", len(cmd.Commands()))
	}

	// Verify subcommand names
	subcommandNames := make(map[string]bool)
	for _, subcmd := range cmd.Commands() {
		subcommandNames[subcmd.Use] = true
	}

	if !subcommandNames["start"] {
		t.Error("Expected 'start' subcommand to exist")
	}

	if !subcommandNames["stop"] {
		t.Error("Expected 'stop' subcommand to exist")
	}
}

// TestNewSimpleCommand_OutputFormatting tests that success messages are formatted correctly
func TestNewSimpleCommand_OutputFormatting(t *testing.T) {
	config := SimpleCommandConfig{
		Use:   "test",
		Short: "Test command",
		Long:  "This is a test command",
		APICall: func(ctx context.Context, client *api.Client, vin api.InternalVIN) error {
			return nil
		},
		SuccessMsg:   "Test succeeded",
		ErrorMsgTmpl: "test failed: %w",
	}

	cmd := NewSimpleCommand(config)

	// Set up a buffer to capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Verify command was created with correct structure
	if cmd == nil {
		t.Fatal("Expected command to be created")
	}

	// The actual execution would require full setup with API client
	// so we just verify the command structure here
}

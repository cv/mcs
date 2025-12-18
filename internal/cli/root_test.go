package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCmd_Version(t *testing.T) {
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"--version"})

	var output bytes.Buffer
	rootCmd.SetOut(&output)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "mcs version") {
		t.Errorf("Expected version output, got: %s", result)
	}
}

func TestRootCmd_Help(t *testing.T) {
	rootCmd := NewRootCmd()
	rootCmd.SetArgs([]string{"--help"})

	var output bytes.Buffer
	rootCmd.SetOut(&output)

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "mcs") {
		t.Errorf("Expected help output to contain 'mcs', got: %s", result)
	}
	// Check for content from the Long description
	if !strings.Contains(result, "manufacturer API") {
		t.Errorf("Expected help output to contain 'manufacturer API', got: %s", result)
	}
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
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

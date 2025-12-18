package cli

import "github.com/spf13/cobra"

// findSubcommand finds a subcommand by name in the given parent command.
// Returns nil if not found.
func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, subCmd := range cmd.Commands() {
		if subCmd.Use == name {
			return subCmd
		}
	}
	return nil
}

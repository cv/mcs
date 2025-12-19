package cli

import (
	"context"
	"fmt"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// SimpleCommandConfig defines the configuration for a simple vehicle command.
// A simple command is one that takes no arguments, calls a single API method,
// and displays a success message.
type SimpleCommandConfig struct {
	Use          string                                                          // Command name (e.g., "lock", "start")
	Short        string                                                          // Short description
	Long         string                                                          // Long description
	APICall      func(ctx context.Context, client *api.Client, vin string) error // API method to invoke
	SuccessMsg   string                                                          // Message to display on success
	ErrorMsgTmpl string                                                          // Error message template (uses fmt.Errorf)
}

// NewSimpleCommand creates a cobra command from a SimpleCommandConfig.
// This factory reduces boilerplate for commands that follow the pattern:
//  1. No arguments
//  2. Create API client
//  3. Get vehicle VIN
//  4. Call single API method
//  5. Display success message
func NewSimpleCommand(cfg SimpleCommandConfig) *cobra.Command {
	return &cobra.Command{
		Use:   cfg.Use,
		Short: cfg.Short,
		Long:  cfg.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
				if err := cfg.APICall(ctx, client, internalVIN); err != nil {
					return fmt.Errorf(cfg.ErrorMsgTmpl, err)
				}
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), cfg.SuccessMsg)
				return nil
			})
		},
		SilenceUsage: true,
	}
}

// NewParentWithSubcommands creates a parent command with simple subcommands.
// This is useful for commands like "charge" or "climate" that have multiple actions.
func NewParentWithSubcommands(use, short, long string, subcommands []SimpleCommandConfig) *cobra.Command {
	parentCmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
	}

	for _, subcfg := range subcommands {
		parentCmd.AddCommand(NewSimpleCommand(subcfg))
	}

	return parentCmd
}

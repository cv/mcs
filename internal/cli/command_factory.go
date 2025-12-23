package cli

import (
	"context"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// CommandSpec holds the specification for building a confirmable command.
type CommandSpec struct {
	// Command metadata
	Use     string
	Short   string
	Long    string
	Example string

	// Flag configuration
	ConfirmFlagUsage   string // e.g., "wait for confirmation that doors are locked"
	ConfirmWaitDefault int    // Default timeout in seconds (use 90 if not specified)

	// Command configuration
	Config ConfirmableCommandConfig
}

// buildConfirmableCommand creates a cobra command from a CommandSpec.
// This eliminates the boilerplate of creating commands with confirm/confirm-wait flags.
func buildConfirmableCommand(spec CommandSpec) *cobra.Command {
	var confirm bool
	var confirmWait int

	// Set default confirm wait if not specified
	if spec.ConfirmWaitDefault == 0 {
		spec.ConfirmWaitDefault = 90
	}

	cmd := &cobra.Command{
		Use:     spec.Use,
		Short:   spec.Short,
		Long:    spec.Long,
		Example: spec.Example,
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
				return executeConfirmableCommand(ctx, cmd.OutOrStdout(), client, internalVIN, spec.Config, confirm, confirmWait)
			})
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&confirm, "confirm", true, spec.ConfirmFlagUsage)
	cmd.Flags().IntVar(&confirmWait, "confirm-wait", spec.ConfirmWaitDefault, "max seconds to wait for confirmation")

	return cmd
}

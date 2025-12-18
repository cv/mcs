package cli

import (
	"context"
	"fmt"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewClimateCmd creates the climate command
func NewClimateCmd() *cobra.Command {
	climateCmd := &cobra.Command{
		Use:   "climate",
		Short: "Control vehicle climate (HVAC)",
		Long:  `Control vehicle climate system (on/off).`,
	}

	// Add subcommands
	climateCmd.AddCommand(&cobra.Command{
		Use:   "on",
		Short: "Turn climate on",
		Long:  `Turn the vehicle HVAC system on.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClimateOn(cmd)
		},
		SilenceUsage: true,
	})

	climateCmd.AddCommand(&cobra.Command{
		Use:   "off",
		Short: "Turn climate off",
		Long:  `Turn the vehicle HVAC system off.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClimateOff(cmd)
		},
		SilenceUsage: true,
	})

	return climateCmd
}

// runClimateOn executes the climate on command
func runClimateOn(cmd *cobra.Command) error {
	return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
		if err := client.HVACOn(ctx, internalVIN); err != nil {
			return fmt.Errorf("failed to turn HVAC on: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Climate turned on successfully")
		return nil
	})
}

// runClimateOff executes the climate off command
func runClimateOff(cmd *cobra.Command) error {
	return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
		if err := client.HVACOff(ctx, internalVIN); err != nil {
			return fmt.Errorf("failed to turn HVAC off: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Climate turned off successfully")
		return nil
	})
}

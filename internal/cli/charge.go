package cli

import (
	"fmt"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewChargeCmd creates the charge command
func NewChargeCmd() *cobra.Command {
	chargeCmd := &cobra.Command{
		Use:   "charge",
		Short: "Control vehicle charging",
		Long:  `Control vehicle charging (start/stop).`,
	}

	// Add subcommands
	chargeCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start charging",
		Long:  `Start charging the vehicle battery.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChargeStart(cmd)
		},
		SilenceUsage: true,
	})

	chargeCmd.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "Stop charging",
		Long:  `Stop charging the vehicle battery.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChargeStop(cmd)
		},
		SilenceUsage: true,
	})

	return chargeCmd
}

// runChargeStart executes the charge start command
func runChargeStart(cmd *cobra.Command) error {
	return withVehicleClient(func(client *api.Client, internalVIN string) error {
		if err := client.ChargeStart(internalVIN); err != nil {
			return fmt.Errorf("failed to start charging: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Charging started successfully")
		return nil
	})
}

// runChargeStop executes the charge stop command
func runChargeStop(cmd *cobra.Command) error {
	return withVehicleClient(func(client *api.Client, internalVIN string) error {
		if err := client.ChargeStop(internalVIN); err != nil {
			return fmt.Errorf("failed to stop charging: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Charging stopped successfully")
		return nil
	})
}

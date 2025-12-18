package cli

import (
	"fmt"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewStartCmd creates the start command
func NewStartCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start vehicle engine",
		Long:  `Start the vehicle engine remotely.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(cmd)
		},
		SilenceUsage: true,
	}

	return startCmd
}

// NewStopCmd creates the stop command
func NewStopCmd() *cobra.Command {
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop vehicle engine",
		Long:  `Stop the vehicle engine remotely.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop(cmd)
		},
		SilenceUsage: true,
	}

	return stopCmd
}

// runStart executes the start command
func runStart(cmd *cobra.Command) error {
	return withVehicleClient(func(client *api.Client, internalVIN string) error {
		if err := client.EngineStart(internalVIN); err != nil {
			return fmt.Errorf("failed to start engine: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Engine started successfully")
		return nil
	})
}

// runStop executes the stop command
func runStop(cmd *cobra.Command) error {
	return withVehicleClient(func(client *api.Client, internalVIN string) error {
		if err := client.EngineStop(internalVIN); err != nil {
			return fmt.Errorf("failed to stop engine: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Engine stopped successfully")
		return nil
	})
}

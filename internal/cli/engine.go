package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewStartCmd creates the start command
func NewStartCmd() *cobra.Command {
	var confirm bool
	var confirmWait int

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start vehicle engine",
		Long:  `Start the vehicle engine remotely.`,
		Example: `  # Start the vehicle engine remotely
  mcs start

  # Expected output on success:
  # Engine started successfully

  # Start engine without waiting for confirmation
  mcs start --confirm=false

  # Start engine and wait up to 60 seconds for confirmation
  mcs start --confirm-wait 60`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
				// Send start command
				if err := client.EngineStart(ctx, internalVIN); err != nil {
					return fmt.Errorf("failed to start engine: %w", err)
				}

				// If confirmation disabled, return immediately
				if !confirm {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Engine started successfully")
					return nil
				}

				// Wait for confirmation
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Start command sent, waiting for confirmation...")
				result := waitForEngineRunning(
					ctx,
					cmd.OutOrStdout(),
					client,
					internalVIN,
					time.Duration(confirmWait)*time.Second,
					5*time.Second, // poll every 5 seconds
				)

				if result.err != nil {
					return fmt.Errorf("failed to confirm engine status: %w", result.err)
				}

				if result.success {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Engine started successfully")
				} else {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Start command sent (confirmation timeout)")
				}

				return nil
			})
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&confirm, "confirm", true, "wait for confirmation that engine is running")
	cmd.Flags().IntVar(&confirmWait, "confirm-wait", 90, "max seconds to wait for confirmation")

	return cmd
}

// NewStopCmd creates the stop command
func NewStopCmd() *cobra.Command {
	var confirm bool
	var confirmWait int

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop vehicle engine",
		Long:  `Stop the vehicle engine remotely.`,
		Example: `  # Stop the vehicle engine remotely
  mcs stop

  # Expected output on success:
  # Engine stopped successfully

  # Stop engine without waiting for confirmation
  mcs stop --confirm=false

  # Stop engine and wait up to 60 seconds for confirmation
  mcs stop --confirm-wait 60`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
				// Send stop command
				if err := client.EngineStop(ctx, internalVIN); err != nil {
					return fmt.Errorf("failed to stop engine: %w", err)
				}

				// If confirmation disabled, return immediately
				if !confirm {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Engine stopped successfully")
					return nil
				}

				// Wait for confirmation
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Stop command sent, waiting for confirmation...")
				result := waitForEngineStopped(
					ctx,
					cmd.OutOrStdout(),
					client,
					internalVIN,
					time.Duration(confirmWait)*time.Second,
					5*time.Second, // poll every 5 seconds
				)

				if result.err != nil {
					return fmt.Errorf("failed to confirm engine status: %w", result.err)
				}

				if result.success {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Engine stopped successfully")
				} else {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Stop command sent (confirmation timeout)")
				}

				return nil
			})
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&confirm, "confirm", true, "wait for confirmation that engine is stopped")
	cmd.Flags().IntVar(&confirmWait, "confirm-wait", 90, "max seconds to wait for confirmation")

	return cmd
}

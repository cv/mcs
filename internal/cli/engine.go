package cli

import (
	"context"
	"io"
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
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
				config := ConfirmableCommandConfig{
					ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
						return client.EngineStart(ctx, string(internalVIN))
					},
					WaitFunc: func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
						return waitForEngineRunning(ctx, out, &clientAdapter{Client: client}, internalVIN, timeout, pollInterval)
					},
					SuccessMsg:    "Engine started successfully",
					WaitingMsg:    "Start command sent, waiting for confirmation...",
					ActionName:    "start engine",
					ConfirmName:   "engine status",
					TimeoutSuffix: "confirmation timeout",
				}
				return executeConfirmableCommand(ctx, cmd.OutOrStdout(), client, internalVIN, config, confirm, confirmWait)
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
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
				config := ConfirmableCommandConfig{
					ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
						return client.EngineStop(ctx, string(internalVIN))
					},
					WaitFunc: func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
						return waitForEngineStopped(ctx, out, &clientAdapter{Client: client}, internalVIN, timeout, pollInterval)
					},
					SuccessMsg:    "Engine stopped successfully",
					WaitingMsg:    "Stop command sent, waiting for confirmation...",
					ActionName:    "stop engine",
					ConfirmName:   "engine status",
					TimeoutSuffix: "confirmation timeout",
				}
				return executeConfirmableCommand(ctx, cmd.OutOrStdout(), client, internalVIN, config, confirm, confirmWait)
			})
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&confirm, "confirm", true, "wait for confirmation that engine is stopped")
	cmd.Flags().IntVar(&confirmWait, "confirm-wait", 90, "max seconds to wait for confirmation")

	return cmd
}

package cli

import (
	"context"
	"io"
	"time"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewChargeCmd creates the charge command
func NewChargeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "charge",
		Short: "Control vehicle charging",
		Long:  `Control vehicle charging (start/stop).`,
		Example: `  # Start charging the vehicle battery
  mcs charge start

  # Stop charging the vehicle battery
  mcs charge stop`,
	}

	cmd.AddCommand(NewChargeStartCmd())
	cmd.AddCommand(NewChargeStopCmd())

	return cmd
}

// NewChargeStartCmd creates the charge start subcommand
func NewChargeStartCmd() *cobra.Command {
	var confirm bool
	var confirmWait int

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start charging",
		Long:  `Start charging the vehicle battery.`,
		Example: `  # Start charging the vehicle battery
  mcs charge start

  # Expected output on success:
  # Charging started successfully

  # Start charging without waiting for confirmation
  mcs charge start --confirm=false

  # Start charging and wait up to 60 seconds for confirmation
  mcs charge start --confirm-wait 60`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
				config := ConfirmableCommandConfig{
					ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
						return client.ChargeStart(ctx, string(internalVIN))
					},
					WaitFunc: func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
						return waitForCharging(ctx, out, &clientAdapter{Client: client}, internalVIN, timeout, pollInterval)
					},
					InitialDelay:  ConfirmationInitialDelay,
					SuccessMsg:    "Charging started successfully",
					WaitingMsg:    "Charge start command sent, waiting for confirmation...",
					ActionName:    "start charging",
					ConfirmName:   "charging status",
					TimeoutSuffix: "confirmation timeout",
				}
				return executeConfirmableCommand(ctx, cmd.OutOrStdout(), client, internalVIN, config, confirm, confirmWait)
			})
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&confirm, "confirm", true, "wait for confirmation that charging has started")
	cmd.Flags().IntVar(&confirmWait, "confirm-wait", 90, "max seconds to wait for confirmation")

	return cmd
}

// NewChargeStopCmd creates the charge stop subcommand
func NewChargeStopCmd() *cobra.Command {
	var confirm bool
	var confirmWait int

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop charging",
		Long:  `Stop charging the vehicle battery.`,
		Example: `  # Stop charging the vehicle battery
  mcs charge stop

  # Expected output on success:
  # Charging stopped successfully

  # Stop charging without waiting for confirmation
  mcs charge stop --confirm=false

  # Stop charging and wait up to 60 seconds for confirmation
  mcs charge stop --confirm-wait 60`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
				config := ConfirmableCommandConfig{
					ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
						return client.ChargeStop(ctx, string(internalVIN))
					},
					WaitFunc: func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
						return waitForNotCharging(ctx, out, &clientAdapter{Client: client}, internalVIN, timeout, pollInterval)
					},
					InitialDelay:  ConfirmationInitialDelay,
					SuccessMsg:    "Charging stopped successfully",
					WaitingMsg:    "Charge stop command sent, waiting for confirmation...",
					ActionName:    "stop charging",
					ConfirmName:   "charging status",
					TimeoutSuffix: "confirmation timeout",
				}
				return executeConfirmableCommand(ctx, cmd.OutOrStdout(), client, internalVIN, config, confirm, confirmWait)
			})
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&confirm, "confirm", true, "wait for confirmation that charging has stopped")
	cmd.Flags().IntVar(&confirmWait, "confirm-wait", 90, "max seconds to wait for confirmation")

	return cmd
}

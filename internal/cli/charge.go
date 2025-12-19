package cli

import (
	"context"
	"fmt"
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
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
				// Send start command
				if err := client.ChargeStart(ctx, internalVIN); err != nil {
					return fmt.Errorf("failed to start charging: %w", err)
				}

				// If confirmation disabled, return immediately
				if !confirm {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Charging started successfully")
					return nil
				}

				// Wait for confirmation
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Charge start command sent, waiting for confirmation...")
				result := waitForCharging(
					ctx,
					cmd.OutOrStdout(),
					client,
					internalVIN,
					time.Duration(confirmWait)*time.Second,
					5*time.Second, // poll every 5 seconds
				)

				if result.err != nil {
					return fmt.Errorf("failed to confirm charging status: %w", result.err)
				}

				if result.success {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Charging started successfully")
				} else {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Charge start command sent (confirmation timeout)")
				}

				return nil
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
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
				// Send stop command
				if err := client.ChargeStop(ctx, internalVIN); err != nil {
					return fmt.Errorf("failed to stop charging: %w", err)
				}

				// If confirmation disabled, return immediately
				if !confirm {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Charging stopped successfully")
					return nil
				}

				// Wait for confirmation
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Charge stop command sent, waiting for confirmation...")
				result := waitForNotCharging(
					ctx,
					cmd.OutOrStdout(),
					client,
					internalVIN,
					time.Duration(confirmWait)*time.Second,
					5*time.Second, // poll every 5 seconds
				)

				if result.err != nil {
					return fmt.Errorf("failed to confirm charging status: %w", result.err)
				}

				if result.success {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Charging stopped successfully")
				} else {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Charge stop command sent (confirmation timeout)")
				}

				return nil
			})
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&confirm, "confirm", true, "wait for confirmation that charging has stopped")
	cmd.Flags().IntVar(&confirmWait, "confirm-wait", 90, "max seconds to wait for confirmation")

	return cmd
}

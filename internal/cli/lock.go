package cli

import (
	"context"
	"io"
	"time"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewLockCmd creates the lock command
func NewLockCmd() *cobra.Command {
	var confirm bool
	var confirmWait int

	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Lock vehicle doors",
		Long:  `Lock all vehicle doors remotely.`,
		Example: `  # Lock all doors on your vehicle
  mcs lock

  # Expected output on success:
  # Doors locked successfully

  # Lock doors without waiting for confirmation
  mcs lock --confirm=false

  # Lock doors and wait up to 60 seconds for confirmation
  mcs lock --confirm-wait 60`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
				config := ConfirmableCommandConfig{
					ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
						return client.DoorLock(ctx, string(internalVIN))
					},
					WaitFunc: func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
						return waitForDoorsLocked(ctx, out, &clientAdapter{Client: client}, internalVIN, timeout, pollInterval)
					},
					InitialDelay:  ConfirmationInitialDelay,
					SuccessMsg:    "Doors locked successfully",
					WaitingMsg:    "Lock command sent, waiting for confirmation...",
					ActionName:    "lock doors",
					ConfirmName:   "lock status",
					TimeoutSuffix: "confirmation timeout",
				}
				return executeConfirmableCommand(ctx, cmd.OutOrStdout(), client, internalVIN, config, confirm, confirmWait)
			})
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&confirm, "confirm", true, "wait for confirmation that doors are locked")
	cmd.Flags().IntVar(&confirmWait, "confirm-wait", 90, "max seconds to wait for confirmation")

	return cmd
}

// NewUnlockCmd creates the unlock command
func NewUnlockCmd() *cobra.Command {
	var confirm bool
	var confirmWait int

	cmd := &cobra.Command{
		Use:   "unlock",
		Short: "Unlock vehicle doors",
		Long:  `Unlock all vehicle doors remotely.`,
		Example: `  # Unlock all doors on your vehicle
  mcs unlock

  # Expected output on success:
  # Doors unlocked successfully

  # Unlock doors without waiting for confirmation
  mcs unlock --confirm=false

  # Unlock doors and wait up to 60 seconds for confirmation
  mcs unlock --confirm-wait 60`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
				config := ConfirmableCommandConfig{
					ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
						return client.DoorUnlock(ctx, string(internalVIN))
					},
					WaitFunc: func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
						return waitForDoorsUnlocked(ctx, out, &clientAdapter{Client: client}, internalVIN, timeout, pollInterval)
					},
					InitialDelay:  ConfirmationInitialDelay,
					SuccessMsg:    "Doors unlocked successfully",
					WaitingMsg:    "Unlock command sent, waiting for confirmation...",
					ActionName:    "unlock doors",
					ConfirmName:   "unlock status",
					TimeoutSuffix: "confirmation timeout",
				}
				return executeConfirmableCommand(ctx, cmd.OutOrStdout(), client, internalVIN, config, confirm, confirmWait)
			})
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&confirm, "confirm", true, "wait for confirmation that doors are unlocked")
	cmd.Flags().IntVar(&confirmWait, "confirm-wait", 90, "max seconds to wait for confirmation")

	return cmd
}

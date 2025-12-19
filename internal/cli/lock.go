package cli

import (
	"context"
	"fmt"
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
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
				// Send lock command
				if err := client.DoorLock(ctx, internalVIN); err != nil {
					return fmt.Errorf("failed to lock doors: %w", err)
				}

				// If confirmation disabled, return immediately
				if !confirm {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Doors locked successfully")
					return nil
				}

				// Wait for confirmation
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Lock command sent, waiting for confirmation...")
				result := waitForDoorsLocked(
					ctx,
					cmd.OutOrStdout(),
					client,
					internalVIN,
					time.Duration(confirmWait)*time.Second,
					5*time.Second, // poll every 5 seconds
				)

				if result.err != nil {
					return fmt.Errorf("failed to confirm lock status: %w", result.err)
				}

				if result.success {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Doors locked successfully")
				} else {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Lock command sent (confirmation timeout)")
				}

				return nil
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
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
				// Send unlock command
				if err := client.DoorUnlock(ctx, internalVIN); err != nil {
					return fmt.Errorf("failed to unlock doors: %w", err)
				}

				// If confirmation disabled, return immediately
				if !confirm {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Doors unlocked successfully")
					return nil
				}

				// Wait for confirmation
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Unlock command sent, waiting for confirmation...")
				result := waitForDoorsUnlocked(
					ctx,
					cmd.OutOrStdout(),
					client,
					internalVIN,
					time.Duration(confirmWait)*time.Second,
					5*time.Second, // poll every 5 seconds
				)

				if result.err != nil {
					return fmt.Errorf("failed to confirm unlock status: %w", result.err)
				}

				if result.success {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Doors unlocked successfully")
				} else {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Unlock command sent (confirmation timeout)")
				}

				return nil
			})
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&confirm, "confirm", true, "wait for confirmation that doors are unlocked")
	cmd.Flags().IntVar(&confirmWait, "confirm-wait", 90, "max seconds to wait for confirmation")

	return cmd
}

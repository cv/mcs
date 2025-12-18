package cli

import (
	"context"
	"fmt"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewLockCmd creates the lock command
func NewLockCmd() *cobra.Command {
	lockCmd := &cobra.Command{
		Use:   "lock",
		Short: "Lock vehicle doors",
		Long:  `Lock all vehicle doors remotely.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLock(cmd)
		},
		SilenceUsage: true,
	}

	return lockCmd
}

// NewUnlockCmd creates the unlock command
func NewUnlockCmd() *cobra.Command {
	unlockCmd := &cobra.Command{
		Use:   "unlock",
		Short: "Unlock vehicle doors",
		Long:  `Unlock all vehicle doors remotely.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnlock(cmd)
		},
		SilenceUsage: true,
	}

	return unlockCmd
}

// runLock executes the lock command
func runLock(cmd *cobra.Command) error {
	return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
		if err := client.DoorLock(ctx, internalVIN); err != nil {
			return fmt.Errorf("failed to lock doors: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Doors locked successfully")
		return nil
	})
}

// runUnlock executes the unlock command
func runUnlock(cmd *cobra.Command) error {
	return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
		if err := client.DoorUnlock(ctx, internalVIN); err != nil {
			return fmt.Errorf("failed to unlock doors: %w", err)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Doors unlocked successfully")
		return nil
	})
}

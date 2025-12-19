package cli

import (
	"context"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewLockCmd creates the lock command
func NewLockCmd() *cobra.Command {
	return NewSimpleCommand(SimpleCommandConfig{
		Use:   "lock",
		Short: "Lock vehicle doors",
		Long:  `Lock all vehicle doors remotely.`,
		APICall: func(ctx context.Context, client *api.Client, vin string) error {
			return client.DoorLock(ctx, vin)
		},
		SuccessMsg:   "Doors locked successfully",
		ErrorMsgTmpl: "failed to lock doors: %w",
	})
}

// NewUnlockCmd creates the unlock command
func NewUnlockCmd() *cobra.Command {
	return NewSimpleCommand(SimpleCommandConfig{
		Use:   "unlock",
		Short: "Unlock vehicle doors",
		Long:  `Unlock all vehicle doors remotely.`,
		APICall: func(ctx context.Context, client *api.Client, vin string) error {
			return client.DoorUnlock(ctx, vin)
		},
		SuccessMsg:   "Doors unlocked successfully",
		ErrorMsgTmpl: "failed to unlock doors: %w",
	})
}

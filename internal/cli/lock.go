package cli

import (
	"fmt"

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
	// Create API client (with cached credentials if available)
	client, err := createAPIClient()
	if err != nil {
		return err
	}
	defer saveClientCache(client)

	// Get vehicle base info to retrieve internal VIN
	vecBaseInfos, err := client.GetVecBaseInfos()
	if err != nil {
		return fmt.Errorf("failed to get vehicle info: %w", err)
	}

	internalVIN, err := getInternalVIN(vecBaseInfos)
	if err != nil {
		return err
	}

	// Lock doors
	if err := client.DoorLock(internalVIN); err != nil {
		return fmt.Errorf("failed to lock doors: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Doors locked successfully")
	return nil
}

// runUnlock executes the unlock command
func runUnlock(cmd *cobra.Command) error {
	// Create API client (with cached credentials if available)
	client, err := createAPIClient()
	if err != nil {
		return err
	}
	defer saveClientCache(client)

	// Get vehicle base info to retrieve internal VIN
	vecBaseInfos, err := client.GetVecBaseInfos()
	if err != nil {
		return fmt.Errorf("failed to get vehicle info: %w", err)
	}

	internalVIN, err := getInternalVIN(vecBaseInfos)
	if err != nil {
		return err
	}

	// Unlock doors
	if err := client.DoorUnlock(internalVIN); err != nil {
		return fmt.Errorf("failed to unlock doors: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Doors unlocked successfully")
	return nil
}

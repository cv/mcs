package cli

import (
	"fmt"

	"github.com/cv/cx90/internal/api"
	"github.com/cv/cx90/internal/config"
	"github.com/spf13/cobra"
)

// NewChargeCmd creates the charge command
func NewChargeCmd() *cobra.Command {
	chargeCmd := &cobra.Command{
		Use:   "charge",
		Short: "Control vehicle charging",
		Long:  `Control vehicle charging (start/stop).`,
	}

	// Add subcommands
	chargeCmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Start charging",
		Long:  `Start charging the vehicle battery.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChargeStart(cmd)
		},
		SilenceUsage: true,
	})

	chargeCmd.AddCommand(&cobra.Command{
		Use:   "stop",
		Short: "Stop charging",
		Long:  `Stop charging the vehicle battery.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChargeStop(cmd)
		},
		SilenceUsage: true,
	})

	return chargeCmd
}

// runChargeStart executes the charge start command
func runChargeStart(cmd *cobra.Command) error {
	// Load configuration
	cfg, err := config.Load(ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Create API client
	client, err := api.NewClient(cfg.Email, cfg.Password, cfg.Region)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Get vehicle base info to retrieve internal VIN
	vecBaseInfos, err := client.GetVecBaseInfos()
	if err != nil {
		return fmt.Errorf("failed to get vehicle info: %w", err)
	}

	internalVIN, err := getInternalVIN(vecBaseInfos)
	if err != nil {
		return err
	}

	// Start charging
	if err := client.ChargeStart(internalVIN); err != nil {
		return fmt.Errorf("failed to start charging: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Charging started successfully")
	return nil
}

// runChargeStop executes the charge stop command
func runChargeStop(cmd *cobra.Command) error {
	// Load configuration
	cfg, err := config.Load(ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Create API client
	client, err := api.NewClient(cfg.Email, cfg.Password, cfg.Region)
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Get vehicle base info to retrieve internal VIN
	vecBaseInfos, err := client.GetVecBaseInfos()
	if err != nil {
		return fmt.Errorf("failed to get vehicle info: %w", err)
	}

	internalVIN, err := getInternalVIN(vecBaseInfos)
	if err != nil {
		return err
	}

	// Stop charging
	if err := client.ChargeStop(internalVIN); err != nil {
		return fmt.Errorf("failed to stop charging: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Charging stopped successfully")
	return nil
}

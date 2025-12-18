package cli

import (
	"fmt"

	"github.com/cv/cx90/internal/api"
	"github.com/cv/cx90/internal/config"
	"github.com/spf13/cobra"
)

// NewStartCmd creates the start command
func NewStartCmd() *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start vehicle engine",
		Long:  `Start the vehicle engine remotely.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(cmd)
		},
		SilenceUsage: true,
	}

	return startCmd
}

// NewStopCmd creates the stop command
func NewStopCmd() *cobra.Command {
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop vehicle engine",
		Long:  `Stop the vehicle engine remotely.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop(cmd)
		},
		SilenceUsage: true,
	}

	return stopCmd
}

// runStart executes the start command
func runStart(cmd *cobra.Command) error {
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

	// Start engine
	if err := client.EngineStart(internalVIN); err != nil {
		return fmt.Errorf("failed to start engine: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Engine started successfully")
	return nil
}

// runStop executes the stop command
func runStop(cmd *cobra.Command) error {
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

	// Stop engine
	if err := client.EngineStop(internalVIN); err != nil {
		return fmt.Errorf("failed to stop engine: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Engine stopped successfully")
	return nil
}

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewClimateCmd creates the climate command
func NewClimateCmd() *cobra.Command {
	climateCmd := &cobra.Command{
		Use:   "climate",
		Short: "Control vehicle climate (HVAC)",
		Long:  `Control vehicle climate system (on/off).`,
	}

	// Add subcommands
	climateCmd.AddCommand(&cobra.Command{
		Use:   "on",
		Short: "Turn climate on",
		Long:  `Turn the vehicle HVAC system on.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClimateOn(cmd)
		},
		SilenceUsage: true,
	})

	climateCmd.AddCommand(&cobra.Command{
		Use:   "off",
		Short: "Turn climate off",
		Long:  `Turn the vehicle HVAC system off.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClimateOff(cmd)
		},
		SilenceUsage: true,
	})

	return climateCmd
}

// runClimateOn executes the climate on command
func runClimateOn(cmd *cobra.Command) error {
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

	// Turn HVAC on
	if err := client.HVACOn(internalVIN); err != nil {
		return fmt.Errorf("failed to turn HVAC on: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Climate turned on successfully")
	return nil
}

// runClimateOff executes the climate off command
func runClimateOff(cmd *cobra.Command) error {
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

	// Turn HVAC off
	if err := client.HVACOff(internalVIN); err != nil {
		return fmt.Errorf("failed to turn HVAC off: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Climate turned off successfully")
	return nil
}

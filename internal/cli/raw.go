package cli

import (
	"encoding/json"
	"fmt"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewRawCmd creates the raw command for debugging
func NewRawCmd() *cobra.Command {
	rawCmd := &cobra.Command{
		Use:   "raw",
		Short: "Output raw API responses (for debugging)",
		Long:  `Output raw JSON responses from the API for debugging purposes.`,
	}

	// Add subcommands
	rawCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Get raw vehicle status",
		Long:  `Get the raw vehicle status JSON from the API.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawStatus(cmd)
		},
		SilenceUsage: true,
	})

	rawCmd.AddCommand(&cobra.Command{
		Use:   "ev",
		Short: "Get raw EV status",
		Long:  `Get the raw EV vehicle status JSON from the API.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawEV(cmd)
		},
		SilenceUsage: true,
	})

	return rawCmd
}

// runRawStatus executes the raw status command
func runRawStatus(cmd *cobra.Command) error {
	return withVehicleClient(func(client *api.Client, internalVIN string) error {
		vehicleStatus, err := client.GetVehicleStatus(internalVIN)
		if err != nil {
			return fmt.Errorf("failed to get vehicle status: %w", err)
		}

		jsonBytes, err := json.MarshalIndent(vehicleStatus, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), string(jsonBytes))
		return nil
	})
}

// runRawEV executes the raw ev command
func runRawEV(cmd *cobra.Command) error {
	return withVehicleClient(func(client *api.Client, internalVIN string) error {
		evStatus, err := client.GetEVVehicleStatus(internalVIN)
		if err != nil {
			return fmt.Errorf("failed to get EV status: %w", err)
		}

		jsonBytes, err := json.MarshalIndent(evStatus, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

		fmt.Fprintln(cmd.OutOrStdout(), string(jsonBytes))
		return nil
	})
}

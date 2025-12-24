package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewRawCmd creates the raw command for debugging.
func NewRawCmd() *cobra.Command {
	rawCmd := &cobra.Command{
		Use:   "raw",
		Short: "Output raw API responses (for debugging)",
		Long:  `Output raw JSON responses from the API for debugging purposes.`,
		Example: `  # Get raw vehicle status JSON
  mcs raw status

  # Get raw EV status JSON
  mcs raw ev

  # Get raw vehicle info JSON
  mcs raw vehicle

  # Example output (truncated):
  # {
  #   "remoteInfos": [{
  #     "vin": "JM3XXXXXXXXXX1234",
  #     "vehicleStatus": {
  #       "batteryLevel": 85,
  #       "fuelLevel": 75,
  #       ...
  #     }
  #   }],
  #   "alertInfos": [...],
  #   ...
  # }`,
	}

	// Add subcommands
	rawCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Get raw vehicle status",
		Long:  `Get the raw vehicle status JSON from the API.`,
		Example: `  # Get raw vehicle status
  mcs raw status

  # Output includes remoteInfos, alertInfos, and vehicle status data`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawStatus(cmd)
		},
		SilenceUsage: true,
	})

	rawCmd.AddCommand(&cobra.Command{
		Use:   "ev",
		Short: "Get raw EV status",
		Long:  `Get the raw EV vehicle status JSON from the API.`,
		Example: `  # Get raw EV status
  mcs raw ev

  # Output includes battery, charging, and EV-specific data`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawEV(cmd)
		},
		SilenceUsage: true,
	})

	rawCmd.AddCommand(&cobra.Command{
		Use:   "vehicle",
		Short: "Get raw vehicle info",
		Long:  `Get the raw vehicle base info JSON from the API.`,
		Example: `  # Get raw vehicle base info
  mcs raw vehicle

  # Output includes VIN, model, year, and vehicle metadata`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRawVehicle(cmd)
		},
		SilenceUsage: true,
	})

	return rawCmd
}

// runRawStatus executes the raw status command.
func runRawStatus(cmd *cobra.Command) error {
	return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
		vehicleStatus, err := client.GetVehicleStatus(ctx, string(internalVIN))
		if err != nil {
			return fmt.Errorf("failed to get vehicle status: %w", err)
		}

		jsonBytes, err := json.MarshalIndent(vehicleStatus, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(jsonBytes))

		return nil
	})
}

// runRawEV executes the raw ev command.
func runRawEV(cmd *cobra.Command) error {
	return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
		evStatus, err := client.GetEVVehicleStatus(ctx, string(internalVIN))
		if err != nil {
			return fmt.Errorf("failed to get EV status: %w", err)
		}

		jsonBytes, err := json.MarshalIndent(evStatus, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

		_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(jsonBytes))

		return nil
	})
}

// runRawVehicle executes the raw vehicle command.
func runRawVehicle(cmd *cobra.Command) error {
	client, err := createAPIClient(cmd.Context())
	if err != nil {
		return err
	}
	defer saveClientCache(client)

	vecBaseInfos, err := client.GetVecBaseInfos(cmd.Context())
	if err != nil {
		return fmt.Errorf("failed to get vehicle info: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(vecBaseInfos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(jsonBytes))

	return nil
}

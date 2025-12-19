package cli

import (
	"context"
	"fmt"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewClimateCmd creates the climate command
func NewClimateCmd() *cobra.Command {
	climateCmd := &cobra.Command{
		Use:   "climate",
		Short: "Control vehicle climate (HVAC)",
		Long:  `Control vehicle climate system (on/off/set).`,
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

	// Add set subcommand with flags
	climateCmd.AddCommand(newClimateSetCmd())

	return climateCmd
}

// newClimateSetCmd creates the climate set subcommand
func newClimateSetCmd() *cobra.Command {
	var temperature float64
	var tempUnit string
	var frontDefroster bool
	var rearDefroster bool

	setCmd := &cobra.Command{
		Use:   "set",
		Short: "Set climate temperature and defroster settings",
		Long: `Set the vehicle HVAC temperature and defroster settings.

Examples:
  mcs climate set --temp 22                    # Set to 22°C
  mcs climate set --temp 72 --unit f           # Set to 72°F
  mcs climate set --temp 20 --front-defrost    # Set 20°C with front defroster on
  mcs climate set --temp 21 --rear-defrost     # Set 21°C with rear defroster on`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClimateSet(cmd, temperature, tempUnit, frontDefroster, rearDefroster)
		},
		SilenceUsage: true,
	}

	setCmd.Flags().Float64Var(&temperature, "temp", 0, "temperature to set (required)")
	setCmd.Flags().StringVar(&tempUnit, "unit", "c", "temperature unit: 'c' for Celsius, 'f' for Fahrenheit")
	setCmd.Flags().BoolVar(&frontDefroster, "front-defrost", false, "enable front defroster")
	setCmd.Flags().BoolVar(&rearDefroster, "rear-defrost", false, "enable rear defroster")

	_ = setCmd.MarkFlagRequired("temp")

	return setCmd
}

// runClimateOn executes the climate on command
func runClimateOn(cmd *cobra.Command) error {
	return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
		if err := client.HVACOn(ctx, internalVIN); err != nil {
			return fmt.Errorf("failed to turn HVAC on: %w", err)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Climate turned on successfully")
		return nil
	})
}

// runClimateOff executes the climate off command
func runClimateOff(cmd *cobra.Command) error {
	return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
		if err := client.HVACOff(ctx, internalVIN); err != nil {
			return fmt.Errorf("failed to turn HVAC off: %w", err)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Climate turned off successfully")
		return nil
	})
}

// runClimateSet executes the climate set command
func runClimateSet(cmd *cobra.Command, temperature float64, tempUnit string, frontDefroster, rearDefroster bool) error {
	unit, err := api.ParseTemperatureUnit(tempUnit)
	if err != nil {
		return err
	}

	return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
		if err := client.SetHVACSetting(ctx, internalVIN, temperature, unit, frontDefroster, rearDefroster); err != nil {
			return fmt.Errorf("failed to set HVAC settings: %w", err)
		}

		msg := fmt.Sprintf("Climate set to %.1f%s", temperature, unit.String())
		if frontDefroster {
			msg += " with front defroster on"
		}
		if rearDefroster {
			if frontDefroster {
				msg += " and rear defroster on"
			} else {
				msg += " with rear defroster on"
			}
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), msg)
		return nil
	})
}

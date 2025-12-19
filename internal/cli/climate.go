package cli

import (
	"context"
	"fmt"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewClimateCmd creates the climate command
func NewClimateCmd() *cobra.Command {
	climateCmd := NewParentWithSubcommands(
		"climate",
		"Control vehicle climate (HVAC)",
		`Control vehicle climate system (on/off/set).`,
		`  # Turn climate on
  mcs climate on

  # Turn climate off
  mcs climate off

  # Set temperature to 22°C
  mcs climate set --temp 22`,
		[]SimpleCommandConfig{
			{
				Use:   "on",
				Short: "Turn climate on",
				Long:  `Turn the vehicle HVAC system on.`,
				Example: `  # Turn the vehicle HVAC system on
  mcs climate on

  # Expected output on success:
  # Climate turned on successfully`,
				APICall: func(ctx context.Context, client *api.Client, vin string) error {
					return client.HVACOn(ctx, vin)
				},
				SuccessMsg:   "Climate turned on successfully",
				ErrorMsgTmpl: "failed to turn HVAC on: %w",
			},
			{
				Use:   "off",
				Short: "Turn climate off",
				Long:  `Turn the vehicle HVAC system off.`,
				Example: `  # Turn the vehicle HVAC system off
  mcs climate off

  # Expected output on success:
  # Climate turned off successfully`,
				APICall: func(ctx context.Context, client *api.Client, vin string) error {
					return client.HVACOff(ctx, vin)
				},
				SuccessMsg:   "Climate turned off successfully",
				ErrorMsgTmpl: "failed to turn HVAC off: %w",
			},
		},
	)

	// Add set subcommand with flags (more complex, keep separate)
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
		Long:  `Set the vehicle HVAC temperature and defroster settings.`,
		Example: `  # Set temperature to 22°C
  mcs climate set --temp 22

  # Set temperature to 72°F
  mcs climate set --temp 72 --unit f

  # Set temperature with front defroster on
  mcs climate set --temp 20 --front-defrost

  # Set temperature with rear defroster on
  mcs climate set --temp 21 --rear-defrost

  # Expected output on success:
  # Climate set to 22.0°C`,
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

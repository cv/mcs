package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewClimateCmd creates the climate command.
func NewClimateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "climate",
		Short: "Control vehicle climate (HVAC)",
		Long:  `Control vehicle climate system (on/off/set).`,
		Example: `  # Turn climate on
  mcs climate on

  # Turn climate off
  mcs climate off

  # Set temperature to 22°C
  mcs climate set --temp 22`,
	}

	cmd.AddCommand(newClimateOnCmd())
	cmd.AddCommand(newClimateOffCmd())
	cmd.AddCommand(newClimateSetCmd())

	return cmd
}

// newClimateOnCmd creates the climate on subcommand.
func newClimateOnCmd() *cobra.Command {
	return buildConfirmableCommand(CommandSpec{
		Use:   "on",
		Short: "Turn climate on",
		Long:  `Turn the vehicle HVAC system on.`,
		Example: `  # Turn the vehicle HVAC system on
  mcs climate on

  # Expected output on success:
  # Climate turned on successfully

  # Turn climate on without waiting for confirmation
  mcs climate on --confirm=false

  # Turn climate on and wait up to 60 seconds for confirmation
  mcs climate on --confirm-wait 60`,
		ConfirmFlagUsage: "wait for confirmation that climate has turned on",
		Config: ConfirmableCommandConfig{
			ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
				return client.HVACOn(ctx, string(internalVIN))
			},
			WaitFunc: func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
				return waitForHvacOn(ctx, out, &clientAdapter{Client: client}, internalVIN, timeout, pollInterval)
			},
			InitialDelay:  ConfirmationInitialDelay,
			SuccessMsg:    "Climate turned on successfully",
			WaitingMsg:    "Climate on command sent, waiting for confirmation...",
			ActionName:    "turn HVAC on",
			ConfirmName:   "HVAC status",
			TimeoutSuffix: "confirmation timeout",
		},
	})
}

// newClimateOffCmd creates the climate off subcommand.
func newClimateOffCmd() *cobra.Command {
	return buildConfirmableCommand(CommandSpec{
		Use:   "off",
		Short: "Turn climate off",
		Long:  `Turn the vehicle HVAC system off.`,
		Example: `  # Turn the vehicle HVAC system off
  mcs climate off

  # Expected output on success:
  # Climate turned off successfully

  # Turn climate off without waiting for confirmation
  mcs climate off --confirm=false

  # Turn climate off and wait up to 60 seconds for confirmation
  mcs climate off --confirm-wait 60`,
		ConfirmFlagUsage: "wait for confirmation that climate has turned off",
		Config: ConfirmableCommandConfig{
			ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
				return client.HVACOff(ctx, string(internalVIN))
			},
			WaitFunc: func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
				return waitForHvacOff(ctx, out, &clientAdapter{Client: client}, internalVIN, timeout, pollInterval)
			},
			InitialDelay:  ConfirmationInitialDelay,
			SuccessMsg:    "Climate turned off successfully",
			WaitingMsg:    "Climate off command sent, waiting for confirmation...",
			ActionName:    "turn HVAC off",
			ConfirmName:   "HVAC status",
			TimeoutSuffix: "confirmation timeout",
		},
	})
}

// newClimateSetCmd creates the climate set subcommand.
func newClimateSetCmd() *cobra.Command {
	var temperature float64
	var tempUnit string
	var frontDefroster bool
	var rearDefroster bool
	var confirm bool
	var confirmWait int

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

  # Set temperature without waiting for confirmation
  mcs climate set --temp 22 --confirm=false

  # Set temperature and wait up to 60 seconds for confirmation
  mcs climate set --temp 22 --confirm-wait 60

  # Expected output on success:
  # Climate set to 22.0°C`,
		RunE: func(cmd *cobra.Command, args []string) error {
			unit, err := api.ParseTemperatureUnit(tempUnit)
			if err != nil {
				return err
			}

			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
				// Build success message
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

				// Convert temperature to Celsius for comparison (API returns Celsius)
				targetTempC := temperature
				if unit == api.Fahrenheit {
					targetTempC = (temperature - 32) * 5 / 9
				}

				config := ConfirmableCommandConfig{
					ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
						return client.SetHVACSetting(ctx, string(internalVIN), temperature, unit, frontDefroster, rearDefroster)
					},
					WaitFunc: func(ctx context.Context, out io.Writer, client *api.Client, internalVIN api.InternalVIN, timeout, pollInterval time.Duration) confirmationResult {
						return waitForHvacSettings(ctx, out, &clientAdapter{Client: client}, internalVIN, targetTempC, frontDefroster, rearDefroster, timeout, pollInterval)
					},
					InitialDelay:  ConfirmationInitialDelay,
					SuccessMsg:    msg,
					WaitingMsg:    "Climate set command sent, waiting for confirmation...",
					ActionName:    "set HVAC settings",
					ConfirmName:   "HVAC settings",
					TimeoutSuffix: "confirmation timeout",
				}

				return executeConfirmableCommand(ctx, cmd.OutOrStdout(), client, internalVIN, config, confirm, confirmWait)
			})
		},
		SilenceUsage: true,
	}

	setCmd.Flags().Float64Var(&temperature, "temp", 0, "temperature to set (required)")
	setCmd.Flags().StringVar(&tempUnit, "unit", "c", "temperature unit: 'c' for Celsius, 'f' for Fahrenheit")
	setCmd.Flags().BoolVar(&frontDefroster, "front-defrost", false, "enable front defroster")
	setCmd.Flags().BoolVar(&rearDefroster, "rear-defrost", false, "enable rear defroster")
	setCmd.Flags().BoolVar(&confirm, "confirm", true, "wait for confirmation that settings have been applied")
	setCmd.Flags().IntVar(&confirmWait, "confirm-wait", 90, "max seconds to wait for confirmation")

	_ = setCmd.MarkFlagRequired("temp")

	return setCmd
}

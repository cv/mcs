package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewClimateCmd creates the climate command
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

// newClimateOnCmd creates the climate on subcommand
func newClimateOnCmd() *cobra.Command {
	var confirm bool
	var confirmWait int

	cmd := &cobra.Command{
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
				// Send HVAC on command
				if err := client.HVACOn(ctx, internalVIN); err != nil {
					return fmt.Errorf("failed to turn HVAC on: %w", err)
				}

				// If confirmation disabled, return immediately
				if !confirm {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Climate turned on successfully")
					return nil
				}

				// Wait for confirmation
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Climate on command sent, waiting for confirmation...")
				result := waitForHvacOn(
					ctx,
					cmd.OutOrStdout(),
					client,
					internalVIN,
					time.Duration(confirmWait)*time.Second,
					5*time.Second, // poll every 5 seconds
				)

				if result.err != nil {
					return fmt.Errorf("failed to confirm HVAC status: %w", result.err)
				}

				if result.success {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Climate turned on successfully")
				} else {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Climate on command sent (confirmation timeout)")
				}

				return nil
			})
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&confirm, "confirm", true, "wait for confirmation that climate has turned on")
	cmd.Flags().IntVar(&confirmWait, "confirm-wait", 90, "max seconds to wait for confirmation")

	return cmd
}

// newClimateOffCmd creates the climate off subcommand
func newClimateOffCmd() *cobra.Command {
	var confirm bool
	var confirmWait int

	cmd := &cobra.Command{
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
				// Send HVAC off command
				if err := client.HVACOff(ctx, internalVIN); err != nil {
					return fmt.Errorf("failed to turn HVAC off: %w", err)
				}

				// If confirmation disabled, return immediately
				if !confirm {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Climate turned off successfully")
					return nil
				}

				// Wait for confirmation
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Climate off command sent, waiting for confirmation...")
				result := waitForHvacOff(
					ctx,
					cmd.OutOrStdout(),
					client,
					internalVIN,
					time.Duration(confirmWait)*time.Second,
					5*time.Second, // poll every 5 seconds
				)

				if result.err != nil {
					return fmt.Errorf("failed to confirm HVAC status: %w", result.err)
				}

				if result.success {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Climate turned off successfully")
				} else {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Climate off command sent (confirmation timeout)")
				}

				return nil
			})
		},
		SilenceUsage: true,
	}

	cmd.Flags().BoolVar(&confirm, "confirm", true, "wait for confirmation that climate has turned off")
	cmd.Flags().IntVar(&confirmWait, "confirm-wait", 90, "max seconds to wait for confirmation")

	return cmd
}

// newClimateSetCmd creates the climate set subcommand
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

			return withVehicleClient(cmd.Context(), func(ctx context.Context, client *api.Client, internalVIN string) error {
				// Send HVAC settings command
				if err := client.SetHVACSetting(ctx, internalVIN, temperature, unit, frontDefroster, rearDefroster); err != nil {
					return fmt.Errorf("failed to set HVAC settings: %w", err)
				}

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

				// If confirmation disabled, return immediately
				if !confirm {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), msg)
					return nil
				}

				// Convert temperature to Celsius for comparison (API returns Celsius)
				targetTempC := temperature
				if unit == api.Fahrenheit {
					targetTempC = (temperature - 32) * 5 / 9
				}

				// Wait for confirmation
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Climate set command sent, waiting for confirmation...")
				result := waitForHvacSettings(
					ctx,
					cmd.OutOrStdout(),
					client,
					internalVIN,
					targetTempC,
					frontDefroster,
					rearDefroster,
					time.Duration(confirmWait)*time.Second,
					5*time.Second, // poll every 5 seconds
				)

				if result.err != nil {
					return fmt.Errorf("failed to confirm HVAC settings: %w", result.err)
				}

				if result.success {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), msg)
				} else {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Climate set command sent (confirmation timeout)")
				}

				return nil
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

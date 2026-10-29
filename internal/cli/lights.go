package cli

import (
	"context"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewLightsCmd creates the lights command (the hazard lights, to find the car).
func NewLightsCmd() *cobra.Command {
	lightsCmd := &cobra.Command{
		Use:   "lights",
		Short: "Flash the hazard lights",
		Long:  `Turn the vehicle's hazard lights on or off remotely (to find it in a car park).`,
	}

	for _, on := range []bool{true, false} {
		use, desc, msg := "off", "turn hazard lights off", "Hazard lights off command sent"
		if on {
			use, desc, msg = "on", "turn hazard lights on", "Hazard lights on command sent"
		}
		lightsCmd.AddCommand(buildConfirmableCommand(CommandSpec{
			Use:              use,
			Short:            "Turn the hazard lights " + use,
			Long:             "Turn the vehicle's hazard lights " + use + " remotely.",
			Example:          "  mcs lights " + use,
			ConfirmFlagUsage: "unused: the API reports no light status",
			Config: ConfirmableCommandConfig{
				ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
					if on {
						return client.LightsOn(ctx, string(internalVIN))
					}

					return client.LightsOff(ctx, string(internalVIN))
				},
				WaitFunc:      nil,
				SuccessMsg:    msg,
				WaitingMsg:    msg,
				ActionName:    desc,
				ConfirmName:   "light status",
				TimeoutSuffix: confirmationTimeoutSuffix,
			},
		}))
	}

	return lightsCmd
}

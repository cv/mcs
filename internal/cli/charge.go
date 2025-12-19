package cli

import (
	"context"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewChargeCmd creates the charge command
func NewChargeCmd() *cobra.Command {
	return NewParentWithSubcommands(
		"charge",
		"Control vehicle charging",
		`Control vehicle charging (start/stop).`,
		[]SimpleCommandConfig{
			{
				Use:   "start",
				Short: "Start charging",
				Long:  `Start charging the vehicle battery.`,
				APICall: func(ctx context.Context, client *api.Client, vin string) error {
					return client.ChargeStart(ctx, vin)
				},
				SuccessMsg:   "Charging started successfully",
				ErrorMsgTmpl: "failed to start charging: %w",
			},
			{
				Use:   "stop",
				Short: "Stop charging",
				Long:  `Stop charging the vehicle battery.`,
				APICall: func(ctx context.Context, client *api.Client, vin string) error {
					return client.ChargeStop(ctx, vin)
				},
				SuccessMsg:   "Charging stopped successfully",
				ErrorMsgTmpl: "failed to stop charging: %w",
			},
		},
	)
}

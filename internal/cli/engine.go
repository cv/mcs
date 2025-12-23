package cli

import (
	"context"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewStartCmd creates the start command
func NewStartCmd() *cobra.Command {
	return buildConfirmableCommand(CommandSpec{
		Use:   "start",
		Short: "Start vehicle engine",
		Long:  `Start the vehicle engine remotely.`,
		Example: `  # Start the vehicle engine remotely
  mcs start

  # Expected output on success:
  # Engine started successfully

  # Start engine without waiting for confirmation
  mcs start --confirm=false

  # Start engine and wait up to 60 seconds for confirmation
  mcs start --confirm-wait 60`,
		ConfirmFlagUsage: "wait for confirmation that engine is running",
		Config: ConfirmableCommandConfig{
			ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
				return client.EngineStart(ctx, string(internalVIN))
			},
			// WaitFunc: nil - No reliable API field for engine status
			// Previously used HVAC status as proxy, which was incorrect
			WaitFunc:      nil,
			SuccessMsg:    "Engine start command sent",
			WaitingMsg:    "Start command sent, waiting for confirmation...",
			ActionName:    "start engine",
			ConfirmName:   "engine status",
			TimeoutSuffix: "confirmation timeout",
		},
	})
}

// NewStopCmd creates the stop command
func NewStopCmd() *cobra.Command {
	return buildConfirmableCommand(CommandSpec{
		Use:   "stop",
		Short: "Stop vehicle engine",
		Long:  `Stop the vehicle engine remotely.`,
		Example: `  # Stop the vehicle engine remotely
  mcs stop

  # Expected output on success:
  # Engine stopped successfully

  # Stop engine without waiting for confirmation
  mcs stop --confirm=false

  # Stop engine and wait up to 60 seconds for confirmation
  mcs stop --confirm-wait 60`,
		ConfirmFlagUsage: "wait for confirmation that engine is stopped",
		Config: ConfirmableCommandConfig{
			ActionFunc: func(ctx context.Context, client *api.Client, internalVIN api.InternalVIN) error {
				return client.EngineStop(ctx, string(internalVIN))
			},
			// WaitFunc: nil - No reliable API field for engine status
			// Previously used HVAC status as proxy, which was incorrect
			WaitFunc:      nil,
			SuccessMsg:    "Engine stop command sent",
			WaitingMsg:    "Stop command sent, waiting for confirmation...",
			ActionName:    "stop engine",
			ConfirmName:   "engine status",
			TimeoutSuffix: "confirmation timeout",
		},
	})
}

package cli

import (
	"context"

	"github.com/cv/mcs/internal/api"
	"github.com/spf13/cobra"
)

// NewStartCmd creates the start command
func NewStartCmd() *cobra.Command {
	return NewSimpleCommand(SimpleCommandConfig{
		Use:   "start",
		Short: "Start vehicle engine",
		Long:  `Start the vehicle engine remotely.`,
		APICall: func(ctx context.Context, client *api.Client, vin string) error {
			return client.EngineStart(ctx, vin)
		},
		SuccessMsg:   "Engine started successfully",
		ErrorMsgTmpl: "failed to start engine: %w",
	})
}

// NewStopCmd creates the stop command
func NewStopCmd() *cobra.Command {
	return NewSimpleCommand(SimpleCommandConfig{
		Use:   "stop",
		Short: "Stop vehicle engine",
		Long:  `Stop the vehicle engine remotely.`,
		APICall: func(ctx context.Context, client *api.Client, vin string) error {
			return client.EngineStop(ctx, vin)
		},
		SuccessMsg:   "Engine stopped successfully",
		ErrorMsgTmpl: "failed to stop engine: %w",
	})
}

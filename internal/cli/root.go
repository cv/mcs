package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time
	Version = "dev"

	// ConfigFile is the path to the config file
	ConfigFile string
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mcs",
		Short: "Control your connected vehicle",
		Long: `mcs is a CLI tool for controlling your connected vehicle via manufacturer API.

Features:
  - Check vehicle status (battery, fuel, location, etc.)
  - Lock/unlock doors
  - Start/stop engine
  - Control charging
  - Control climate (HVAC)

Configuration:
  Configuration can be loaded from ~/.config/mcs/config.toml or via environment variables.

  Environment variables:
    MCS_EMAIL     - Your account email
    MCS_PASSWORD  - Your account password
    MCS_REGION    - Region (MNAO, MME, or MJO)

Example config.toml:
  email = "your.email@example.com"
  password = "your-password"
  region = "MNAO"
`,
		Example: `  # Check vehicle status
  $ mcs status
  Battery: 85% (320 km range)
  Fuel: 45% (380 km range)
  Doors: Locked
  Location: 37.7749, -122.4194

  # Get status as JSON
  $ mcs status --json
  {"battery": {"level": 85, "range_km": 320}, "fuel": {"level": 45, "range_km": 380}, ...}

  # Lock the vehicle
  $ mcs lock
  Vehicle locked successfully

  # Start the engine
  $ mcs start
  Engine started successfully

  # Control charging
  $ mcs charge start
  Charging started successfully

  # Control climate
  $ mcs climate on --temp 22
  Climate control turned on successfully`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add version flag
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("mcs version {{.Version}}\n")

	// Add global flags
	rootCmd.PersistentFlags().StringVarP(&ConfigFile, "config", "c", "", "config file (default is ~/.config/mcs/config.toml)")

	return rootCmd
}

// Execute runs the root command with signal-aware context
func Execute() error {
	// Create context that cancels on SIGINT or SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	rootCmd := NewRootCmd()

	// Add subcommands
	rootCmd.AddCommand(NewStatusCmd())
	rootCmd.AddCommand(NewLockCmd())
	rootCmd.AddCommand(NewUnlockCmd())
	rootCmd.AddCommand(NewStartCmd())
	rootCmd.AddCommand(NewStopCmd())
	rootCmd.AddCommand(NewChargeCmd())
	rootCmd.AddCommand(NewClimateCmd())
	rootCmd.AddCommand(NewRawCmd())

	return rootCmd.ExecuteContext(ctx)
}

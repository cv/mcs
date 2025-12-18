package cli

import (
	"fmt"

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
		Use:   "cx90",
		Short: "Control your Mazda CX-90 PHEV",
		Long: `cx90 is a CLI tool for controlling your Mazda CX-90 PHEV via the MyMazda API.

Features:
  - Check vehicle status (battery, fuel, location, etc.)
  - Lock/unlock doors
  - Start/stop engine
  - Control charging
  - Control climate (HVAC)

Configuration:
  Configuration can be loaded from ~/.config/cx90/config.toml or via environment variables.

  Environment variables:
    MYMAZDA_EMAIL     - Your MyMazda account email
    MYMAZDA_PASSWORD  - Your MyMazda account password
    MYMAZDA_REGION    - Region (MNAO, MME, or MJO)

Example config.toml:
  email = "your.email@example.com"
  password = "your-password"
  region = "MNAO"
`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Add version flag
	rootCmd.Version = Version
	rootCmd.SetVersionTemplate("cx90 version {{.Version}}\n")

	// Add global flags
	rootCmd.PersistentFlags().StringVarP(&ConfigFile, "config", "c", "", "config file (default is ~/.config/cx90/config.toml)")

	return rootCmd
}

// Execute runs the root command
func Execute() error {
	rootCmd := NewRootCmd()
	return rootCmd.Execute()
}

// PrintConfigPath prints the configuration file path
func PrintConfigPath(cmd *cobra.Command, args []string) error {
	configPath := ConfigFile
	if configPath == "" {
		configPath = "~/.config/cx90/config.toml (default)"
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Config path: %s\n", configPath)
	return nil
}

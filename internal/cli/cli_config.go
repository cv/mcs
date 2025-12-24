package cli

import "context"

// CLIConfig holds CLI configuration that was previously stored in package-level globals.
// Using a struct allows tests to run in parallel without race conditions.
type CLIConfig struct {
	// Version is the CLI version, set at build time via ldflags.
	Version string

	// ConfigFile is the path to the config file, set via --config flag.
	ConfigFile string

	// NoColor disables colored output, set via --no-color flag.
	NoColor bool

	// CacheFile is the path to the token cache file.
	// If empty, uses the default location (~/.cache/mcs/token.json).
	// This is primarily used for testing to avoid setting HOME.
	CacheFile string
}

// cliConfigKey is the context key for CLIConfig.
type cliConfigKey struct{}

// ConfigFromContext retrieves the CLIConfig from the context.
// Returns nil if no config is stored in the context.
func ConfigFromContext(ctx context.Context) *CLIConfig {
	if cfg, ok := ctx.Value(cliConfigKey{}).(*CLIConfig); ok {
		return cfg
	}

	return nil
}

// ContextWithConfig returns a new context with the CLIConfig attached.
func ContextWithConfig(ctx context.Context, cfg *CLIConfig) context.Context {
	return context.WithValue(ctx, cliConfigKey{}, cfg)
}

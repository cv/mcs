package cli

import (
	"fmt"
	"log"

	"github.com/cv/mcs/internal/api"
	"github.com/cv/mcs/internal/cache"
	"github.com/cv/mcs/internal/config"
)

// createAPIClient creates an API client with cached credentials if available
func createAPIClient() (*api.Client, error) {
	// Load configuration
	cfg, err := config.Load(ConfigFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create API client
	client, err := api.NewClient(cfg.Email, cfg.Password, cfg.Region)
	if err != nil {
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	// Try to load cached credentials
	cachedCreds, err := cache.Load()
	if err != nil {
		// If cache load fails, just continue without it
		// The client will authenticate normally
		return client, nil
	}

	// If we have valid cached credentials, use them
	if cachedCreds != nil && cachedCreds.IsValid() {
		client.SetCachedCredentials(
			cachedCreds.AccessToken,
			cachedCreds.AccessTokenExpirationTs,
			cachedCreds.EncKey,
			cachedCreds.SignKey,
		)
	}

	return client, nil
}

// saveClientCache saves the client's current credentials to cache
func saveClientCache(client *api.Client) {
	accessToken, expirationTs, encKey, signKey := client.GetCredentials()

	// Only save if we have valid credentials
	if accessToken == "" || expirationTs == 0 || encKey == "" || signKey == "" {
		return
	}

	tokenCache := &cache.TokenCache{
		AccessToken:             accessToken,
		AccessTokenExpirationTs: expirationTs,
		EncKey:                  encKey,
		SignKey:                 signKey,
	}

	if err := cache.Save(tokenCache); err != nil {
		// Log the error but don't fail the command
		// Cache save failures shouldn't break the CLI
		log.Printf("Warning: failed to save token cache: %v", err)
	}
}

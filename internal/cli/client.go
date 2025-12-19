package cli

import (
	"context"
	"fmt"
	"log"

	"github.com/cv/mcs/internal/api"
	"github.com/cv/mcs/internal/cache"
	"github.com/cv/mcs/internal/config"
)

// createAPIClient creates an API client with cached credentials if available
func createAPIClient(ctx context.Context) (*api.Client, error) {
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

// VehicleInfo contains identification information about a vehicle
type VehicleInfo struct {
	InternalVIN string
	VIN         string
	Nickname    string
	ModelName   string
	ModelYear   string
}

// setupVehicleClient is a shared helper that creates the API client and retrieves vehicle info.
// It returns the authenticated client and full vehicle info, deferring cache save to the caller.
func setupVehicleClient(ctx context.Context) (*api.Client, VehicleInfo, error) {
	client, err := createAPIClient(ctx)
	if err != nil {
		return nil, VehicleInfo{}, err
	}

	vecBaseInfos, err := client.GetVecBaseInfos(ctx)
	if err != nil {
		return nil, VehicleInfo{}, fmt.Errorf("failed to get vehicle info: %w", err)
	}

	internalVIN, err := vecBaseInfos.GetInternalVIN()
	if err != nil {
		return nil, VehicleInfo{}, err
	}

	vin, nickname, modelName, modelYear, _ := vecBaseInfos.GetVehicleInfo()

	vehicleInfo := VehicleInfo{
		InternalVIN: internalVIN,
		VIN:         vin,
		Nickname:    nickname,
		ModelName:   modelName,
		ModelYear:   modelYear,
	}

	return client, vehicleInfo, nil
}

// withVehicleClient handles the common CLI setup: create client, get VIN, execute command, save cache.
// The callback receives the context, authenticated client, and the vehicle's internal VIN.
func withVehicleClient(ctx context.Context, fn func(context.Context, *api.Client, string) error) error {
	client, vehicleInfo, err := setupVehicleClient(ctx)
	if err != nil {
		return err
	}
	defer saveClientCache(client)

	return fn(ctx, client, vehicleInfo.InternalVIN)
}

// withVehicleClientEx handles CLI setup and provides extended vehicle info.
// The callback receives the context, authenticated client, and full vehicle info.
func withVehicleClientEx(ctx context.Context, fn func(context.Context, *api.Client, VehicleInfo) error) error {
	client, vehicleInfo, err := setupVehicleClient(ctx)
	if err != nil {
		return err
	}
	defer saveClientCache(client)

	return fn(ctx, client, vehicleInfo)
}

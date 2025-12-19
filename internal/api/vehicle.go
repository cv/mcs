package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// GetVecBaseInfos retrieves the base information for all vehicles associated with the account
func (c *Client) GetVecBaseInfos(ctx context.Context) (*VecBaseInfosResponse, error) {
	bodyParams := map[string]interface{}{
		"internaluserid": InternalUserID,
	}

	responseBytes, err := c.APIRequestJSON(ctx, "POST", "remoteServices/getVecBaseInfos/v4", nil, bodyParams, true, true)
	if err != nil {
		return nil, err
	}

	var typed VecBaseInfosResponse
	if err := json.Unmarshal(responseBytes, &typed); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &typed, nil
}

// buildVehicleStatusParams creates the standard body parameters for vehicle status requests
func buildVehicleStatusParams(internalVIN string) map[string]interface{} {
	return map[string]interface{}{
		"internaluserid": InternalUserID,
		"internalvin":    internalVIN,
		"limit":          1,
		"offset":         0,
		"vecinfotype":    "0",
	}
}

// GetVehicleStatus retrieves the current status of a vehicle
func (c *Client) GetVehicleStatus(ctx context.Context, internalVIN string) (*VehicleStatusResponse, error) {
	bodyParams := buildVehicleStatusParams(internalVIN)

	responseBytes, err := c.APIRequestJSON(ctx, "POST", "remoteServices/getVehicleStatus/v4", nil, bodyParams, true, true)
	if err != nil {
		return nil, err
	}

	var typed VehicleStatusResponse
	if err := json.Unmarshal(responseBytes, &typed); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check result code
	if typed.ResultCode != "200S00" {
		return nil, fmt.Errorf("failed to get vehicle status: result code %s", typed.ResultCode)
	}

	return &typed, nil
}

// GetEVVehicleStatus retrieves the current EV status of a vehicle (battery, charging, HVAC)
func (c *Client) GetEVVehicleStatus(ctx context.Context, internalVIN string) (*EVVehicleStatusResponse, error) {
	bodyParams := buildVehicleStatusParams(internalVIN)

	responseBytes, err := c.APIRequestJSON(ctx, "POST", "remoteServices/getEVVehicleStatus/v4", nil, bodyParams, true, true)
	if err != nil {
		return nil, err
	}

	var typed EVVehicleStatusResponse
	if err := json.Unmarshal(responseBytes, &typed); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check result code
	if typed.ResultCode != "200S00" {
		return nil, fmt.Errorf("failed to get EV vehicle status: result code %s", typed.ResultCode)
	}

	return &typed, nil
}

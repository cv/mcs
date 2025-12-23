package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// Vehicle information endpoint constants
const (
	EndpointGetVecBaseInfos    = "remoteServices/getVecBaseInfos/v4"
	EndpointGetVehicleStatus   = "remoteServices/getVehicleStatus/v4"
	EndpointGetEVVehicleStatus = "remoteServices/getEVVehicleStatus/v4"
)

// GetVecBaseInfos retrieves the base information for all vehicles associated with the account
func (c *Client) GetVecBaseInfos(ctx context.Context) (*VecBaseInfosResponse, error) {
	bodyParams := map[string]any{
		"internaluserid": InternalUserID,
	}

	responseBytes, err := c.APIRequestJSON(ctx, "POST", EndpointGetVecBaseInfos, nil, bodyParams, true, true)
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
func buildVehicleStatusParams(internalVIN string) map[string]any {
	return map[string]any{
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

	responseBytes, err := c.APIRequestJSON(ctx, "POST", EndpointGetVehicleStatus, nil, bodyParams, true, true)
	if err != nil {
		return nil, err
	}

	var typed VehicleStatusResponse
	if err := json.Unmarshal(responseBytes, &typed); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check result code
	if err := checkResultCode(typed.ResultCode, "get vehicle status"); err != nil {
		return nil, err
	}

	return &typed, nil
}

// GetEVVehicleStatus retrieves the current EV status of a vehicle (battery, charging, HVAC)
func (c *Client) GetEVVehicleStatus(ctx context.Context, internalVIN string) (*EVVehicleStatusResponse, error) {
	bodyParams := buildVehicleStatusParams(internalVIN)

	responseBytes, err := c.APIRequestJSON(ctx, "POST", EndpointGetEVVehicleStatus, nil, bodyParams, true, true)
	if err != nil {
		return nil, err
	}

	var typed EVVehicleStatusResponse
	if err := json.Unmarshal(responseBytes, &typed); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check result code
	if err := checkResultCode(typed.ResultCode, "get EV vehicle status"); err != nil {
		return nil, err
	}

	return &typed, nil
}

package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// unmarshalResponse converts a map[string]interface{} response to a typed struct
func unmarshalResponse[T any](response map[string]interface{}) (*T, error) {
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var typed T
	if err := json.Unmarshal(jsonBytes, &typed); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &typed, nil
}

// GetVecBaseInfos retrieves the base information for all vehicles associated with the account
func (c *Client) GetVecBaseInfos(ctx context.Context) (*VecBaseInfosResponse, error) {
	bodyParams := map[string]interface{}{
		"internaluserid": InternalUserID,
	}

	response, err := c.APIRequest(ctx, "POST", "remoteServices/getVecBaseInfos/v4", nil, bodyParams, true, true)
	if err != nil {
		return nil, err
	}

	return unmarshalResponse[VecBaseInfosResponse](response)
}

// GetVehicleStatus retrieves the current status of a vehicle
func (c *Client) GetVehicleStatus(ctx context.Context, internalVIN string) (*VehicleStatusResponse, error) {
	bodyParams := map[string]interface{}{
		"internaluserid": InternalUserID,
		"internalvin":    internalVIN,
		"limit":          1,
		"offset":         0,
		"vecinfotype":    "0",
	}

	response, err := c.APIRequest(ctx, "POST", "remoteServices/getVehicleStatus/v4", nil, bodyParams, true, true)
	if err != nil {
		return nil, err
	}

	typed, err := unmarshalResponse[VehicleStatusResponse](response)
	if err != nil {
		return nil, err
	}

	// Check result code
	if typed.ResultCode != "200S00" {
		return nil, fmt.Errorf("failed to get vehicle status: result code %s", typed.ResultCode)
	}

	return typed, nil
}

// GetEVVehicleStatus retrieves the current EV status of a vehicle (battery, charging, HVAC)
func (c *Client) GetEVVehicleStatus(ctx context.Context, internalVIN string) (*EVVehicleStatusResponse, error) {
	bodyParams := map[string]interface{}{
		"internaluserid": InternalUserID,
		"internalvin":    internalVIN,
		"limit":          1,
		"offset":         0,
		"vecinfotype":    "0",
	}

	response, err := c.APIRequest(ctx, "POST", "remoteServices/getEVVehicleStatus/v4", nil, bodyParams, true, true)
	if err != nil {
		return nil, err
	}

	typed, err := unmarshalResponse[EVVehicleStatusResponse](response)
	if err != nil {
		return nil, err
	}

	// Check result code
	if typed.ResultCode != "200S00" {
		return nil, fmt.Errorf("failed to get EV vehicle status: result code %s", typed.ResultCode)
	}

	return typed, nil
}

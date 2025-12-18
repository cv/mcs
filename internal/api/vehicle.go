package api

import (
	"encoding/json"
	"fmt"
)

// GetVecBaseInfos retrieves the base information for all vehicles associated with the account
func (c *Client) GetVecBaseInfos() (*VecBaseInfosResponse, error) {
	bodyParams := map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
	}

	response, err := c.APIRequest("POST", "remoteServices/getVecBaseInfos/v4", nil, bodyParams, true, true)
	if err != nil {
		return nil, err
	}

	// Convert map to typed response
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var typed VecBaseInfosResponse
	if err := json.Unmarshal(jsonBytes, &typed); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &typed, nil
}

// GetVehicleStatus retrieves the current status of a vehicle
func (c *Client) GetVehicleStatus(internalVIN string) (*VehicleStatusResponse, error) {
	bodyParams := map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    internalVIN,
		"limit":          1,
		"offset":         0,
		"vecinfotype":    "0",
	}

	response, err := c.APIRequest("POST", "remoteServices/getVehicleStatus/v4", nil, bodyParams, true, true)
	if err != nil {
		return nil, err
	}

	// Convert map to typed response
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var typed VehicleStatusResponse
	if err := json.Unmarshal(jsonBytes, &typed); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check result code
	if typed.ResultCode != "200S00" {
		return nil, fmt.Errorf("failed to get vehicle status: result code %s", typed.ResultCode)
	}

	return &typed, nil
}

// GetEVVehicleStatus retrieves the current EV status of a vehicle (battery, charging, HVAC)
func (c *Client) GetEVVehicleStatus(internalVIN string) (*EVVehicleStatusResponse, error) {
	bodyParams := map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    internalVIN,
		"limit":          1,
		"offset":         0,
		"vecinfotype":    "0",
	}

	response, err := c.APIRequest("POST", "remoteServices/getEVVehicleStatus/v4", nil, bodyParams, true, true)
	if err != nil {
		return nil, err
	}

	// Convert map to typed response
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var typed EVVehicleStatusResponse
	if err := json.Unmarshal(jsonBytes, &typed); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check result code
	if typed.ResultCode != "200S00" {
		return nil, fmt.Errorf("failed to get EV vehicle status: result code %s", typed.ResultCode)
	}

	return &typed, nil
}

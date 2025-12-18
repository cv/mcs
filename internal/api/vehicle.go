package api

import "fmt"

// GetVecBaseInfos retrieves the base information for all vehicles associated with the account
func (c *Client) GetVecBaseInfos() (map[string]interface{}, error) {
	bodyParams := map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
	}

	return c.APIRequest("POST", "remoteServices/getVecBaseInfos/v4", nil, bodyParams, true, true)
}

// GetVehicleStatus retrieves the current status of a vehicle
func (c *Client) GetVehicleStatus(internalVIN string) (map[string]interface{}, error) {
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

	// Check result code
	resultCode, ok := response["resultCode"].(string)
	if !ok || resultCode != "200S00" {
		return nil, fmt.Errorf("Failed to get vehicle status: result code %s", resultCode)
	}

	return response, nil
}

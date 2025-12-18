package api

import "fmt"

// DoorLock locks the vehicle doors
func (c *Client) DoorLock(internalVIN string) error {
	bodyParams := map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    internalVIN,
	}

	response, err := c.APIRequest("POST", "remoteServices/doorLock/v4", nil, bodyParams, true, true)
	if err != nil {
		return err
	}

	resultCode, ok := response["resultCode"].(string)
	if !ok || resultCode != "200S00" {
		return fmt.Errorf("Failed to lock doors: result code %s", resultCode)
	}

	return nil
}

// DoorUnlock unlocks the vehicle doors
func (c *Client) DoorUnlock(internalVIN string) error {
	bodyParams := map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    internalVIN,
	}

	response, err := c.APIRequest("POST", "remoteServices/doorUnlock/v4", nil, bodyParams, true, true)
	if err != nil {
		return err
	}

	resultCode, ok := response["resultCode"].(string)
	if !ok || resultCode != "200S00" {
		return fmt.Errorf("Failed to unlock doors: result code %s", resultCode)
	}

	return nil
}

// LightsOn turns the vehicle hazard lights on
func (c *Client) LightsOn(internalVIN string) error {
	bodyParams := map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    internalVIN,
	}

	response, err := c.APIRequest("POST", "remoteServices/lightOn/v4", nil, bodyParams, true, true)
	if err != nil {
		return err
	}

	resultCode, ok := response["resultCode"].(string)
	if !ok || resultCode != "200S00" {
		return fmt.Errorf("Failed to turn lights on: result code %s", resultCode)
	}

	return nil
}

// LightsOff turns the vehicle hazard lights off
func (c *Client) LightsOff(internalVIN string) error {
	bodyParams := map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    internalVIN,
	}

	response, err := c.APIRequest("POST", "remoteServices/lightOff/v4", nil, bodyParams, true, true)
	if err != nil {
		return err
	}

	resultCode, ok := response["resultCode"].(string)
	if !ok || resultCode != "200S00" {
		return fmt.Errorf("Failed to turn lights off: result code %s", resultCode)
	}

	return nil
}

// EngineStart starts the vehicle engine remotely
func (c *Client) EngineStart(internalVIN string) error {
	bodyParams := map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    internalVIN,
	}

	response, err := c.APIRequest("POST", "remoteServices/engineStart/v4", nil, bodyParams, true, true)
	if err != nil {
		return err
	}

	resultCode, ok := response["resultCode"].(string)
	if !ok || resultCode != "200S00" {
		return fmt.Errorf("Failed to start engine: result code %s", resultCode)
	}

	return nil
}

// EngineStop stops the vehicle engine remotely
func (c *Client) EngineStop(internalVIN string) error {
	bodyParams := map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    internalVIN,
	}

	response, err := c.APIRequest("POST", "remoteServices/engineStop/v4", nil, bodyParams, true, true)
	if err != nil {
		return err
	}

	resultCode, ok := response["resultCode"].(string)
	if !ok || resultCode != "200S00" {
		return fmt.Errorf("Failed to stop engine: result code %s", resultCode)
	}

	return nil
}

// ChargeStart starts charging the vehicle (EV/PHEV only)
func (c *Client) ChargeStart(internalVIN string) error {
	bodyParams := map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    internalVIN,
	}

	response, err := c.APIRequest("POST", "remoteServices/chargeStart/v4", nil, bodyParams, true, true)
	if err != nil {
		return err
	}

	resultCode, ok := response["resultCode"].(string)
	if !ok || resultCode != "200S00" {
		return fmt.Errorf("Failed to start charging: result code %s", resultCode)
	}

	return nil
}

// ChargeStop stops charging the vehicle (EV/PHEV only)
func (c *Client) ChargeStop(internalVIN string) error {
	bodyParams := map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    internalVIN,
	}

	response, err := c.APIRequest("POST", "remoteServices/chargeStop/v4", nil, bodyParams, true, true)
	if err != nil {
		return err
	}

	resultCode, ok := response["resultCode"].(string)
	if !ok || resultCode != "200S00" {
		return fmt.Errorf("Failed to stop charging: result code %s", resultCode)
	}

	return nil
}

// HVACOn turns the vehicle HVAC system on
func (c *Client) HVACOn(internalVIN string) error {
	bodyParams := map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    internalVIN,
	}

	response, err := c.APIRequest("POST", "remoteServices/hvacOn/v4", nil, bodyParams, true, true)
	if err != nil {
		return err
	}

	resultCode, ok := response["resultCode"].(string)
	if !ok || resultCode != "200S00" {
		return fmt.Errorf("Failed to turn HVAC on: result code %s", resultCode)
	}

	return nil
}

// HVACOff turns the vehicle HVAC system off
func (c *Client) HVACOff(internalVIN string) error {
	bodyParams := map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    internalVIN,
	}

	response, err := c.APIRequest("POST", "remoteServices/hvacOff/v4", nil, bodyParams, true, true)
	if err != nil {
		return err
	}

	resultCode, ok := response["resultCode"].(string)
	if !ok || resultCode != "200S00" {
		return fmt.Errorf("Failed to turn HVAC off: result code %s", resultCode)
	}

	return nil
}

// RefreshVehicleStatus requests the vehicle to refresh its status
func (c *Client) RefreshVehicleStatus(internalVIN string) error {
	bodyParams := map[string]interface{}{
		"internaluserid": "__INTERNAL_ID__",
		"internalvin":    internalVIN,
	}

	response, err := c.APIRequest("POST", "remoteServices/refreshVehicleStatus/v4", nil, bodyParams, true, true)
	if err != nil {
		return err
	}

	resultCode, ok := response["resultCode"].(string)
	if !ok || resultCode != "200S00" {
		return fmt.Errorf("Failed to refresh vehicle status: result code %s", resultCode)
	}

	return nil
}

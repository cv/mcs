package api

import "fmt"

// executeControl sends a control command to the vehicle and validates the response.
func (c *Client) executeControl(endpoint, actionDesc, internalVIN string) error {
	bodyParams := map[string]interface{}{
		"internaluserid": InternalUserID,
		"internalvin":    internalVIN,
	}

	response, err := c.APIRequest("POST", endpoint, nil, bodyParams, true, true)
	if err != nil {
		return err
	}

	resultCode, ok := response["resultCode"].(string)
	if !ok || resultCode != "200S00" {
		return fmt.Errorf("failed to %s: result code %s", actionDesc, resultCode)
	}

	return nil
}

// DoorLock locks the vehicle doors
func (c *Client) DoorLock(internalVIN string) error {
	return c.executeControl("remoteServices/doorLock/v4", "lock doors", internalVIN)
}

// DoorUnlock unlocks the vehicle doors
func (c *Client) DoorUnlock(internalVIN string) error {
	return c.executeControl("remoteServices/doorUnlock/v4", "unlock doors", internalVIN)
}

// LightsOn turns the vehicle hazard lights on
func (c *Client) LightsOn(internalVIN string) error {
	return c.executeControl("remoteServices/lightOn/v4", "turn lights on", internalVIN)
}

// LightsOff turns the vehicle hazard lights off
func (c *Client) LightsOff(internalVIN string) error {
	return c.executeControl("remoteServices/lightOff/v4", "turn lights off", internalVIN)
}

// EngineStart starts the vehicle engine remotely
func (c *Client) EngineStart(internalVIN string) error {
	return c.executeControl("remoteServices/engineStart/v4", "start engine", internalVIN)
}

// EngineStop stops the vehicle engine remotely
func (c *Client) EngineStop(internalVIN string) error {
	return c.executeControl("remoteServices/engineStop/v4", "stop engine", internalVIN)
}

// ChargeStart starts charging the vehicle (EV/PHEV only)
func (c *Client) ChargeStart(internalVIN string) error {
	return c.executeControl("remoteServices/chargeStart/v4", "start charging", internalVIN)
}

// ChargeStop stops charging the vehicle (EV/PHEV only)
func (c *Client) ChargeStop(internalVIN string) error {
	return c.executeControl("remoteServices/chargeStop/v4", "stop charging", internalVIN)
}

// HVACOn turns the vehicle HVAC system on
func (c *Client) HVACOn(internalVIN string) error {
	return c.executeControl("remoteServices/hvacOn/v4", "turn HVAC on", internalVIN)
}

// HVACOff turns the vehicle HVAC system off
func (c *Client) HVACOff(internalVIN string) error {
	return c.executeControl("remoteServices/hvacOff/v4", "turn HVAC off", internalVIN)
}

// RefreshVehicleStatus requests the vehicle to refresh its status (PHEV/EV only)
func (c *Client) RefreshVehicleStatus(internalVIN string) error {
	return c.executeControl("remoteServices/activeRealTimeVehicleStatus/v4", "refresh vehicle status", internalVIN)
}

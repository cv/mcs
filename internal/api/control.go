package api

import (
	"context"
	"fmt"
)

// executeControl sends a control command to the vehicle and validates the response.
func (c *Client) executeControl(ctx context.Context, endpoint, actionDesc, internalVIN string) error {
	bodyParams := map[string]interface{}{
		"internaluserid": InternalUserID,
		"internalvin":    internalVIN,
	}

	response, err := c.APIRequest(ctx, "POST", endpoint, nil, bodyParams, true, true)
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
func (c *Client) DoorLock(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, "remoteServices/doorLock/v4", "lock doors", internalVIN)
}

// DoorUnlock unlocks the vehicle doors
func (c *Client) DoorUnlock(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, "remoteServices/doorUnlock/v4", "unlock doors", internalVIN)
}

// LightsOn turns the vehicle hazard lights on
func (c *Client) LightsOn(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, "remoteServices/lightOn/v4", "turn lights on", internalVIN)
}

// LightsOff turns the vehicle hazard lights off
func (c *Client) LightsOff(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, "remoteServices/lightOff/v4", "turn lights off", internalVIN)
}

// EngineStart starts the vehicle engine remotely
func (c *Client) EngineStart(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, "remoteServices/engineStart/v4", "start engine", internalVIN)
}

// EngineStop stops the vehicle engine remotely
func (c *Client) EngineStop(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, "remoteServices/engineStop/v4", "stop engine", internalVIN)
}

// ChargeStart starts charging the vehicle (EV/PHEV only)
func (c *Client) ChargeStart(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, "remoteServices/chargeStart/v4", "start charging", internalVIN)
}

// ChargeStop stops charging the vehicle (EV/PHEV only)
func (c *Client) ChargeStop(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, "remoteServices/chargeStop/v4", "stop charging", internalVIN)
}

// HVACOn turns the vehicle HVAC system on
func (c *Client) HVACOn(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, "remoteServices/hvacOn/v4", "turn HVAC on", internalVIN)
}

// HVACOff turns the vehicle HVAC system off
func (c *Client) HVACOff(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, "remoteServices/hvacOff/v4", "turn HVAC off", internalVIN)
}

// RefreshVehicleStatus requests the vehicle to refresh its status (PHEV/EV only)
func (c *Client) RefreshVehicleStatus(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, "remoteServices/activeRealTimeVehicleStatus/v4", "refresh vehicle status", internalVIN)
}

// SetHVACSetting sets HVAC temperature and defroster settings
// tempUnit should be "c" for Celsius or "f" for Fahrenheit
func (c *Client) SetHVACSetting(ctx context.Context, internalVIN string, temperature float64, tempUnit string, frontDefroster, rearDefroster bool) error {
	var tempType int
	switch tempUnit {
	case "c", "C":
		tempType = 1
	case "f", "F":
		tempType = 2
	default:
		return fmt.Errorf("invalid temperature unit: %s (must be 'c' or 'f')", tempUnit)
	}

	frontDefrost := 0
	if frontDefroster {
		frontDefrost = 1
	}
	rearDefrost := 0
	if rearDefroster {
		rearDefrost = 1
	}

	bodyParams := map[string]interface{}{
		"internaluserid":  InternalUserID,
		"internalvin":     internalVIN,
		"Temperature":     temperature,
		"TemperatureType": tempType,
		"FrontDefroster":  frontDefrost,
		"RearDefogger":    rearDefrost,
	}

	response, err := c.APIRequest(ctx, "POST", "remoteServices/updateHVACSetting/v4", nil, bodyParams, true, true)
	if err != nil {
		return err
	}

	resultCode, ok := response["resultCode"].(string)
	if !ok || resultCode != "200S00" {
		return fmt.Errorf("failed to set HVAC settings: result code %s", resultCode)
	}

	return nil
}

package api

import (
	"context"
	"fmt"
)

// Control endpoint constants
const (
	EndpointDoorLock             = "remoteServices/doorLock/v4"
	EndpointDoorUnlock           = "remoteServices/doorUnlock/v4"
	EndpointLightOn              = "remoteServices/lightOn/v4"
	EndpointLightOff             = "remoteServices/lightOff/v4"
	EndpointEngineStart          = "remoteServices/engineStart/v4"
	EndpointEngineStop           = "remoteServices/engineStop/v4"
	EndpointChargeStart          = "remoteServices/chargeStart/v4"
	EndpointChargeStop           = "remoteServices/chargeStop/v4"
	EndpointHVACOn               = "remoteServices/hvacOn/v4"
	EndpointHVACOff              = "remoteServices/hvacOff/v4"
	EndpointRefreshVehicleStatus = "remoteServices/activeRealTimeVehicleStatus/v4"
	EndpointUpdateHVACSetting    = "remoteServices/updateHVACSetting/v4"
)

// boolToInt converts a boolean to an integer (true=1, false=0)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

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

	resultCode, ok := getString(response, "resultCode")
	if !ok || resultCode != ResultCodeSuccess {
		return fmt.Errorf("failed to %s: result code %s", actionDesc, resultCode)
	}

	return nil
}

// DoorLock locks the vehicle doors
func (c *Client) DoorLock(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, EndpointDoorLock, "lock doors", internalVIN)
}

// DoorUnlock unlocks the vehicle doors
func (c *Client) DoorUnlock(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, EndpointDoorUnlock, "unlock doors", internalVIN)
}

// LightsOn turns the vehicle hazard lights on
func (c *Client) LightsOn(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, EndpointLightOn, "turn lights on", internalVIN)
}

// LightsOff turns the vehicle hazard lights off
func (c *Client) LightsOff(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, EndpointLightOff, "turn lights off", internalVIN)
}

// EngineStart starts the vehicle engine remotely
func (c *Client) EngineStart(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, EndpointEngineStart, "start engine", internalVIN)
}

// EngineStop stops the vehicle engine remotely
func (c *Client) EngineStop(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, EndpointEngineStop, "stop engine", internalVIN)
}

// ChargeStart starts charging the vehicle (EV/PHEV only)
func (c *Client) ChargeStart(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, EndpointChargeStart, "start charging", internalVIN)
}

// ChargeStop stops charging the vehicle (EV/PHEV only)
func (c *Client) ChargeStop(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, EndpointChargeStop, "stop charging", internalVIN)
}

// HVACOn turns the vehicle HVAC system on
func (c *Client) HVACOn(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, EndpointHVACOn, "turn HVAC on", internalVIN)
}

// HVACOff turns the vehicle HVAC system off
func (c *Client) HVACOff(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, EndpointHVACOff, "turn HVAC off", internalVIN)
}

// RefreshVehicleStatus requests the vehicle to refresh its status (PHEV/EV only)
func (c *Client) RefreshVehicleStatus(ctx context.Context, internalVIN string) error {
	return c.executeControl(ctx, EndpointRefreshVehicleStatus, "refresh vehicle status", internalVIN)
}

// SetHVACSetting sets HVAC temperature and defroster settings
func (c *Client) SetHVACSetting(ctx context.Context, internalVIN string, temperature float64, tempUnit TemperatureUnit, frontDefroster, rearDefroster bool) error {
	bodyParams := map[string]interface{}{
		"internaluserid":  InternalUserID,
		"internalvin":     internalVIN,
		"Temperature":     temperature,
		"TemperatureType": int(tempUnit),
		"FrontDefroster":  boolToInt(frontDefroster),
		"RearDefogger":    boolToInt(rearDefroster),
	}

	response, err := c.APIRequest(ctx, "POST", EndpointUpdateHVACSetting, nil, bodyParams, true, true)
	if err != nil {
		return err
	}

	resultCode, ok := getString(response, "resultCode")
	if !ok || resultCode != ResultCodeSuccess {
		return fmt.Errorf("failed to set HVAC settings: result code %s", resultCode)
	}

	return nil
}

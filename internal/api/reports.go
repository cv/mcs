package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// Endpoints for the reads that have no typed response yet: the health report
// (oil life, maintenance, warnings) and the remote climate setting.
// paramInternalVIN is the request field naming the vehicle.
const paramInternalVIN = "internalvin"

const (
	EndpointGetHealthReport = "remoteServices/getHealthReport/v4"
	EndpointGetHVACSetting  = "remoteServices/getHVACSetting/v4"
)

// GetHealthReport returns the vehicle's latest health report as the API sends it.
func (c *Client) GetHealthReport(ctx context.Context, internalVIN string) (map[string]any, error) {
	body := map[string]any{
		"internaluserid": "__INTERNAL_ID__",
		paramInternalVIN: internalVIN,
		"limit":          1,
		"offset":         0,
	}

	return c.rawRead(ctx, EndpointGetHealthReport, body, "get health report")
}

// GetHVACSetting returns the remote climate setting (temperature, defrosters).
func (c *Client) GetHVACSetting(ctx context.Context, internalVIN string) (map[string]any, error) {
	body := map[string]any{
		"internaluserid": "__INTERNAL_ID__",
		paramInternalVIN: internalVIN,
	}

	return c.rawRead(ctx, EndpointGetHVACSetting, body, "get HVAC setting")
}

func (c *Client) rawRead(ctx context.Context, endpoint string, body map[string]any, desc string) (map[string]any, error) {
	responseBytes, err := c.APIRequestJSON(ctx, "POST", endpoint, nil, body, true, true)
	if err != nil {
		return nil, err
	}

	var out map[string]any
	if err := json.Unmarshal(responseBytes, &out); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	code, _ := out["resultCode"].(string)
	if err := checkResultCode(code, desc); err != nil {
		return nil, err
	}

	return out, nil
}

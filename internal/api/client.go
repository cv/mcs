package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	// MaxRetries is the maximum number of retries for API requests
	MaxRetries = 4
)

// APIRequest makes an API request with proper encryption, signing, and error handling
func (c *Client) APIRequest(ctx context.Context, method, uri string, queryParams map[string]string, bodyParams map[string]interface{}, needsKeys, needsAuth bool) (map[string]interface{}, error) {
	return c.apiRequestWithRetry(ctx, method, uri, queryParams, bodyParams, needsKeys, needsAuth, 0)
}

func (c *Client) apiRequestWithRetry(ctx context.Context, method, uri string, queryParams map[string]string, bodyParams map[string]interface{}, needsKeys, needsAuth bool, retryCount int) (map[string]interface{}, error) {
	if retryCount > MaxRetries {
		return nil, NewAPIError("Request exceeded max number of retries")
	}

	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if needsKeys {
		if err := c.ensureKeysPresent(ctx); err != nil {
			return nil, err
		}
	}

	if needsAuth {
		if err := c.ensureTokenValid(ctx); err != nil {
			return nil, err
		}
	}

	response, err := c.sendAPIRequest(ctx, method, uri, queryParams, bodyParams, needsKeys, needsAuth)
	if err != nil {
		// Handle retryable errors
		switch err.(type) {
		case *EncryptionError:
			// Retrieve new encryption keys and retry
			if err := c.GetEncryptionKeys(ctx); err != nil {
				return nil, fmt.Errorf("failed to retrieve encryption keys: %w", err)
			}
			return c.apiRequestWithRetry(ctx, method, uri, queryParams, bodyParams, needsKeys, needsAuth, retryCount+1)
		case *TokenExpiredError:
			// Login again and retry
			if err := c.Login(ctx); err != nil {
				return nil, fmt.Errorf("failed to login: %w", err)
			}
			return c.apiRequestWithRetry(ctx, method, uri, queryParams, bodyParams, needsKeys, needsAuth, retryCount+1)
		default:
			return nil, err
		}
	}

	return response, nil
}

func (c *Client) sendAPIRequest(ctx context.Context, method, uri string, queryParams map[string]string, bodyParams map[string]interface{}, needsKeys, needsAuth bool) (map[string]interface{}, error) {
	timestamp := getTimestampStrMs()

	// Prepare query parameters (encrypted if provided)
	originalQueryStr := ""
	encryptedQueryParams := url.Values{}
	if len(queryParams) > 0 {
		queryValues := url.Values{}
		for k, v := range queryParams {
			queryValues.Add(k, v)
		}
		originalQueryStr = queryValues.Encode()

		encrypted, err := c.encryptPayloadUsingKey(originalQueryStr)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt query params: %w", err)
		}
		encryptedQueryParams.Add("params", encrypted)
	}

	// Prepare body (encrypted if provided)
	originalBodyStr := ""
	encryptedBody := ""
	if len(bodyParams) > 0 {
		bodyJSON, err := json.Marshal(bodyParams)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body params: %w", err)
		}
		originalBodyStr = string(bodyJSON)

		encrypted, err := c.encryptPayloadUsingKey(originalBodyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt body: %w", err)
		}
		encryptedBody = encrypted
	}

	// Build URL
	requestURL := c.baseURL + uri
	if len(encryptedQueryParams) > 0 {
		requestURL += "?" + encryptedQueryParams.Encode()
	}

	// Create request with context
	var req *http.Request
	var err error
	if encryptedBody != "" {
		req, err = http.NewRequestWithContext(ctx, method, requestURL, bytes.NewBufferString(encryptedBody))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, requestURL, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Generate sensor data
	sensorData, err := c.sensorDataBuilder.GenerateSensorData()
	if err != nil {
		return nil, fmt.Errorf("failed to generate sensor data: %w", err)
	}

	// Set headers
	headers := map[string]string{
		"device-id":          c.baseAPIDeviceID,
		"app-code":           c.appCode,
		"app-os":             AppOS,
		"user-agent":         UserAgentBaseAPI,
		"app-version":        AppVersion,
		"app-unique-id":      AppPackageID,
		"req-id":             "req_" + timestamp,
		"timestamp":          timestamp,
		"Content-Type":       "application/json",
		"X-acf-sensor-data":  sensorData,
	}

	if needsAuth {
		headers["access-token"] = c.accessToken
	} else {
		headers["access-token"] = ""
	}

	// Calculate signature
	if uri == "service/checkVersion" {
		headers["sign"] = c.getSignFromTimestamp(timestamp)
	} else if method == "GET" {
		headers["sign"] = c.getSignFromPayloadAndTimestamp(originalQueryStr, timestamp)
	} else if method == "POST" {
		headers["sign"] = c.getSignFromPayloadAndTimestamp(originalBodyStr, timestamp)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	c.logRequest(method, requestURL, headers, originalBodyStr)

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	c.logResponse(resp.StatusCode, body)

	var response APIBaseResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check response state
	if response.State == "S" {
		// Success - decrypt payload
		if response.Payload == "" {
			return nil, fmt.Errorf("payload not found in response")
		}

		return c.decryptPayloadUsingKey(response.Payload)
	}

	// Handle errors
	switch int(response.ErrorCode) {
	case 600001:
		return nil, NewEncryptionError()
	case 600002:
		return nil, NewTokenExpiredError()
	case 920000:
		if response.ExtraCode == "400S01" {
			return nil, NewRequestInProgressError()
		} else if response.ExtraCode == "400S11" {
			return nil, NewEngineStartLimitError()
		}
	}

	// Generic error
	if response.Message != "" {
		return nil, NewAPIError(fmt.Sprintf("Request failed: %s", response.Message))
	}
	if response.Error != "" {
		return nil, NewAPIError(fmt.Sprintf("Request failed: %s", response.Error))
	}

	return nil, NewAPIError("Request failed for an unknown reason")
}

// ensureKeysPresent ensures encryption keys are available
func (c *Client) ensureKeysPresent(ctx context.Context) error {
	if c.encKey == "" || c.signKey == "" {
		return c.GetEncryptionKeys(ctx)
	}
	return nil
}

// ensureTokenValid ensures access token is valid
func (c *Client) ensureTokenValid(ctx context.Context) error {
	if !c.IsTokenValid() {
		return c.Login(ctx)
	}
	return nil
}

// encryptPayloadUsingKey encrypts a payload using the client's encryption key
func (c *Client) encryptPayloadUsingKey(payload string) (string, error) {
	if c.encKey == "" {
		return "", NewAPIError("Missing encryption key")
	}
	if payload == "" {
		return "", nil
	}
	return EncryptAES128CBC([]byte(payload), c.encKey, IV)
}

// decryptPayloadUsingKey decrypts a payload using the client's encryption key
func (c *Client) decryptPayloadUsingKey(payload string) (map[string]interface{}, error) {
	if c.encKey == "" {
		return nil, NewAPIError("Missing encryption key")
	}

	decrypted, err := DecryptAES128CBC(payload, c.encKey, IV)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt payload: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(decrypted, &result); err != nil {
		return nil, fmt.Errorf("failed to parse decrypted payload: %w", err)
	}

	return result, nil
}

// getSignFromPayloadAndTimestamp generates a signature from payload and timestamp
func (c *Client) getSignFromPayloadAndTimestamp(payload, timestamp string) string {
	if timestamp == "" {
		return ""
	}
	if c.signKey == "" {
		return ""
	}

	encryptedPayload, _ := c.encryptPayloadUsingKey(payload)
	timestampExtended := timestamp + timestamp[6:] + timestamp[3:]
	dataToSign := encryptedPayload + timestampExtended + c.signKey

	return SignWithSHA256(dataToSign)
}

// logRequest logs request details when debug mode is enabled
func (c *Client) logRequest(method, url string, headers map[string]string, body string) {
	if !c.debug {
		return
	}

	fmt.Printf("DEBUG: %s %s\n", method, url)
	fmt.Printf("DEBUG: Headers:\n")
	for k, v := range headers {
		if k == "access-token" && v != "" {
			fmt.Printf("  %s: [REDACTED]\n", k)
		} else {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}
	if body != "" {
		fmt.Printf("DEBUG: Original body: %s\n", body)
	}
}

// logResponse logs response details when debug mode is enabled
func (c *Client) logResponse(statusCode int, body []byte) {
	if !c.debug {
		return
	}

	fmt.Printf("DEBUG: Response status: %d\n", statusCode)
	fmt.Printf("DEBUG: Response body: %s\n", string(body))
}

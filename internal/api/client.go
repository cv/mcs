package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	// MaxRetries is the maximum number of retries for API requests.
	MaxRetries = 4
)

// calculateBackoff returns the backoff duration for a given retry count.
// Uses exponential backoff: 1s, 2s, 4s, 8s.
func calculateBackoff(retryCount int) time.Duration {
	if retryCount <= 0 {
		return 0
	}
	// 2^(retryCount-1) seconds, capped at 8 seconds
	backoffSeconds := min(1<<(retryCount-1), 8)

	return time.Duration(backoffSeconds) * time.Second
}

// sleepWithContext sleeps for the specified duration, but returns early if context is cancelled.
func sleepWithContext(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// APIRequest makes an API request with proper encryption, signing, and error handling.
func (c *Client) APIRequest(ctx context.Context, method, uri string, queryParams map[string]string, bodyParams map[string]any, needsKeys, needsAuth bool) (map[string]any, error) {
	return c.apiRequestWithRetry(ctx, method, uri, queryParams, bodyParams, needsKeys, needsAuth, 0)
}

// APIRequestJSON makes an API request and returns the raw decrypted JSON bytes.
func (c *Client) APIRequestJSON(ctx context.Context, method, uri string, queryParams map[string]string, bodyParams map[string]any, needsKeys, needsAuth bool) ([]byte, error) {
	return c.apiRequestJSONWithRetry(ctx, method, uri, queryParams, bodyParams, needsKeys, needsAuth, 0)
}

// retryFunc is the type for functions that can be retried.
type retryFunc[T any] func(ctx context.Context, method, uri string, queryParams map[string]string, bodyParams map[string]any, needsKeys, needsAuth bool) (T, error)

// handleRetryableError attempts to recover from an encryption or token error by refreshing credentials.
// Returns true if the error was handled and a retry should be attempted.
func handleRetryableError[T any](
	ctx context.Context,
	c *Client,
	err error,
	retryCount int,
) (shouldRetry bool, retryErr error) {
	var encErr *EncryptionError
	var tokenErr *TokenExpiredError

	if errors.As(err, &encErr) {
		// Retrieve new encryption keys and retry
		if err := c.GetEncryptionKeys(ctx); err != nil {
			return false, fmt.Errorf("failed to retrieve encryption keys: %w", err)
		}
		// Apply backoff delay before retry
		backoff := calculateBackoff(retryCount + 1)
		if err := c.sleepFunc(ctx, backoff); err != nil {
			return false, err
		}

		return true, nil
	}

	if errors.As(err, &tokenErr) {
		// Login again and retry
		if err := c.Login(ctx); err != nil {
			return false, fmt.Errorf("failed to login: %w", err)
		}
		// Apply backoff delay before retry
		backoff := calculateBackoff(retryCount + 1)
		if err := c.sleepFunc(ctx, backoff); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, nil
}

// genericRetry implements the retry logic with exponential backoff for API requests.
// It handles encryption errors and token expiration by refreshing credentials and retrying.
func genericRetry[T any](
	ctx context.Context,
	c *Client,
	method, uri string,
	queryParams map[string]string,
	bodyParams map[string]any,
	needsKeys, needsAuth bool,
	retryCount int,
	executeFunc retryFunc[T],
) (T, error) {
	var zero T // zero value for type T

	if retryCount > MaxRetries {
		return zero, NewAPIError("Request exceeded max number of retries")
	}

	// Check for context cancellation
	if err := ctx.Err(); err != nil {
		return zero, err
	}

	if needsKeys {
		if err := c.ensureKeysPresent(ctx); err != nil {
			return zero, err
		}
	}

	if needsAuth {
		if err := c.ensureTokenValid(ctx); err != nil {
			return zero, err
		}
	}

	response, err := executeFunc(ctx, method, uri, queryParams, bodyParams, needsKeys, needsAuth)
	if err != nil {
		// Handle retryable errors
		shouldRetry, retryErr := handleRetryableError[T](ctx, c, err, retryCount)
		if retryErr != nil {
			return zero, retryErr
		}
		if shouldRetry {
			return genericRetry(ctx, c, method, uri, queryParams, bodyParams, needsKeys, needsAuth, retryCount+1, executeFunc)
		}

		return zero, err
	}

	return response, nil
}

func (c *Client) apiRequestWithRetry(ctx context.Context, method, uri string, queryParams map[string]string, bodyParams map[string]any, needsKeys, needsAuth bool, retryCount int) (map[string]any, error) {
	return genericRetry(ctx, c, method, uri, queryParams, bodyParams, needsKeys, needsAuth, retryCount, c.sendAPIRequest)
}

func (c *Client) apiRequestJSONWithRetry(ctx context.Context, method, uri string, queryParams map[string]string, bodyParams map[string]any, needsKeys, needsAuth bool, retryCount int) ([]byte, error) {
	return genericRetry(ctx, c, method, uri, queryParams, bodyParams, needsKeys, needsAuth, retryCount, c.sendAPIRequestJSON)
}

// handleAPIResponse processes the API response and returns the encrypted payload or an error.
// It centralizes error handling logic for all API responses.
func handleAPIResponse(response *APIBaseResponse) (string, error) {
	// Check response state
	if response.State == "S" {
		// Success - return encrypted payload for caller to decrypt
		if response.Payload == "" {
			return "", errors.New("payload not found in response")
		}

		return response.Payload, nil
	}

	// Handle errors
	switch int(response.ErrorCode) {
	case ErrorCodeEncryption:
		return "", NewEncryptionError()
	case ErrorCodeTokenExpired:
		return "", NewTokenExpiredError()
	case ErrorCodeRequestIssue:
		switch response.ExtraCode {
		case ExtraCodeRequestInProgress:
			return "", NewRequestInProgressError()
		case ExtraCodeEngineStartLimit:
			return "", NewEngineStartLimitError()
		}
	}

	// Generic error
	if response.Message != "" {
		return "", NewAPIError("Request failed: " + response.Message)
	}
	if response.Error != "" {
		return "", NewAPIError("Request failed: " + response.Error)
	}

	return "", NewAPIError("Request failed for an unknown reason")
}

// preparedParams holds the prepared and encrypted request parameters.
type preparedParams struct {
	originalQueryStr     string
	encryptedQueryParams url.Values
	originalBodyStr      string
	encryptedBody        string
}

// prepareRequestParams encrypts query and body parameters for an API request.
func (c *Client) prepareRequestParams(queryParams map[string]string, bodyParams map[string]any) (preparedParams, error) {
	var params preparedParams

	// Prepare query parameters (encrypted if provided)
	if len(queryParams) > 0 {
		queryValues := url.Values{}
		for k, v := range queryParams {
			queryValues.Add(k, v)
		}
		params.originalQueryStr = queryValues.Encode()

		encrypted, err := c.encryptPayloadUsingKey(params.originalQueryStr)
		if err != nil {
			return params, fmt.Errorf("failed to encrypt query params: %w", err)
		}
		params.encryptedQueryParams = url.Values{}
		params.encryptedQueryParams.Add("params", encrypted)
	}

	// Prepare body (encrypted if provided)
	if len(bodyParams) > 0 {
		bodyJSON, err := json.Marshal(bodyParams)
		if err != nil {
			return params, fmt.Errorf("failed to marshal body params: %w", err)
		}
		params.originalBodyStr = string(bodyJSON)

		encrypted, err := c.encryptPayloadUsingKey(params.originalBodyStr)
		if err != nil {
			return params, fmt.Errorf("failed to encrypt body: %w", err)
		}
		params.encryptedBody = encrypted
	}

	return params, nil
}

// calculateSignature determines the appropriate signature for the request.
func (c *Client) calculateSignature(method, uri, originalQueryStr, originalBodyStr, timestamp string) string {
	switch {
	case uri == EndpointCheckVersion:
		return c.getSignFromTimestamp(timestamp)
	case method == http.MethodGet:
		return c.getSignFromPayloadAndTimestamp(originalQueryStr, timestamp)
	case method == http.MethodPost:
		return c.getSignFromPayloadAndTimestamp(originalBodyStr, timestamp)
	default:
		return ""
	}
}

// buildHTTPRequest creates an HTTP request with all necessary headers.
func (c *Client) buildHTTPRequest(ctx context.Context, method, uri, timestamp string, params preparedParams, needsAuth bool) (*http.Request, error) {
	// Build URL
	requestURL := c.baseURL + uri
	if len(params.encryptedQueryParams) > 0 {
		requestURL += "?" + params.encryptedQueryParams.Encode()
	}

	// Create request with context
	var req *http.Request
	var err error
	if params.encryptedBody != "" {
		req, err = http.NewRequestWithContext(ctx, method, requestURL, bytes.NewBufferString(params.encryptedBody))
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
	accessToken := ""
	if needsAuth {
		accessToken = c.accessToken
	}

	headers := map[string]string{
		"device-id":         c.baseAPIDeviceID,
		"app-code":          c.appCode,
		"app-os":            AppOS,
		"user-agent":        UserAgentBaseAPI,
		"app-version":       AppVersion,
		"app-unique-id":     AppPackageID,
		"req-id":            "req_" + timestamp,
		"timestamp":         timestamp,
		"Content-Type":      "application/json",
		"X-acf-sensor-data": sensorData,
		"access-token":      accessToken,
		"sign":              c.calculateSignature(method, uri, params.originalQueryStr, params.originalBodyStr, timestamp),
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	c.logRequest(method, requestURL, headers, params.originalBodyStr)

	return req, nil
}

// executeAPIRequest handles the common logic for making API requests.
// It returns the encrypted payload string on success, or an error.
func (c *Client) executeAPIRequest(ctx context.Context, method, uri string, queryParams map[string]string, bodyParams map[string]any, needsAuth bool) (string, error) {
	timestamp := getTimestampStrMs()

	// Prepare and encrypt parameters
	params, err := c.prepareRequestParams(queryParams, bodyParams)
	if err != nil {
		return "", err
	}

	// Build HTTP request with headers
	req, err := c.buildHTTPRequest(ctx, method, uri, timestamp, params, needsAuth)
	if err != nil {
		return "", err
	}

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	c.logResponse(resp.StatusCode, body)

	var response APIBaseResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return handleAPIResponse(&response)
}

func (c *Client) sendAPIRequest(ctx context.Context, method, uri string, queryParams map[string]string, bodyParams map[string]any, _, needsAuth bool) (map[string]any, error) {
	encryptedPayload, err := c.executeAPIRequest(ctx, method, uri, queryParams, bodyParams, needsAuth)
	if err != nil {
		return nil, err
	}

	return c.decryptPayloadUsingKey(encryptedPayload)
}

func (c *Client) sendAPIRequestJSON(ctx context.Context, method, uri string, queryParams map[string]string, bodyParams map[string]any, _, needsAuth bool) ([]byte, error) {
	encryptedPayload, err := c.executeAPIRequest(ctx, method, uri, queryParams, bodyParams, needsAuth)
	if err != nil {
		return nil, err
	}

	return c.decryptPayloadBytes(encryptedPayload)
}

// ensureKeysPresent ensures encryption keys are available.
func (c *Client) ensureKeysPresent(ctx context.Context) error {
	if c.Keys.EncKey == "" || c.Keys.SignKey == "" {
		return c.GetEncryptionKeys(ctx)
	}

	return nil
}

// ensureTokenValid ensures access token is valid.
func (c *Client) ensureTokenValid(ctx context.Context) error {
	if !c.IsTokenValid() {
		return c.Login(ctx)
	}

	return nil
}

// encryptPayloadUsingKey encrypts a payload using the client's encryption key.
func (c *Client) encryptPayloadUsingKey(payload string) (string, error) {
	if c.Keys.EncKey == "" {
		return "", NewAPIError("Missing encryption key")
	}
	if payload == "" {
		return "", nil
	}

	return EncryptAES128CBC([]byte(payload), c.Keys.EncKey, IV)
}

// decryptPayloadUsingKey decrypts a payload using the client's encryption key.
func (c *Client) decryptPayloadUsingKey(payload string) (map[string]any, error) {
	if c.Keys.EncKey == "" {
		return nil, NewAPIError("Missing encryption key")
	}

	decrypted, err := DecryptAES128CBC(payload, c.Keys.EncKey, IV)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt payload: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(decrypted, &result); err != nil {
		return nil, fmt.Errorf("failed to parse decrypted payload: %w", err)
	}

	return result, nil
}

// decryptPayloadBytes decrypts a payload and returns raw JSON bytes.
func (c *Client) decryptPayloadBytes(payload string) ([]byte, error) {
	if c.Keys.EncKey == "" {
		return nil, NewAPIError("Missing encryption key")
	}

	decrypted, err := DecryptAES128CBC(payload, c.Keys.EncKey, IV)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt payload: %w", err)
	}

	return decrypted, nil
}

// getSignFromPayloadAndTimestamp generates a signature from payload and timestamp.
func (c *Client) getSignFromPayloadAndTimestamp(payload, timestamp string) string {
	if timestamp == "" {
		return ""
	}
	if c.Keys.SignKey == "" {
		return ""
	}

	encryptedPayload, _ := c.encryptPayloadUsingKey(payload)
	timestampExtended := timestamp + timestamp[6:] + timestamp[3:]
	dataToSign := encryptedPayload + timestampExtended + c.Keys.SignKey

	return SignWithSHA256(dataToSign)
}

// logRequest logs request details when debug mode is enabled.
func (c *Client) logRequest(method, url string, headers map[string]string, body string) {
	if !c.debug {
		return
	}

	fmt.Fprintf(os.Stderr, "DEBUG: %s %s\n", method, url)
	fmt.Fprintf(os.Stderr, "DEBUG: Headers:\n")
	for k, v := range headers {
		if k == "access-token" && v != "" {
			fmt.Fprintf(os.Stderr, "  %s: [REDACTED]\n", k)
		} else {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", k, v)
		}
	}
	if body != "" {
		fmt.Fprintf(os.Stderr, "DEBUG: Original body: %s\n", body)
	}
}

// logResponse logs response details when debug mode is enabled.
func (c *Client) logResponse(statusCode int, body []byte) {
	if !c.debug {
		return
	}

	fmt.Fprintf(os.Stderr, "DEBUG: Response status: %d\n", statusCode)
	fmt.Fprintf(os.Stderr, "DEBUG: Response body: %s\n", string(body))
}

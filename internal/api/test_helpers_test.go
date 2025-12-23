package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testEncKey is the encryption key used in tests (must be 16 bytes for AES-128)
const testEncKey = "testenckey123456"

// testSignKey is the signing key used in tests
const testSignKey = "testsignkey12345"

// createTestClient creates a test API client with mock credentials
func createTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	client, err := NewClient("test@example.com", "password", RegionMNAO)
	require.NoError(t, err, "Failed to create client: %v")
	client.baseURL = serverURL + "/"
	client.Keys.EncKey = testEncKey
	client.Keys.SignKey = testSignKey
	client.accessToken = "test-token"
	client.accessTokenExpirationTs = 9999999999
	return client
}

// testServerOptions configures the behavior of a test server
type testServerOptions struct {
	expectedPath   string
	expectedMethod string
	validateBody   bool
}

// TestServerOption is a functional option for configuring test servers
type TestServerOption func(*testServerOptions)

// WithPath validates that the request path matches the expected value
func WithPath(path string) TestServerOption {
	return func(opts *testServerOptions) {
		opts.expectedPath = path
	}
}

// WithMethod validates that the request method matches the expected value
func WithMethod(method string) TestServerOption {
	return func(opts *testServerOptions) {
		opts.expectedMethod = method
	}
}

// WithBodyValidation ensures the request has a non-empty body
func WithBodyValidation() TestServerOption {
	return func(opts *testServerOptions) {
		opts.validateBody = true
	}
}

// createTestServer creates a flexible test server that returns encrypted JSON responses
// Use the functional options to configure path, method, and body validation
func createTestServer(t *testing.T, responseData map[string]interface{}, options ...TestServerOption) *httptest.Server {
	t.Helper()

	opts := &testServerOptions{}
	for _, option := range options {
		option(opts)
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate path if specified
		if opts.expectedPath != "" {
			assert.Equal(t, opts.expectedPath, r.URL.Path)
		}

		// Validate method if specified
		if opts.expectedMethod != "" {
			assert.Equal(t, opts.expectedMethod, r.Method)
		}

		// Validate body if requested
		if opts.validateBody {
			body, _ := io.ReadAll(r.Body)
			assert.NotEmpty(t, body, "Expected non-empty body")
		}

		// Encrypt and wrap the response
		responseJSON, _ := json.Marshal(responseData)
		encrypted, _ := EncryptAES128CBC(responseJSON, testEncKey, IV)

		response := map[string]interface{}{
			"state":   "S",
			"payload": encrypted,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)

	}))
}

// createSuccessServer creates a test server that returns an encrypted success response with path validation
// This is a convenience wrapper around createTestServer for simple GET/retrieve endpoints
// For control endpoints (POST with body validation), use createControlTestServer instead
func createSuccessServer(t *testing.T, expectedPath string, responseData map[string]interface{}) *httptest.Server {
	t.Helper()
	return createTestServer(t, responseData, WithPath(expectedPath))
}

// createErrorServer creates a test server that returns an encrypted error response
// Optionally accepts HTTP status code as third parameter (defaults to 200)
func createErrorServer(t *testing.T, resultCode, message string, httpStatusCode ...int) *httptest.Server {
	t.Helper()
	errorResponse := map[string]interface{}{
		"resultCode": resultCode,
		"message":    message,
	}

	statusCode := 200
	if len(httpStatusCode) > 0 {
		statusCode = httpStatusCode[0]
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Encrypt and wrap the response
		responseJSON, _ := json.Marshal(errorResponse)
		encrypted, _ := EncryptAES128CBC(responseJSON, testEncKey, IV)

		response := map[string]interface{}{
			"state":   "S",
			"payload": encrypted,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(response)

	}))
}

// createControlTestServer creates a test server for control endpoints
// All control endpoints expect POST requests with non-empty bodies and return standard success responses
func createControlTestServer(t *testing.T, expectedPath string) *httptest.Server {
	t.Helper()
	successResponse := map[string]interface{}{
		"resultCode": "200S00",
		"message":    "Success",
	}

	return createTestServer(t, successResponse,
		WithPath(expectedPath),
		WithMethod("POST"),
		WithBodyValidation(),
	)
}

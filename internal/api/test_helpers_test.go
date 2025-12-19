package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// testEncKey is the encryption key used in tests (must be 16 bytes for AES-128)
const testEncKey = "testenckey123456"

// testSignKey is the signing key used in tests
const testSignKey = "testsignkey12345"

// createTestClient creates a test API client with mock credentials
func createTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.baseURL = serverURL + "/"
	client.encKey = testEncKey
	client.signKey = testSignKey
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
		if opts.expectedPath != "" && r.URL.Path != opts.expectedPath {
			t.Errorf("Expected path %s, got %s", opts.expectedPath, r.URL.Path)
		}

		// Validate method if specified
		if opts.expectedMethod != "" && r.Method != opts.expectedMethod {
			t.Errorf("Expected method %s, got %s", opts.expectedMethod, r.Method)
		}

		// Validate body if requested
		if opts.validateBody {
			body, _ := io.ReadAll(r.Body)
			if len(body) == 0 {
				t.Error("Expected non-empty body")
			}
		}

		// Encrypt and wrap the response
		responseJSON, _ := json.Marshal(responseData)
		encrypted, _ := EncryptAES128CBC(responseJSON, testEncKey, IV)

		response := map[string]interface{}{
			"state":   "S",
			"payload": encrypted,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
}

// createSuccessServer creates a test server that returns an encrypted success response
func createSuccessServer(t *testing.T, expectedPath string, responseData map[string]interface{}) *httptest.Server {
	t.Helper()
	return createTestServer(t, responseData, WithPath(expectedPath))
}

// createErrorServer creates a test server that returns an encrypted error response
func createErrorServer(t *testing.T, resultCode, message string) *httptest.Server {
	t.Helper()
	errorResponse := map[string]interface{}{
		"resultCode": resultCode,
		"message":    message,
	}
	return createTestServer(t, errorResponse)
}

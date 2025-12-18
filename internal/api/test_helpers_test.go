package api

import (
	"encoding/json"
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

// createSuccessServer creates a test server that returns an encrypted success response
func createSuccessServer(t *testing.T, expectedPath string, responseData map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if expectedPath != "" && r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

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

// createErrorServer creates a test server that returns an encrypted error response
func createErrorServer(t *testing.T, resultCode, message string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		testResponse := map[string]interface{}{
			"resultCode": resultCode,
			"message":    message,
		}

		responseJSON, _ := json.Marshal(testResponse)
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

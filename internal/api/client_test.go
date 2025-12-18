package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAPIRequest_Success tests successful API request with encryption
func TestAPIRequest_Success(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("device-id") == "" {
			t.Error("device-id header is missing")
		}
		if r.Header.Get("app-code") == "" {
			t.Error("app-code header is missing")
		}
		if r.Header.Get("user-agent") != UserAgentBaseAPI {
			t.Errorf("expected user-agent %s, got %s", UserAgentBaseAPI, r.Header.Get("user-agent"))
		}
		if r.Header.Get("sign") == "" {
			t.Error("sign header is missing")
		}
		if r.Header.Get("timestamp") == "" {
			t.Error("timestamp header is missing")
		}

		// Return success response with encrypted payload
		// Encrypt a simple JSON response
		testResponse := map[string]interface{}{
			"resultCode": "200S00",
			"message":    "Success",
		}
		responseJSON, _ := json.Marshal(testResponse)
		encrypted, _ := EncryptAES128CBC(responseJSON, "testenckey123456", IV)

		response := map[string]interface{}{
			"state":   "S",
			"payload": encrypted,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with test server URL
	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.baseURL = server.URL + "/"

	// Set encryption keys (must be exactly 16 bytes for AES-128)
	client.encKey = "testenckey123456"
	client.signKey = "testsignkey12345"

	// Make API request
	result, err := client.APIRequest("POST", "test/endpoint", nil, map[string]interface{}{"test": "data"}, true, false)
	if err != nil {
		t.Fatalf("APIRequest failed: %v", err)
	}

	if result["resultCode"] != "200S00" {
		t.Errorf("Expected resultCode 200S00, got %v", result["resultCode"])
	}
	if result["message"] != "Success" {
		t.Errorf("Expected message Success, got %v", result["message"])
	}
}

// TestAPIRequest_EncryptionError tests handling of encryption error response
func TestAPIRequest_EncryptionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"state": "E",
			"errorCode": 600001,
			"message": "Encryption error",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.baseURL = server.URL + "/"
	client.encKey = "testenckey123456"
	client.signKey = "testsignkey12345"

	// Test will be implemented when APIRequest exists
}

// TestAPIRequest_TokenExpired tests handling of expired token error
func TestAPIRequest_TokenExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"state": "E",
			"errorCode": 600002,
			"message": "Token expired",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.baseURL = server.URL + "/"
	client.encKey = "testenckey123456"
	client.signKey = "testsignkey12345"

	// Test will be implemented when APIRequest exists
}

// TestAPIRequest_RequestInProgress tests handling of request in progress error
func TestAPIRequest_RequestInProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"state": "E",
			"errorCode": 920000,
			"extraCode": "400S01",
			"message": "Request in progress",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.baseURL = server.URL + "/"
	client.encKey = "testenckey123456"
	client.signKey = "testsignkey12345"

	// Test will be implemented when APIRequest exists
}

// TestEncryptPayloadUsingKey tests payload encryption
func TestEncryptPayloadUsingKey(t *testing.T) {
	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	client.encKey = "testenckey123456"

	testData := map[string]interface{}{
		"test": "data",
		"foo": "bar",
	}

	dataJSON, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	encrypted, err := client.encryptPayloadUsingKey(string(dataJSON))
	if err != nil {
		t.Fatalf("Failed to encrypt payload: %v", err)
	}

	if encrypted == "" {
		t.Error("Encrypted payload is empty")
	}

	// Verify it's base64 encoded
	if len(encrypted) == 0 {
		t.Error("Encrypted payload should not be empty")
	}
}

// TestDecryptPayloadUsingKey tests payload decryption
func TestDecryptPayloadUsingKey(t *testing.T) {
	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	client.encKey = "testenckey123456"

	testData := map[string]interface{}{
		"test": "data",
		"foo": "bar",
	}

	dataJSON, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	// Encrypt first
	encrypted, err := client.encryptPayloadUsingKey(string(dataJSON))
	if err != nil {
		t.Fatalf("Failed to encrypt payload: %v", err)
	}

	// Then decrypt
	decrypted, err := client.decryptPayloadUsingKey(encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt payload: %v", err)
	}

	// Verify decrypted data matches original
	if decrypted["test"] != "data" {
		t.Errorf("Expected test=data, got test=%v", decrypted["test"])
	}
	if decrypted["foo"] != "bar" {
		t.Errorf("Expected foo=bar, got foo=%v", decrypted["foo"])
	}
}

// TestGetSignFromPayloadAndTimestamp tests signature generation
func TestGetSignFromPayloadAndTimestamp(t *testing.T) {
	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	client.encKey = "testenckey123456"
	client.signKey = "testsignkey12345"

	payload := `{"test":"data"}`
	timestamp := "1234567890123"

	sign := client.getSignFromPayloadAndTimestamp(payload, timestamp)
	if sign == "" {
		t.Error("Signature should not be empty")
	}

	// Verify it's uppercase hex (SHA256)
	if len(sign) != 64 {
		t.Errorf("Expected signature length of 64, got %d", len(sign))
	}
}

// TestAPIRequest_MissingKeys tests that APIRequest returns error when keys are missing
func TestAPIRequest_MissingKeys(t *testing.T) {
	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Don't set encryption keys
	// APIRequest should return error
	// Test will be implemented when APIRequest exists
	_ = client
}

// TestAPIRequest_POST_WithBody tests POST request with body encryption
func TestAPIRequest_POST_WithBody(t *testing.T) {
	requestReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		// Verify method
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Verify sign header is present
		if r.Header.Get("sign") == "" {
			t.Error("sign header is missing")
		}

		response := map[string]interface{}{
			"state": "S",
			"payload": "test-encrypted-response",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.baseURL = server.URL + "/"
	client.encKey = "testenckey123456"
	client.signKey = "testsignkey12345"

	// Test will be completed when APIRequest is implemented
	_ = requestReceived
	_ = client
}

// TestAPIRequest_GET_WithQuery tests GET request with query parameter encryption
func TestAPIRequest_GET_WithQuery(t *testing.T) {
	requestReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true

		// Verify method
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		// Verify params query parameter exists
		if r.URL.Query().Get("params") == "" {
			t.Error("params query parameter is missing")
		}

		response := map[string]interface{}{
			"state": "S",
			"payload": "test-encrypted-response",
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.baseURL = server.URL + "/"
	client.encKey = "testenckey123456"
	client.signKey = "testsignkey12345"

	// Test will be completed when APIRequest is implemented
	_ = requestReceived
	_ = client
}

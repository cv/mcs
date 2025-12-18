package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestAPIRequest_RetryOnEncryptionError tests that encryption errors trigger retry with new keys
func TestAPIRequest_RetryOnEncryptionError(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		var response map[string]interface{}
		if requestCount == 1 {
			// First request: return encryption error
			response = map[string]interface{}{
				"state":     "E",
				"errorCode": 600001,
				"message":   "Encryption error",
			}
		} else if requestCount == 2 && r.URL.Path == "/service/checkVersion" {
			// Second request: return new keys
			testResponse := map[string]interface{}{
				"encKey":  "newtestenckey123",
				"signKey": "newtestsignkey12",
			}
			responseJSON, _ := json.Marshal(testResponse)
			// Use the app code derived key for encryption
			client := &Client{appCode: "202007270941270111799"}
			key := client.getDecryptionKeyFromAppCode()
			encrypted, _ := EncryptAES128CBC(responseJSON, key, IV)

			response = map[string]interface{}{
				"state":   "S",
				"payload": encrypted,
			}
		} else {
			// Subsequent request: return success
			testResponse := map[string]interface{}{
				"resultCode": "200S00",
				"message":    "Success",
			}
			responseJSON, _ := json.Marshal(testResponse)
			encrypted, _ := EncryptAES128CBC(responseJSON, "newtestenckey123", IV)

			response = map[string]interface{}{
				"state":   "S",
				"payload": encrypted,
			}
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
	client.encKey = "oldtestenckey123"
	client.signKey = "oldtestsignkey12"

	// Make API request - should retry after encryption error
	result, err := client.APIRequest("POST", "test/endpoint", nil, map[string]interface{}{"test": "data"}, true, false)
	if err != nil {
		t.Fatalf("APIRequest failed: %v", err)
	}

	if result["resultCode"] != "200S00" {
		t.Errorf("Expected resultCode 200S00, got %v", result["resultCode"])
	}

	// Verify that retry occurred (3 requests: error, get keys, retry)
	if requestCount != 3 {
		t.Errorf("Expected 3 requests (error + get keys + retry), got %d", requestCount)
	}
}

// TestAPIRequest_MaxRetries tests that max retries is enforced
func TestAPIRequest_MaxRetries(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		// Always return encryption error
		response := map[string]interface{}{
			"state":     "E",
			"errorCode": 600001,
			"message":   "Encryption error",
		}

		if r.URL.Path == "/service/checkVersion" {
			// Return keys to allow retry
			testResponse := map[string]interface{}{
				"encKey":  "newtestenckey123",
				"signKey": "newtestsignkey12",
			}
			responseJSON, _ := json.Marshal(testResponse)
			client := &Client{appCode: "202007270941270111799"}
			key := client.getDecryptionKeyFromAppCode()
			encrypted, _ := EncryptAES128CBC(responseJSON, key, IV)

			response = map[string]interface{}{
				"state":   "S",
				"payload": encrypted,
			}
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

	// Make API request - should fail after max retries
	_, err = client.APIRequest("POST", "test/endpoint", nil, map[string]interface{}{"test": "data"}, true, false)
	if err == nil {
		t.Fatal("Expected error due to max retries, got nil")
	}

	if err.Error() != "Request exceeded max number of retries" {
		t.Errorf("Expected max retries error, got: %v", err)
	}
}

// TestAPIRequest_EngineStartLimitError tests the engine start limit error
func TestAPIRequest_EngineStartLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"state":     "E",
			"errorCode": 920000,
			"extraCode": "400S11",
			"message":   "Engine start limit",
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

	_, err = client.APIRequest("POST", "test/endpoint", nil, map[string]interface{}{"test": "data"}, true, false)
	if err == nil {
		t.Fatal("Expected engine start limit error, got nil")
	}

	if _, ok := err.(*EngineStartLimitError); !ok {
		t.Errorf("Expected EngineStartLimitError, got %T", err)
	}
}

// TestAPIRequest_WithQueryParams tests GET request with query parameters
func TestAPIRequest_WithQueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify params query parameter is present
		if r.URL.Query().Get("params") == "" {
			t.Error("Expected params query parameter to be present")
		}

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

	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	client.baseURL = server.URL + "/"
	client.encKey = "testenckey123456"
	client.signKey = "testsignkey12345"

	result, err := client.APIRequest("GET", "test/endpoint", map[string]string{"foo": "bar", "baz": "qux"}, nil, true, false)
	if err != nil {
		t.Fatalf("APIRequest failed: %v", err)
	}

	if result["resultCode"] != "200S00" {
		t.Errorf("Expected resultCode 200S00, got %v", result["resultCode"])
	}
}

// TestEncryptPayloadUsingKey_EmptyPayload tests encryption of empty payload
func TestEncryptPayloadUsingKey_EmptyPayload(t *testing.T) {
	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	client.encKey = "testenckey123456"

	encrypted, err := client.encryptPayloadUsingKey("")
	if err != nil {
		t.Fatalf("Failed to encrypt empty payload: %v", err)
	}

	if encrypted != "" {
		t.Error("Expected empty encrypted payload for empty input")
	}
}

// TestEncryptPayloadUsingKey_MissingKey tests error when encryption key is missing
func TestEncryptPayloadUsingKey_MissingKey(t *testing.T) {
	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Don't set encryption key
	_, err = client.encryptPayloadUsingKey("test data")
	if err == nil {
		t.Fatal("Expected error when encryption key is missing, got nil")
	}

	if _, ok := err.(*APIError); !ok {
		t.Errorf("Expected APIError, got %T", err)
	}
}

// TestDecryptPayloadUsingKey_MissingKey tests error when decryption key is missing
func TestDecryptPayloadUsingKey_MissingKey(t *testing.T) {
	client, err := NewClient("test@example.com", "password", "MNAO")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Don't set encryption key
	_, err = client.decryptPayloadUsingKey("test data")
	if err == nil {
		t.Fatal("Expected error when decryption key is missing, got nil")
	}

	if _, ok := err.(*APIError); !ok {
		t.Errorf("Expected APIError, got %T", err)
	}
}

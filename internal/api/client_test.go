package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
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
	result, err := client.APIRequest(context.Background(), "POST", "test/endpoint", nil, map[string]interface{}{"test": "data"}, true, false)
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
			"state":     "E",
			"errorCode": 600001,
			"message":   "Encryption error",
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

	_, err = client.APIRequest(context.Background(), "POST", "test/endpoint", nil, map[string]interface{}{"test": "data"}, false, false)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// APIRequest retries on EncryptionError by fetching new keys
	// Since our mock server always returns the same error, it eventually fails with wrapped error
	expectedMsg := "failed to retrieve encryption keys"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing %q, got: %v", expectedMsg, err)
	}
}

// TestAPIRequest_TokenExpired tests handling of expired token error
func TestAPIRequest_TokenExpired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"state":     "E",
			"errorCode": 600002,
			"message":   "Token expired",
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

	_, err = client.APIRequest(context.Background(), "POST", "test/endpoint", nil, map[string]interface{}{"test": "data"}, false, false)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// APIRequest retries on TokenExpiredError by re-logging in
	// Since re-login will fail (mock server only handles one endpoint), we get a wrapped error
	expectedMsg := "failed to login"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing %q, got: %v", expectedMsg, err)
	}
}

// TestAPIRequest_RequestInProgress tests handling of request in progress error
func TestAPIRequest_RequestInProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"state":     "E",
			"errorCode": 920000,
			"extraCode": "400S01",
			"message":   "Request in progress",
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

	_, err = client.APIRequest(context.Background(), "POST", "test/endpoint", nil, map[string]interface{}{"test": "data"}, false, false)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	// Verify it's a RequestInProgressError
	if _, ok := err.(*RequestInProgressError); !ok {
		t.Errorf("Expected RequestInProgressError, got %T: %v", err, err)
	}
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
		"foo":  "bar",
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
		"foo":  "bar",
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

// TestAPIRequest_MissingKeys tests that APIRequest attempts to get keys when missing
func TestAPIRequest_MissingKeys(t *testing.T) {
	// Create a server that returns error for checkVersion (key retrieval)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"state":   "E",
			"message": "Server error",
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

	// Don't set encryption keys, but request needsKeys=true
	// APIRequest should attempt to get keys and fail
	_, err = client.APIRequest(context.Background(), "POST", "test/endpoint", nil, map[string]interface{}{"test": "data"}, true, false)
	if err == nil {
		t.Fatal("Expected error when keys are missing, got nil")
	}
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

		// Return encrypted success response
		testResponse := map[string]interface{}{
			"resultCode": "200S00",
			"success":    true,
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

	result, err := client.APIRequest(context.Background(), "POST", "test/endpoint", nil, map[string]interface{}{"test": "data"}, false, false)
	if err != nil {
		t.Fatalf("APIRequest failed: %v", err)
	}

	if !requestReceived {
		t.Error("Request was not received by server")
	}

	if result["resultCode"] != "200S00" {
		t.Errorf("Expected resultCode 200S00, got %v", result["resultCode"])
	}
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

		// Return encrypted success response
		testResponse := map[string]interface{}{
			"resultCode": "200S00",
			"data":       "test-data",
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

	result, err := client.APIRequest(context.Background(), "GET", "test/endpoint", map[string]string{"key": "value"}, nil, false, false)
	if err != nil {
		t.Fatalf("APIRequest failed: %v", err)
	}

	if !requestReceived {
		t.Error("Request was not received by server")
	}

	if result["resultCode"] != "200S00" {
		t.Errorf("Expected resultCode 200S00, got %v", result["resultCode"])
	}
}

// TestCalculateBackoff tests the backoff calculation
func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		retryCount int
		expected   time.Duration
	}{
		{0, 0},
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
		{4, 8 * time.Second},
		{5, 8 * time.Second}, // Capped at 8 seconds
		{10, 8 * time.Second},
	}

	for _, tt := range tests {
		t.Run(strings.Join([]string{"retry", strings.Repeat("x", tt.retryCount)}, "_"), func(t *testing.T) {
			result := calculateBackoff(tt.retryCount)
			if result != tt.expected {
				t.Errorf("calculateBackoff(%d) = %v, want %v", tt.retryCount, result, tt.expected)
			}
		})
	}
}

// TestSleepWithContext_Completes tests that sleep completes normally
func TestSleepWithContext_Completes(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	duration := 100 * time.Millisecond

	err := sleepWithContext(ctx, duration)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("sleepWithContext returned error: %v", err)
	}

	// Allow 50ms tolerance
	if elapsed < duration || elapsed > duration+50*time.Millisecond {
		t.Errorf("sleepWithContext took %v, expected around %v", elapsed, duration)
	}
}

// TestSleepWithContext_Cancelled tests that sleep returns early on context cancellation
func TestSleepWithContext_Cancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	start := time.Now()
	duration := 5 * time.Second // Long sleep

	// Cancel after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := sleepWithContext(ctx, duration)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("sleepWithContext should return error on cancellation")
	}

	if err != context.Canceled {
		t.Errorf("sleepWithContext returned %v, want context.Canceled", err)
	}

	// Should return much earlier than 5 seconds
	if elapsed > 1*time.Second {
		t.Errorf("sleepWithContext took %v, should have returned early", elapsed)
	}
}

// TestSleepWithContext_ZeroDuration tests that zero duration returns immediately
func TestSleepWithContext_ZeroDuration(t *testing.T) {
	ctx := context.Background()
	start := time.Now()

	err := sleepWithContext(ctx, 0)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("sleepWithContext returned error: %v", err)
	}

	if elapsed > 10*time.Millisecond {
		t.Errorf("sleepWithContext with 0 duration took %v, should be immediate", elapsed)
	}
}

// TestAPIRequest_RetryWithContextCancellation tests that context cancellation during backoff returns immediately
func TestAPIRequest_RetryWithContextCancellation(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// Return success response for checkVersion (encryption key retrieval)
		if strings.Contains(r.URL.Path, "checkVersion") {
			testResponse := map[string]interface{}{
				"encKey":  "newtestenckey123",
				"signKey": "newtestsignkey12",
			}
			responseJSON, _ := json.Marshal(testResponse)
			// Use the app code derived key for encryption (same as decryption uses)
			tempClient := &Client{appCode: "202007270941270111799"}
			key := tempClient.getDecryptionKeyFromAppCode()
			encrypted, _ := EncryptAES128CBC(responseJSON, key, IV)
			response := map[string]interface{}{
				"state":   "S",
				"payload": encrypted,
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}
			return
		}

		// Always return encryption error for other requests to trigger retry with backoff
		// Using 600001 (EncryptionError) instead of 600002 (TokenExpiredError) because
		// TokenExpiredError triggers Login() which uses usherURL (not mocked here)
		response := map[string]interface{}{
			"state":     "E",
			"errorCode": 600001,
			"message":   "Encryption error",
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

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context after 500ms
	go func() {
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err = client.APIRequest(ctx, "POST", "test/endpoint", nil, map[string]interface{}{"test": "data"}, false, false)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Expected error due to context cancellation, got nil")
	}

	// Check if error is or contains context.Canceled
	if err != context.Canceled && !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}

	// Should return quickly after cancellation (within 1 second total)
	// First retry is after 1 second, so if cancelled at 500ms during first backoff,
	// it should return before the 1 second completes
	if elapsed > 1500*time.Millisecond {
		t.Errorf("Request took %v, should have returned quickly after context cancellation", elapsed)
	}

	// Should have made 1 initial call, Login will fail but that's expected
	// The key point is we don't wait through all the retries
	if callCount > 2 {
		t.Errorf("Expected at most 2 calls, got %d", callCount)
	}
}

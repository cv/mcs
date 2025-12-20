package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetEncryptionKeys(t *testing.T) {
	// Create a test client to get the correct decryption key
	testClient := &Client{
		email:    "test@example.com",
		password: "password",
		region:   RegionMNAO,
		appCode:  "202007270941270111799",
	}
	decryptionKey := testClient.getDecryptionKeyFromAppCode()

	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/prod/"+EndpointCheckVersion {
			t.Errorf("Expected path /prod/"+EndpointCheckVersion+", got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Verify headers
		if r.Header.Get("app-code") == "" {
			t.Error("Expected app-code header")
		}

		// Mock response - encrypt a simple JSON payload
		response := map[string]interface{}{
			"encKey":  "test-enc-key-123",
			"signKey": "test-sign-key-456",
		}
		responseJSON, _ := json.Marshal(response)

		// Encrypt the response payload using the app code derived key
		encrypted, _ := EncryptAES128CBC(responseJSON, decryptionKey, "0102030405060708")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"state":   "S",
			"payload": encrypted,
		})
	}))
	defer server.Close()

	client := &Client{
		email:      "test@example.com",
		password:   "password",
		region:     RegionMNAO,
		baseURL:    server.URL + "/prod/",
		usherURL:   server.URL + "/appapi/v1/",
		appCode:    "202007270941270111799",
		httpClient: server.Client(),
	}
	client.baseAPIDeviceID = GenerateUUIDFromSeed(client.email)
	client.usherAPIDeviceID = GenerateUsherDeviceID(client.email)

	err := client.GetEncryptionKeys(context.Background())
	if err != nil {
		t.Fatalf("GetEncryptionKeys() error = %v", err)
	}

	if client.Keys.EncKey != "test-enc-key-123" {
		t.Errorf("Expected encKey='test-enc-key-123', got '%s'", client.Keys.EncKey)
	}
	if client.Keys.SignKey != "test-sign-key-456" {
		t.Errorf("Expected signKey='test-sign-key-456', got '%s'", client.Keys.SignKey)
	}
}

func TestClient_GetUsherEncryptionKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/appapi/v1/"+EndpointEncryptionKey {
			t.Errorf("Expected path /appapi/v1/"+EndpointEncryptionKey+", got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		// Verify query params
		appId := r.URL.Query().Get("appId")
		if appId != "MazdaApp" {
			t.Errorf("Expected appId=MazdaApp, got %s", appId)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"publicKey":     "test-public-key",
				"versionPrefix": "v1:",
			},
		})
	}))
	defer server.Close()

	client := &Client{
		email:      "test@example.com",
		region:     RegionMNAO,
		usherURL:   server.URL + "/appapi/v1/",
		httpClient: server.Client(),
	}
	client.usherAPIDeviceID = GenerateUsherDeviceID(client.email)

	pubKey, versionPrefix, err := client.GetUsherEncryptionKey(context.Background())
	if err != nil {
		t.Fatalf("GetUsherEncryptionKey() error = %v", err)
	}

	if pubKey != "test-public-key" {
		t.Errorf("Expected publicKey='test-public-key', got %s", pubKey)
	}
	if versionPrefix != "v1:" {
		t.Errorf("Expected versionPrefix='v1:', got %s", versionPrefix)
	}
}

func TestClient_Login(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/appapi/v1/" + EndpointEncryptionKey:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"publicKey":     "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAlVKZRa1pkk88B1ydifsFNEv/pOf854egpFu1HHf1wr3YKqmLSG1p39YhNqGLQzIDit1jTLz3MYAOeWiFQSz7h5hvMNccq76zh3Hsg93LurcKA9EmYoj9VsqUetk0evXoqOSGKXPgZosbGT0t8AW2CC7s8FeSPz2tH9T7zjvKQvdyS0BFrVFo1EUBa1UEdMfYW0jLsvLOCYP911X1zTlewV/sTQnAtiTHCrd3jfH2of8PYtTOsmfqCDdL476yGMgeHJ+ZXA/IX2beSrHXU0gCNc/agD+ScCZgpRjfptSbRtBHqtmU4IyF0eqQXCCcrcutjzSHg+3ppmB9x/YvhJvmGQIDAQAB",
					"versionPrefix": "v1:",
				},
			})

		case "/appapi/v1/" + EndpointLogin:
			// Parse request body
			var loginReq map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&loginReq)

			// Verify request structure
			if loginReq["appId"] != "MazdaApp" {
				t.Errorf("Expected appId=MazdaApp, got %s", loginReq["appId"])
			}
			if loginReq["userId"] != "test@example.com" {
				t.Errorf("Expected userId=test@example.com, got %s", loginReq["userId"])
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "OK",
				"data": map[string]interface{}{
					"accessToken":             "test-access-token-12345",
					"accessTokenExpirationTs": time.Now().Unix() + 3600,
				},
			})
		}
	}))
	defer server.Close()

	client := &Client{
		email:      "test@example.com",
		password:   "password123",
		region:     RegionMNAO,
		usherURL:   server.URL + "/appapi/v1/",
		httpClient: server.Client(),
	}
	client.usherAPIDeviceID = GenerateUsherDeviceID(client.email)

	err := client.Login(context.Background())
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if client.accessToken == "" {
		t.Error("Expected accessToken to be set")
	}
	if client.accessTokenExpirationTs == 0 {
		t.Error("Expected accessTokenExpirationTs to be set")
	}
}

func TestClient_IsTokenValid(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		expiration int64
		want       bool
	}{
		{
			name:       "no token",
			token:      "",
			expiration: 0,
			want:       false,
		},
		{
			name:       "expired token",
			token:      "token",
			expiration: time.Now().Unix() - 100,
			want:       false,
		},
		{
			name:       "valid token",
			token:      "token",
			expiration: time.Now().Unix() + 3600,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				accessToken:             tt.token,
				accessTokenExpirationTs: tt.expiration,
			}
			got := client.IsTokenValid()
			if got != tt.want {
				t.Errorf("IsTokenValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegion_String(t *testing.T) {
	tests := []struct {
		name   string
		region Region
		want   string
	}{
		{
			name:   "MNAO region",
			region: RegionMNAO,
			want:   "MNAO",
		},
		{
			name:   "MME region",
			region: RegionMME,
			want:   "MME",
		},
		{
			name:   "MJO region",
			region: RegionMJO,
			want:   "MJO",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.region.String()
			if got != tt.want {
				t.Errorf("Region.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRegion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Region
		wantErr bool
	}{
		{
			name:    "valid MNAO",
			input:   "MNAO",
			want:    RegionMNAO,
			wantErr: false,
		},
		{
			name:    "valid MME",
			input:   "MME",
			want:    RegionMME,
			wantErr: false,
		},
		{
			name:    "valid MJO",
			input:   "MJO",
			want:    RegionMJO,
			wantErr: false,
		},
		{
			name:    "invalid region",
			input:   "INVALID",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "lowercase mnao",
			input:   "mnao",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRegion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRegion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseRegion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegion_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		region Region
		want   bool
	}{
		{
			name:   "valid MNAO",
			region: RegionMNAO,
			want:   true,
		},
		{
			name:   "valid MME",
			region: RegionMME,
			want:   true,
		},
		{
			name:   "valid MJO",
			region: RegionMJO,
			want:   true,
		},
		{
			name:   "invalid empty",
			region: "",
			want:   false,
		},
		{
			name:   "invalid random",
			region: "INVALID",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.region.IsValid()
			if got != tt.want {
				t.Errorf("Region.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

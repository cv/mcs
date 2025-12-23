package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetEncryptionKeys(t *testing.T) {
	t.Parallel()
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
		assert.Equalf(t, "/prod/"+EndpointCheckVersion, r.URL.Path, "Expected path /prod/"+EndpointCheckVersion+", got %s", r.URL.Path)
		assert.Equalf(t, "POST", r.Method, "Expected POST method, got %s", r.Method)

		// Verify headers
		assert.NotEmpty(t, r.Header.Get("App-Code"), "Expected app-code header")

		// Mock response - encrypt a simple JSON payload
		response := map[string]any{
			"encKey":  "test-enc-key-123",
			"signKey": "test-sign-key-456",
		}
		responseJSON, _ := json.Marshal(response)

		// Encrypt the response payload using the app code derived key
		encrypted, _ := EncryptAES128CBC(responseJSON, decryptionKey, "0102030405060708")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
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
	require.NoError(t, err, "GetEncryptionKeys() error = %v")

	assert.Equalf(t, "test-enc-key-123", client.Keys.EncKey, "Expected encKey='test-enc-key-123', got '%s'", client.Keys.EncKey)
	assert.Equalf(t, "test-sign-key-456", client.Keys.SignKey, "Expected signKey='test-sign-key-456', got '%s'", client.Keys.SignKey)
}

func TestClient_GetUsherEncryptionKey(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equalf(t, "/appapi/v1/"+EndpointEncryptionKey, r.URL.Path, "Expected path /appapi/v1/"+EndpointEncryptionKey+", got %s", r.URL.Path)
		assert.Equalf(t, "GET", r.Method, "Expected GET method, got %s", r.Method)

		// Verify query params
		appId := r.URL.Query().Get("appId")
		assert.Equalf(t, "MazdaApp", appId, "Expected appId=MazdaApp, got %s", appId)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
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
	require.NoError(t, err, "GetUsherEncryptionKey() error = %v")

	assert.Equalf(t, "test-public-key", pubKey, "Expected publicKey='test-public-key', got %s", pubKey)
	assert.Equalf(t, "v1:", versionPrefix, "Expected versionPrefix='v1:', got %s", versionPrefix)
}

func TestClient_Login(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/appapi/v1/" + EndpointEncryptionKey:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"publicKey":     "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAlVKZRa1pkk88B1ydifsFNEv/pOf854egpFu1HHf1wr3YKqmLSG1p39YhNqGLQzIDit1jTLz3MYAOeWiFQSz7h5hvMNccq76zh3Hsg93LurcKA9EmYoj9VsqUetk0evXoqOSGKXPgZosbGT0t8AW2CC7s8FeSPz2tH9T7zjvKQvdyS0BFrVFo1EUBa1UEdMfYW0jLsvLOCYP911X1zTlewV/sTQnAtiTHCrd3jfH2of8PYtTOsmfqCDdL476yGMgeHJ+ZXA/IX2beSrHXU0gCNc/agD+ScCZgpRjfptSbRtBHqtmU4IyF0eqQXCCcrcutjzSHg+3ppmB9x/YvhJvmGQIDAQAB",
					"versionPrefix": "v1:",
				},
			})

		case "/appapi/v1/" + EndpointLogin:
			// Parse request body
			var loginReq map[string]any
			_ = json.NewDecoder(r.Body).Decode(&loginReq)

			// Verify request structure
			assert.EqualValuesf(t, "MazdaApp", loginReq["appId"], "Expected appId=MazdaApp, got %s", loginReq["appId"])
			assert.EqualValuesf(t, "test@example.com", loginReq["userId"], "Expected userId=test@example.com, got %s", loginReq["userId"])

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"status": "OK",
				"data": map[string]any{
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
	require.NoError(t, err, "Login() error = %v")

	assert.NotEmpty(t, client.accessToken, "Expected accessToken to be set")
	assert.NotEqual(t, 0, client.accessTokenExpirationTs, "Expected accessTokenExpirationTs to be set")
}

func TestClient_IsTokenValid(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			client := &Client{
				accessToken:             tt.token,
				accessTokenExpirationTs: tt.expiration,
			}
			got := client.IsTokenValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegion_String(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			got := tt.region.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseRegion(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			got, err := ParseRegion(tt.input)
			if tt.wantErr {
				require.Error(t, err, "ParseRegion() error = %v, wantErr %v")
			} else {
				require.NoError(t, err, "ParseRegion() error = %v, wantErr %v")
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegion_IsValid(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			got := tt.region.IsValid()
			assert.Equal(t, tt.want, got)
		})
	}
}

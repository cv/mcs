package api

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptAES128CBC(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		key     string
		iv      string
		wantErr bool
	}{
		{
			name:    "basic encryption",
			data:    []byte("Hello, World!"),
			key:     "1234567890123456", // 16 bytes for AES-128
			iv:      "0102030405060708", // 8 bytes IV
			wantErr: false,
		},
		{
			name:    "empty data",
			data:    []byte(""),
			key:     "1234567890123456",
			iv:      "0102030405060708",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := EncryptAES128CBC(tt.data, tt.key, tt.iv)
			if tt.wantErr {
				require.Error(t, err, "EncryptAES128CBC() error = %v, wantErr %v")
			} else {
				require.NoError(t, err, "EncryptAES128CBC() error = %v, wantErr %v")
			}

			if !tt.wantErr {
				// Should return base64 encoded string
				_, err := base64.StdEncoding.DecodeString(encrypted)
				require.NoErrorf(t, err, "EncryptAES128CBC() returned invalid base64: %v", err)
			}
		})
	}
}

func TestDecryptAES128CBC(t *testing.T) {
	key := "1234567890123456"
	iv := "0102030405060708"
	original := []byte("Hello, World!")

	// First encrypt
	encrypted, err := EncryptAES128CBC(original, key, iv)
	require.NoError(t, err, "Failed to encrypt: %v")

	// Then decrypt
	decrypted, err := DecryptAES128CBC(encrypted, key, iv)
	require.NoError(t, err, "DecryptAES128CBC() error = %v")

	assert.Equal(t, string(original), string(decrypted))
}

func TestGenerateUUIDFromSeed(t *testing.T) {
	tests := []struct {
		name string
		seed string
		want string
	}{
		{
			name: "test email seed",
			seed: "test@example.com",
			// SHA256 of "test@example.com" = 973dfe463ec85785f5f95af5ba3906eedb2d931c24e69824a89ea65dba4e813b
			want: "973DFE46-3EC8-5785-F5F9-5AF5BA3906EE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateUUIDFromSeed(tt.seed)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateUsherDeviceID(t *testing.T) {
	tests := []struct {
		name string
		seed string
		want string
	}{
		{
			name: "test email seed",
			seed: "test@example.com",
			// SHA256 of "test@example.com" = 973dfe463ec85785f5f95af5ba3906eedb2d931c24e69824a89ea65dba4e813b
			// First 8 hex chars = 973dfe46, decimal = 2537422406
			want: "ACCT2537422406",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateUsherDeviceID(tt.seed)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEncryptRSA(t *testing.T) {
	// Valid test RSA public key (base64 encoded DER format, 2048-bit)
	// Generated using: openssl genrsa 2048 | openssl rsa -pubout -outform DER | base64
	publicKeyBase64 := "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAlVKZRa1pkk88B1ydifsFNEv/pOf854egpFu1HHf1wr3YKqmLSG1p39YhNqGLQzIDit1jTLz3MYAOeWiFQSz7h5hvMNccq76zh3Hsg93LurcKA9EmYoj9VsqUetk0evXoqOSGKXPgZosbGT0t8AW2CC7s8FeSPz2tH9T7zjvKQvdyS0BFrVFo1EUBa1UEdMfYW0jLsvLOCYP911X1zTlewV/sTQnAtiTHCrd3jfH2of8PYtTOsmfqCDdL476yGMgeHJ+ZXA/IX2beSrHXU0gCNc/agD+ScCZgpRjfptSbRtBHqtmU4IyF0eqQXCCcrcutjzSHg+3ppmB9x/YvhJvmGQIDAQAB"

	tests := []struct {
		name    string
		data    string
		pubKey  string
		wantErr bool
	}{
		{
			name:    "basic encryption",
			data:    "password123",
			pubKey:  publicKeyBase64,
			wantErr: false,
		},
		{
			name:    "invalid public key",
			data:    "password123",
			pubKey:  "invalid-base64-!@#$",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := EncryptRSA(tt.data, tt.pubKey)
			if tt.wantErr {
				require.Error(t, err, "EncryptRSA() error = %v, wantErr %v")
			} else {
				require.NoError(t, err, "EncryptRSA() error = %v, wantErr %v")
			}

			if !tt.wantErr {
				assert.NotEmpty(t, encrypted)
			}

		})
	}
}

func TestSignWithMD5(t *testing.T) {
	data := "test data"
	signature := SignWithMD5(data)

	// MD5 hash should be 32 characters (hex encoded)
	assert.Lenf(t, signature, 32, "SignWithMD5() returned wrong length = %d, want 32", len(signature))

	// Should be uppercase
	assert.Equalf(t, "EB733A00C0C9D336E65691A37AB54293", signature, "SignWithMD5() = %s, want EB733A00C0C9D336E65691A37AB54293", signature)
}

func TestSignWithSHA256(t *testing.T) {
	data := "test data"
	signature := SignWithSHA256(data)

	// SHA256 hash should be 64 characters (hex encoded)
	assert.Lenf(t, signature, 64, "SignWithSHA256() returned wrong length = %d, want 64", len(signature))

	// Should be uppercase
	assert.Equalf(t, "916F0027A575074CE72A331777C3478D6513F786A591BD892DA1A577BF2335F9", signature, "SignWithSHA256() = %s, want 916F0027A575074CE72A331777C3478D6513F786A591BD892DA1A577BF2335F9", signature)
}

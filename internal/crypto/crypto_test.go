package crypto

import (
	"bytes"
	"crypto/aes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPKCS7Pad(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		blockSize int
		wantLen   int
	}{
		{
			name:      "empty data",
			data:      []byte{},
			blockSize: 16,
			wantLen:   16, // Full block of padding
		},
		{
			name:      "partial block",
			data:      []byte("hello"),
			blockSize: 16,
			wantLen:   16, // Padded to full block
		},
		{
			name:      "full block",
			data:      []byte("1234567890123456"),
			blockSize: 16,
			wantLen:   32, // Additional full block of padding
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PKCS7Pad(tt.data, tt.blockSize)
			assert.Equalf(t, tt.wantLen, len(result), "PKCS7Pad() length = %d, want %d")
			assert.Equal(t, 0, len(result)%tt.blockSize, "PKCS7Pad() result not multiple of block size")
		})
	}
}

func TestPKCS7Unpad(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    []byte
		wantErr bool
	}{
		{
			name:    "empty data",
			data:    []byte{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "valid padding",
			data:    []byte{'h', 'e', 'l', 'l', 'o', 11, 11, 11, 11, 11, 11, 11, 11, 11, 11, 11},
			want:    []byte("hello"),
			wantErr: false,
		},
		{
			name:    "invalid padding bytes",
			data:    []byte{'h', 'e', 'l', 'l', 'o', 1, 2, 3},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := PKCS7Unpad(tt.data)
			if tt.wantErr {
				require.Error(t, err, "PKCS7Unpad() error = %v, wantErr %v")
			} else {
				require.NoError(t, err, "PKCS7Unpad() error = %v, wantErr %v")
			}

			if !tt.wantErr {
				assert.Truef(t, bytes.Equal(result, tt.want), "PKCS7Unpad() = %v, want %v", result, tt.want)
			}

		})
	}
}

func TestPKCS7PadUnpad_RoundTrip(t *testing.T) {
	data := []byte("test data for padding")
	padded := PKCS7Pad(data, aes.BlockSize)
	unpadded, err := PKCS7Unpad(padded)
	require.NoError(t, err, "PKCS7Unpad() error = %v")
	assert.Truef(t, bytes.Equal(unpadded, data), "Round trip failed: got %v, want %v", unpadded, data)
}

func TestEncryptDecryptAES128CBC(t *testing.T) {
	key := []byte("0123456789abcdef") // 16 bytes for AES-128
	iv := []byte("abcdef0123456789")  // 16 bytes
	plaintext := []byte("Hello, World! This is a test message.")

	encrypted, err := EncryptAES128CBC(plaintext, key, iv)
	require.NoError(t, err, "EncryptAES128CBC() error = %v")

	decrypted, err := DecryptAES128CBC(encrypted, key, iv)
	require.NoError(t, err, "DecryptAES128CBC() error = %v")

	assert.Truef(t, bytes.Equal(decrypted, plaintext), "DecryptAES128CBC() = %v, want %v", decrypted, plaintext)
}

func TestEncryptAES128CBC_InvalidKey(t *testing.T) {
	key := []byte("short") // Invalid key length
	iv := []byte("abcdef0123456789")
	plaintext := []byte("test")

	_, err := EncryptAES128CBC(plaintext, key, iv)
	assert.Error(t, err, "EncryptAES128CBC() expected error for invalid key length")
}

func TestDecryptAES128CBC_InvalidCiphertext(t *testing.T) {
	key := []byte("0123456789abcdef")
	iv := []byte("abcdef0123456789")
	// Invalid ciphertext length (not multiple of block size)
	ciphertext := []byte("not valid")

	_, err := DecryptAES128CBC(ciphertext, key, iv)
	assert.Error(t, err, "DecryptAES128CBC() expected error for invalid ciphertext")
}

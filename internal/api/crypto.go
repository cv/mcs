package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

// EncryptAES128CBC encrypts data using AES-128-CBC and returns base64 encoded string
func EncryptAES128CBC(data []byte, key, iv string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Apply PKCS7 padding
	paddedData := pkcs7Pad(data, aes.BlockSize)

	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	mode.CryptBlocks(ciphertext, paddedData)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAES128CBC decrypts base64 encoded AES-128-CBC encrypted data
func DecryptAES128CBC(encryptedBase64, key, iv string) ([]byte, error) {
	encrypted, err := base64.StdEncoding.DecodeString(encryptedBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	if len(encrypted)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext is not a multiple of block size")
	}

	plaintext := make([]byte, len(encrypted))
	mode := cipher.NewCBCDecrypter(block, []byte(iv))
	mode.CryptBlocks(plaintext, encrypted)

	// Remove PKCS7 padding
	unpaddedData, err := pkcs7Unpad(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to unpad: %w", err)
	}

	return unpaddedData, nil
}

// EncryptRSA encrypts data using RSA-ECB-PKCS1 padding
func EncryptRSA(data, publicKeyBase64 string) ([]byte, error) {
	// Decode base64 public key
	publicKeyDER, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	// Parse DER encoded public key
	pubKey, err := x509.ParsePKIXPublicKey(publicKeyDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	// Encrypt using PKCS1v15
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPubKey, []byte(data))
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt: %w", err)
	}

	return encrypted, nil
}

// GenerateUUIDFromSeed generates a UUID from a seed string using SHA256
func GenerateUUIDFromSeed(seed string) string {
	hash := sha256.Sum256([]byte(seed))
	hexHash := strings.ToUpper(hex.EncodeToString(hash[:]))

	// Format as UUID: XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hexHash[0:8],
		hexHash[8:12],
		hexHash[12:16],
		hexHash[16:20],
		hexHash[20:32])
}

// GenerateUsherDeviceID generates a device ID from a seed string
func GenerateUsherDeviceID(seed string) string {
	hash := sha256.Sum256([]byte(seed))
	hexHash := strings.ToUpper(hex.EncodeToString(hash[:]))

	// Convert first 8 hex characters to decimal
	id, _ := strconv.ParseInt(hexHash[0:8], 16, 64)
	return fmt.Sprintf("ACCT%d", id)
}

// SignWithMD5 creates an MD5 hash of the data and returns uppercase hex string
func SignWithMD5(data string) string {
	hash := md5.Sum([]byte(data))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// SignWithSHA256 creates a SHA256 hash of the data and returns uppercase hex string
func SignWithSHA256(data string) string {
	hash := sha256.Sum256([]byte(data))
	return strings.ToUpper(hex.EncodeToString(hash[:]))
}

// pkcs7Pad applies PKCS7 padding to data
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

// pkcs7Unpad removes PKCS7 padding from data
func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("invalid padding: empty data")
	}

	padding := int(data[length-1])
	if padding > length || padding > aes.BlockSize {
		return nil, fmt.Errorf("invalid padding size")
	}

	// Verify all padding bytes are correct
	for i := 0; i < padding; i++ {
		if data[length-1-i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding bytes")
		}
	}

	return data[:length-padding], nil
}

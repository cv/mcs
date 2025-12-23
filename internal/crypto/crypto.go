package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
)

// PKCS7Pad applies PKCS7 padding to data.
func PKCS7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}

	return append(data, padtext...)
}

// PKCS7Unpad removes PKCS7 padding from data.
func PKCS7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("invalid padding: empty data")
	}

	padding := int(data[length-1])
	if padding > length || padding > aes.BlockSize {
		return nil, errors.New("invalid padding size")
	}

	// Verify all padding bytes are correct
	for i := range padding {
		if data[length-1-i] != byte(padding) {
			return nil, errors.New("invalid padding bytes")
		}
	}

	return data[:length-padding], nil
}

// EncryptAES128CBC encrypts data using AES-128-CBC.
func EncryptAES128CBC(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Apply PKCS7 padding
	paddedData := PKCS7Pad(data, aes.BlockSize)

	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedData)

	return ciphertext, nil
}

// DecryptAES128CBC decrypts AES-128-CBC encrypted data.
func DecryptAES128CBC(encrypted, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	if len(encrypted)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of block size")
	}

	plaintext := make([]byte, len(encrypted))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, encrypted)

	// Remove PKCS7 padding
	unpaddedData, err := PKCS7Unpad(plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to unpad: %w", err)
	}

	return unpaddedData, nil
}

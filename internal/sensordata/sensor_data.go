package sensordata

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	mathrand "math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/cv/mcs/internal/crypto"
)

const (
	sdkVersion   = "2.2.3"
	rsaPublicKey = "MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC4sA7vA7N/t1SRBS8tugM2X4bByl0jaCZLqxPOql+qZ3sP4UFayqJTvXjd7eTjMwg1T70PnmPWyh1hfQr4s12oSVphTKAjPiWmEBvcpnPPMjr5fGgv0w6+KM9DLTxcktThPZAGoVcoyM/cTO/YsAMIxlmTzpXBaxddHRwi8S2NvwIDAQAB"
)

// SensorDataBuilder builds sensor data for anti-bot protection
type SensorDataBuilder struct {
	sensorCollectionStartTimestamp time.Time
	deviceInfoTime                 int
	systemInfo                     *SystemInfo
	touchEventList                 *TouchEventList
	keyEventList                   *KeyEventList
	backgroundEventList            *BackgroundEventList
	performanceTestResults         *PerformanceTestResults
}

// NewSensorDataBuilder creates a new sensor data builder
func NewSensorDataBuilder() *SensorDataBuilder {
	builder := &SensorDataBuilder{
		sensorCollectionStartTimestamp: time.Now().UTC(),
		deviceInfoTime:                 mathrand.Intn(5)*1000 + 3000,
		systemInfo:                     NewSystemInfo(),
		touchEventList:                 NewTouchEventList(),
		keyEventList:                   NewKeyEventList(),
		backgroundEventList:            NewBackgroundEventList(),
		performanceTestResults:         NewPerformanceTestResults(),
	}

	builder.systemInfo.Randomize()
	builder.performanceTestResults.Randomize()

	return builder
}

// GenerateSensorData generates the sensor data string
func (b *SensorDataBuilder) GenerateSensorData() (string, error) {
	b.touchEventList.Randomize(b.sensorCollectionStartTimestamp)
	b.keyEventList.Randomize(b.sensorCollectionStartTimestamp)
	b.backgroundEventList.Randomize(b.sensorCollectionStartTimestamp)

	randomNumber := mathrand.Int31()

	// Orientation and motion event counts are always 0 since we don't generate these events
	orientationEventCount := 0
	motionEventCount := 0

	var sb strings.Builder
	sb.WriteString(sdkVersion)
	sb.WriteString("-1,2,-94,-100,")
	sb.WriteString(b.systemInfo.ToString())
	sb.WriteString(",")
	sb.WriteString(strconv.Itoa(b.systemInfo.GetCharCodeSum()))
	sb.WriteString(",")
	sb.WriteString(strconv.Itoa(int(randomNumber)))
	sb.WriteString(",")
	sb.WriteString(strconv.FormatInt(timestampToMillis(b.sensorCollectionStartTimestamp)/2, 10))
	sb.WriteString("-1,2,-94,-101,")
	sb.WriteString("do_en,dm_en,t_en")
	sb.WriteString("-1,2,-94,-102,")
	// Edited text is empty
	sb.WriteString("-1,2,-94,-108,")
	sb.WriteString(b.keyEventList.ToString())
	sb.WriteString("-1,2,-94,-117,")
	sb.WriteString(b.touchEventList.ToString())
	sb.WriteString("-1,2,-94,-111,")
	// Orientation event AA is empty
	sb.WriteString("-1,2,-94,-109,")
	// Motion event AA is empty
	sb.WriteString("-1,2,-94,-144,")
	// Orientation event AC is empty
	sb.WriteString("-1,2,-94,-142,")
	// Orientation event AB is empty
	sb.WriteString("-1,2,-94,-145,")
	// Motion event AC is empty
	sb.WriteString("-1,2,-94,-143,")
	// Motion event is empty
	sb.WriteString("-1,2,-94,-115,")
	sb.WriteString(b.generateMiscStat(orientationEventCount, motionEventCount))
	sb.WriteString("-1,2,-94,-106,")
	sb.WriteString(b.generateStoredValuesF())
	sb.WriteString(",")
	sb.WriteString(b.generateStoredValuesG())
	sb.WriteString("-1,2,-94,-120,")
	// Stored stack traces is empty
	sb.WriteString("-1,2,-94,-112,")
	sb.WriteString(b.performanceTestResults.ToString())
	sb.WriteString("-1,2,-94,-103,")
	sb.WriteString(b.backgroundEventList.ToString())

	return encryptSensorData(sb.String())
}

func (b *SensorDataBuilder) generateMiscStat(orientationDataCount, motionDataCount int) string {
	sumOfTextEventValues := b.keyEventList.GetSum()
	sumOfTouchEventTimestampsAndTypes := b.touchEventList.GetSum()
	orientationDataB := 0
	motionDataB := 0
	overallSum := sumOfTextEventValues + sumOfTouchEventTimestampsAndTypes + orientationDataB + motionDataB

	nowTimestamp := time.Now().UTC()
	timeSinceSensorCollectionStart := int(nowTimestamp.Sub(b.sensorCollectionStartTimestamp).Milliseconds())

	return fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d",
		sumOfTextEventValues,
		sumOfTouchEventTimestampsAndTypes,
		orientationDataB,
		motionDataB,
		overallSum,
		timeSinceSensorCollectionStart,
		len(b.keyEventList.keyEvents),
		len(b.touchEventList.touchEvents),
		orientationDataCount,
		motionDataCount,
		b.deviceInfoTime,
		mathrand.Intn(10)*1000+5000,
		0,
		feistelCipher(overallSum, len(b.keyEventList.keyEvents)+len(b.touchEventList.touchEvents)+orientationDataCount+motionDataCount, timeSinceSensorCollectionStart),
		timestampToMillis(b.sensorCollectionStartTimestamp),
		0,
	)
}

func (b *SensorDataBuilder) generateStoredValuesF() string {
	return "-1"
}

func (b *SensorDataBuilder) generateStoredValuesG() string {
	return "0"
}

func encryptSensorData(sensorData string) (string, error) {
	// Generate random keys
	aesKey := make([]byte, 16)
	if _, err := rand.Read(aesKey); err != nil {
		return "", fmt.Errorf("failed to generate AES key: %w", err)
	}

	aesIV := make([]byte, 16)
	if _, err := rand.Read(aesIV); err != nil {
		return "", fmt.Errorf("failed to generate AES IV: %w", err)
	}

	hmacKey := make([]byte, 32)
	if _, err := rand.Read(hmacKey); err != nil {
		return "", fmt.Errorf("failed to generate HMAC key: %w", err)
	}

	// Encrypt sensor data with AES
	encryptedBytes, err := crypto.EncryptAES128CBC([]byte(sensorData), aesKey, aesIV)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt sensor data: %w", err)
	}

	// Combine IV and encrypted data
	ivAndEncryptedData := append(aesIV, encryptedBytes...)

	// Calculate HMAC
	h := hmac.New(sha256.New, hmacKey)
	h.Write(ivAndEncryptedData)
	hmacResult := h.Sum(nil)

	// Combine all parts
	result := append(ivAndEncryptedData, hmacResult...)

	// Encrypt AES and HMAC keys with RSA
	publicKeyBytes, err := base64.StdEncoding.DecodeString(rsaPublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode RSA public key: %w", err)
	}

	pubKey, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("not an RSA public key")
	}

	encryptedAESKey, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPubKey, aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt AES key: %w", err)
	}

	encryptedHMACKey, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPubKey, hmacKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt HMAC key: %w", err)
	}

	// Generate random timestamps
	aesTimestamp := mathrand.Intn(3) * 1000
	hmacTimestamp := mathrand.Intn(3) * 1000
	base64Timestamp := mathrand.Intn(3) * 1000

	return fmt.Sprintf("1,a,%s,%s$%s$%d,%d,%d",
		base64.StdEncoding.EncodeToString(encryptedAESKey),
		base64.StdEncoding.EncodeToString(encryptedHMACKey),
		base64.StdEncoding.EncodeToString(result),
		aesTimestamp,
		hmacTimestamp,
		base64Timestamp,
	), nil
}

func countSeparators(s string) int {
	count := 0
	for _, char := range s {
		if char == ';' {
			count++
		}
	}
	return count
}

func timestampToMillis(t time.Time) int64 {
	return t.UnixMilli()
}

// feistelCipher implements a simplified Feistel network cipher for obfuscating sensor data metrics.
//
// A Feistel cipher is a symmetric structure used in block cipher designs (e.g., DES).
// It splits data into two halves and iteratively applies a round function, swapping halves each round.
// This makes the transformation reversible while providing confusion and diffusion.
//
// Why it's used here:
// The anti-bot system needs to obscure relationships between raw event counts and timing data
// to prevent bots from easily reverse-engineering valid sensor patterns. This Feistel cipher
// provides a deterministic but non-obvious transformation that the server can verify.
//
// Parameters:
//   - upper32Bits: High-order 32 bits of input (typically event sum)
//   - lower32Bits: Low-order 32 bits of input (typically event count)
//   - key: Shared secret for round function (typically time since collection start)
//
// Returns: 64-bit obfuscated value combining transformed upper and lower halves
func feistelCipher(upper32Bits, lower32Bits, key int) int64 {
	// toInt32 safely converts int64 to int32, truncating to 32 bits
	toInt32 := func(n int64) int32 {
		return int32(n)
	}

	// roundFunction applies the F-function in the Feistel network.
	// It XORs the left half with a rotated version of the right half.
	// The rotation amount is determined by the round index, providing different
	// transformations in each round for better diffusion.
	//
	// Formula: left ^ (right >>> (32-round) | right <<< round)
	// This is a circular bit rotation of 'right' by 'round' positions, then XOR with 'left'.
	roundFunction := func(left, right, round int32) int32 {
		// Circular rotation: shift right by (32-round) and shift left by round, then OR them
		return left ^ (right>>(32-round) | toInt32(int64(right)<<round))
	}

	// Convert inputs to 32-bit signed integers
	rightHalf := toInt32(int64(upper32Bits))
	leftHalf := toInt32(int64(lower32Bits))

	// Combine into 64-bit value (left in lower 32 bits, right in upper 32 bits)
	combined := (int64(leftHalf) & 0xFFFFFFFF) | (int64(rightHalf) << 32)

	// Extract halves from combined value (this is redundant but matches original structure)
	leftHalf = toInt32(combined & 0xFFFFFFFF)
	rightHalf = toInt32((combined >> 32) & 0xFFFFFFFF)

	// Feistel network: 16 rounds of transformation
	// Each round:
	//   1. Apply round function: newLeft = right ^ F(left, key, round)
	//   2. Swap halves: (left, right) = (newLeft, left)
	for round := 0; round < 16; round++ {
		newLeft := rightHalf ^ roundFunction(leftHalf, int32(key), int32(round))
		// Swap: old left becomes new right, new left becomes new left
		oldLeft := leftHalf
		leftHalf = newLeft
		rightHalf = oldLeft
	}

	// Recombine the two halves into final 64-bit result
	// Right half in upper 32 bits, left half in lower 32 bits
	return (int64(rightHalf) << 32) | (int64(leftHalf) & 0xFFFFFFFF)
}

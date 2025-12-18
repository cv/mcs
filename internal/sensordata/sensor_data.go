package sensordata

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	mathrand "math/rand"
	"strconv"
	"time"
)

const (
	sdkVersion = "2.2.3"
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

	orientationEvent := b.generateOrientationDataAA()
	orientationEventCount := countSeparators(orientationEvent)
	motionEvent := b.generateMotionDataAA()
	motionEventCount := countSeparators(motionEvent)

	sensorData := ""
	sensorData += sdkVersion
	sensorData += "-1,2,-94,-100,"
	sensorData += b.systemInfo.ToString()
	sensorData += ","
	sensorData += strconv.Itoa(b.systemInfo.GetCharCodeSum())
	sensorData += ","
	sensorData += strconv.Itoa(int(randomNumber))
	sensorData += ","
	sensorData += strconv.FormatInt(timestampToMillis(b.sensorCollectionStartTimestamp)/2, 10)
	sensorData += "-1,2,-94,-101,"
	sensorData += "do_en"
	sensorData += ","
	sensorData += "dm_en"
	sensorData += ","
	sensorData += "t_en"
	sensorData += "-1,2,-94,-102,"
	sensorData += b.generateEditedText()
	sensorData += "-1,2,-94,-108,"
	sensorData += b.keyEventList.ToString()
	sensorData += "-1,2,-94,-117,"
	sensorData += b.touchEventList.ToString()
	sensorData += "-1,2,-94,-111,"
	sensorData += orientationEvent
	sensorData += "-1,2,-94,-109,"
	sensorData += motionEvent
	sensorData += "-1,2,-94,-144,"
	sensorData += b.generateOrientationDataAC()
	sensorData += "-1,2,-94,-142,"
	sensorData += b.generateOrientationDataAB()
	sensorData += "-1,2,-94,-145,"
	sensorData += b.generateMotionDataAC()
	sensorData += "-1,2,-94,-143,"
	sensorData += b.generateMotionEvent()
	sensorData += "-1,2,-94,-115,"
	sensorData += b.generateMiscStat(orientationEventCount, motionEventCount)
	sensorData += "-1,2,-94,-106,"
	sensorData += b.generateStoredValuesF()
	sensorData += ","
	sensorData += b.generateStoredValuesG()
	sensorData += "-1,2,-94,-120,"
	sensorData += b.generateStoredStackTraces()
	sensorData += "-1,2,-94,-112,"
	sensorData += b.performanceTestResults.ToString()
	sensorData += "-1,2,-94,-103,"
	sensorData += b.backgroundEventList.ToString()

	return encryptSensorData(sensorData)
}

func (b *SensorDataBuilder) generateEditedText() string {
	return ""
}

func (b *SensorDataBuilder) generateOrientationDataAA() string {
	return ""
}

func (b *SensorDataBuilder) generateMotionDataAA() string {
	return ""
}

func (b *SensorDataBuilder) generateOrientationDataAC() string {
	return ""
}

func (b *SensorDataBuilder) generateOrientationDataAB() string {
	return ""
}

func (b *SensorDataBuilder) generateMotionDataAC() string {
	return ""
}

func (b *SensorDataBuilder) generateMotionEvent() string {
	return ""
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

func (b *SensorDataBuilder) generateStoredStackTraces() string {
	return ""
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
	encryptedBytes, err := encryptAES128CBC([]byte(sensorData), aesKey, aesIV)
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

// encryptAES128CBC encrypts data using AES-128-CBC
func encryptAES128CBC(data, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Apply PKCS7 padding
	paddedData := pkcs7Pad(data, aes.BlockSize)

	ciphertext := make([]byte, len(paddedData))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, paddedData)

	return ciphertext, nil
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

func feistelCipher(upper32Bits, lower32Bits, key int) int64 {
	toSigned32 := func(n int64) int32 {
		return int32(n)
	}

	iterate := func(arg1, arg2, arg3 int32) int32 {
		return arg1 ^ (arg2>>(32-arg3) | toSigned32(int64(arg2)<<arg3))
	}

	upper := toSigned32(int64(upper32Bits))
	lower := toSigned32(int64(lower32Bits))

	data := (int64(lower) & 0xFFFFFFFF) | (int64(upper) << 32)

	lower2 := toSigned32(data & 0xFFFFFFFF)
	upper2 := toSigned32((data >> 32) & 0xFFFFFFFF)

	for i := 0; i < 16; i++ {
		v21 := upper2 ^ iterate(lower2, int32(key), int32(i))
		v8 := lower2
		lower2 = v21
		upper2 = v8
	}

	return (int64(upper2) << 32) | (int64(lower2) & 0xFFFFFFFF)
}

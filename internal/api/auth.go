package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cv/mcs/internal/cache"
	"github.com/cv/mcs/internal/sensordata"
)

const (
	// AuthRequestTimeout is the timeout for authentication-related API requests.
	AuthRequestTimeout = 15 * time.Second

	// IV is the initialization vector for AES encryption.
	IV = "0102030405060708"

	// SignatureMD5 is used for key derivation.
	SignatureMD5 = "C383D8C4D279B78130AD52DC71D95CAA"

	// AppPackageID identifies the mobile app package.
	AppPackageID = "com.interrait.mymazda"

	// UserAgentBaseAPI is the User-Agent for base API requests.
	UserAgentBaseAPI = "MyMazda-Android/9.0.5"

	// UserAgentUsherAPI is the User-Agent for Usher API requests.
	UserAgentUsherAPI = "MyMazda/9.0.5 (Google Pixel 3a; Android 11)"

	// AppOS identifies the operating system.
	AppOS = "Android"

	// AppVersion is the mobile app version.
	AppVersion = "9.0.5"

	// UsherSDKVersion is the Usher SDK version.
	UsherSDKVersion = "11.3.0700.001"

	// InternalUserID is a placeholder used in API requests.
	InternalUserID = "__INTERNAL_ID__"
)

// Authentication endpoint constants.
const (
	EndpointCheckVersion  = "service/checkVersion"
	EndpointEncryptionKey = "system/encryptionKey"
	EndpointLogin         = "user/login"
)

// Region represents a valid geographic region.
type Region string

const (
	// RegionMNAO represents Mazda North American Operations.
	RegionMNAO Region = "MNAO"
	// RegionMME represents Mazda Europe.
	RegionMME Region = "MME"
	// RegionMJO represents Mazda Japan.
	RegionMJO Region = "MJO"
)

// String returns the string representation of the region.
func (r Region) String() string {
	return string(r)
}

// IsValid checks if the region is valid.
func (r Region) IsValid() bool {
	_, ok := RegionConfigs[string(r)]

	return ok
}

// ParseRegion parses a string into a Region, returning an error if invalid.
func ParseRegion(s string) (Region, error) {
	r := Region(s)
	if !r.IsValid() {
		return "", fmt.Errorf("invalid region: %s (must be one of: MNAO, MME, MJO)", s)
	}

	return r, nil
}

// RegionConfig holds configuration for a specific region.
type RegionConfig struct {
	AppCode  string
	BaseURL  string
	UsherURL string
}

// RegionConfigs maps region codes to their configurations.
var RegionConfigs = map[string]RegionConfig{
	"MNAO": {
		AppCode:  "202007270941270111799",
		BaseURL:  "https://0cxo7m58.mazda.com/prod/",
		UsherURL: "https://ptznwbh8.mazda.com/appapi/v1/",
	},
	"MME": {
		AppCode:  "202008100250281064816",
		BaseURL:  "https://e9stj7g7.mazda.com/prod/",
		UsherURL: "https://rz97suam.mazda.com/appapi/v1/",
	},
	"MJO": {
		AppCode:  "202009170613074283422",
		BaseURL:  "https://wcs9p6wj.mazda.com/prod/",
		UsherURL: "https://c5ulfwxr.mazda.com/appapi/v1/",
	},
}

// Client represents an API client.
type Client struct {
	email    string
	password string
	region   Region

	baseURL  string
	usherURL string
	appCode  string

	baseAPIDeviceID  string
	usherAPIDeviceID string

	Keys                    Keys
	accessToken             string
	accessTokenExpirationTs int64

	httpClient        *http.Client
	debug             bool
	sensorDataBuilder *sensordata.SensorDataBuilder
	sleepFunc         func(context.Context, time.Duration) error
}

// NewClient creates a new API client.
func NewClient(email, password string, region Region) (*Client, error) {
	if !region.IsValid() {
		return nil, fmt.Errorf("invalid region: %s", region)
	}

	config := RegionConfigs[string(region)]

	return &Client{
		email:             email,
		password:          password,
		region:            region,
		baseURL:           config.BaseURL,
		usherURL:          config.UsherURL,
		appCode:           config.AppCode,
		baseAPIDeviceID:   GenerateUUIDFromSeed(email),
		usherAPIDeviceID:  GenerateUsherDeviceID(email),
		httpClient:        &http.Client{Timeout: 30 * time.Second},
		debug:             false,
		sensorDataBuilder: sensordata.NewSensorDataBuilder(),
		sleepFunc:         sleepWithContext,
	}, nil
}

// SetDebug enables or disables debug logging.
func (c *Client) SetDebug(debug bool) {
	c.debug = debug
}

// SetCachedCredentials sets the client's cached authentication credentials.
func (c *Client) SetCachedCredentials(accessToken string, accessTokenExpirationTs int64, encKey, signKey string) {
	c.accessToken = accessToken
	c.accessTokenExpirationTs = accessTokenExpirationTs
	c.Keys.EncKey = encKey
	c.Keys.SignKey = signKey
}

// GetCredentials returns the current authentication credentials for caching.
func (c *Client) GetCredentials() (accessToken string, accessTokenExpirationTs int64, encKey, signKey string) {
	return c.accessToken, c.accessTokenExpirationTs, c.Keys.EncKey, c.Keys.SignKey
}

// GetEncryptionKeys retrieves the encryption and signing keys from the API.
func (c *Client) GetEncryptionKeys(ctx context.Context) error {
	// Ensure we have a timeout for the request
	ctx, cancel := context.WithTimeout(ctx, AuthRequestTimeout)
	defer cancel()
	timestamp := getTimestampStrMs()

	headers := map[string]string{
		"device-id":     c.baseAPIDeviceID,
		"app-code":      c.appCode,
		"app-os":        AppOS,
		"user-agent":    UserAgentBaseAPI,
		"app-version":   AppVersion,
		"app-unique-id": AppPackageID,
		"access-token":  "",
		"req-id":        "req_" + timestamp,
		"timestamp":     timestamp,
		"sign":          c.getSignFromTimestamp(timestamp),
		"Content-Type":  "application/json",
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+EndpointCheckVersion, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var response APIBaseResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if response.State != "S" {
		return fmt.Errorf("request failed with state: %v", response.State)
	}

	if response.Payload == "" {
		return errors.New("payload not found in response")
	}

	decrypted, err := c.decryptCheckVersionPayload(response.Payload)
	if err != nil {
		return fmt.Errorf("failed to decrypt payload: %w", err)
	}

	c.Keys.EncKey = decrypted.EncKey
	c.Keys.SignKey = decrypted.SignKey

	return nil
}

// GetUsherEncryptionKey retrieves the RSA public key from Usher API.
func (c *Client) GetUsherEncryptionKey(ctx context.Context) (string, string, error) {
	// Ensure we have a timeout for the request
	ctx, cancel := context.WithTimeout(ctx, AuthRequestTimeout)
	defer cancel()
	params := url.Values{
		"appId":      []string{"MazdaApp"},
		"locale":     []string{"en-US"},
		"deviceId":   []string{c.usherAPIDeviceID},
		"sdkVersion": []string{UsherSDKVersion},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.usherURL+EndpointEncryptionKey+"?"+params.Encode(), nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", UserAgentUsherAPI)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	var response UsherEncryptionKeyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", "", fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Data.PublicKey == "" {
		return "", "", errors.New("public key not found in response")
	}

	return response.Data.PublicKey, response.Data.VersionPrefix, nil
}

// Login authenticates with the API and retrieves an access token.
func (c *Client) Login(ctx context.Context) error {
	// Ensure we have a timeout for the request
	ctx, cancel := context.WithTimeout(ctx, AuthRequestTimeout)
	defer cancel()
	// Get RSA public key for password encryption
	publicKey, versionPrefix, err := c.GetUsherEncryptionKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to get encryption key: %w", err)
	}

	// Encrypt password
	encryptedPassword, err := c.encryptPasswordWithPublicKey(c.password, publicKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Prepare login request
	loginData := map[string]any{
		"appId":      "MazdaApp",
		"deviceId":   c.usherAPIDeviceID,
		"locale":     "en-US",
		"password":   versionPrefix + encryptedPassword,
		"sdkVersion": UsherSDKVersion,
		"userId":     c.email,
		"userIdType": "email",
	}

	jsonData, err := json.Marshal(loginData)
	if err != nil {
		return fmt.Errorf("failed to marshal login data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.usherURL+EndpointLogin, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", UserAgentUsherAPI)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var response LoginResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if response.Status == "INVALID_CREDENTIAL" {
		return errors.New("invalid email or password")
	}
	if response.Status == "USER_LOCKED" {
		return errors.New("account is locked")
	}
	if response.Status != "OK" {
		return fmt.Errorf("login failed with status: %s", response.Status)
	}

	if response.Data.AccessToken == "" {
		return errors.New("access token not found in response")
	}

	c.accessToken = response.Data.AccessToken
	c.accessTokenExpirationTs = response.Data.AccessTokenExpirationTs

	return nil
}

// IsTokenValid checks if the access token is present and not expired.
func (c *Client) IsTokenValid() bool {
	return cache.IsTokenValid(c.accessToken, c.accessTokenExpirationTs)
}

// Helper functions

func getTimestampStrMs() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}

func (c *Client) getSignFromTimestamp(timestamp string) string {
	if timestamp == "" {
		return ""
	}

	timestampExtended := strings.ToUpper(timestamp + timestamp[6:] + timestamp[3:])
	temporarySignKey := c.getTemporarySignKeyFromAppCode()

	return SignWithSHA256(timestampExtended + temporarySignKey)
}

func (c *Client) getTemporarySignKeyFromAppCode() string {
	val1 := SignWithMD5(c.appCode + AppPackageID)
	val2 := strings.ToLower(SignWithMD5(val1 + SignatureMD5))

	return val2[20:32] + val2[0:10] + val2[4:6]
}

func (c *Client) getDecryptionKeyFromAppCode() string {
	val1 := SignWithMD5(c.appCode + AppPackageID)
	val2 := strings.ToLower(SignWithMD5(val1 + SignatureMD5))

	return val2[4:20]
}

// decryptCheckVersionPayload decrypts and parses the checkVersion response payload.
func (c *Client) decryptCheckVersionPayload(payload string) (*CheckVersionResponse, error) {
	key := c.getDecryptionKeyFromAppCode()
	decrypted, err := DecryptAES128CBC(payload, key, IV)
	if err != nil {
		return nil, err
	}

	var result CheckVersionResponse
	if err := json.Unmarshal(decrypted, &result); err != nil {
		return nil, fmt.Errorf("failed to parse decrypted payload: %w", err)
	}

	return &result, nil
}

func (c *Client) encryptPasswordWithPublicKey(password, publicKey string) (string, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	data := password + ":" + timestamp

	encrypted, err := EncryptRSA(data, publicKey)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

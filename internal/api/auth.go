package api

import (
	"context"
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

	// SignatureSHA256 is the app's signing-certificate SHA-256, used for key
	// derivation on Mazda's newer "MC" host.
	SignatureSHA256 = "C022C9EE778CF903838F8B9C4B9FF0036A5C516CEFAAD6DC710B717CF97DCFCA"

	// AppPackageID identifies the mobile app package.
	AppPackageID = "com.interrait.mymazda"

	// UserAgentBaseAPI is the User-Agent for base API requests.
	UserAgentBaseAPI = "MyMazda-Android/9.5.0"

	// UserAgentUsherAPI is the User-Agent for Usher API requests.
	UserAgentUsherAPI = "MyMazda/9.0.5 (Google Pixel 3a; Android 11)"

	// AppOS identifies the operating system.
	AppOS = "Android"

	// AppVersion is the mobile app version.
	AppVersion = "9.5.0"

	// UsherSDKVersion is the Usher SDK version.
	UsherSDKVersion = "11.3.0700.001"

	// InternalUserID is a placeholder used in API requests.
	InternalUserID      = "__INTERNAL_ID__"
	internalUserIDField = "internaluserid"
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
	UsherURL string // the old sign-in (gone in 2026; kept for reference)
	B2C      B2CConfig
}

// RegionConfigs maps region codes to their configurations.
var RegionConfigs = map[string]RegionConfig{
	"MNAO": {
		AppCode:  "498345786246797888995",
		BaseURL:  "https://hgs2ivna.mazda.com/",
		UsherURL: "https://ptznwbh8.mazda.com/appapi/v1/",
		B2C: B2CConfig{Host: "na.id.mazda.com", Tenant: "pduspb2c01", ClientID: "2daf581c-65c1-4fdb-b46a-efa98c6ba5b7",
			Scope: "https://pduspb2c01.onmicrosoft.com/0728deea-be48-4382-9ef1-d4ff6d679ffa/cv", Country: "US"},
	},
	"MME": {
		AppCode:  "365747628595648782737",
		BaseURL:  "https://hgs2iveu.mazda.com/",
		UsherURL: "https://rz97suam.mazda.com/appapi/v1/",
		B2C: B2CConfig{Host: "eu.id.mazda.com", Tenant: "pdeupb2c01", ClientID: "cbfe43e1-6949-42fe-996e-1a56f41a891d",
			Scope: "https://pdeupb2c01.onmicrosoft.com/dcd35c5a-b32f-4add-ac6c-ba6e8bbfa11b/cv", Country: "GB"},
	},
	"MJO": {
		AppCode:  "438849393836584965983",
		BaseURL:  "https://hgs2ivap.mazda.com/",
		UsherURL: "https://c5ulfwxr.mazda.com/appapi/v1/",
		B2C: B2CConfig{Host: "ap.id.mazda.com", Tenant: "pdjppb2c01", ClientID: "455191ae-e98e-4748-8bc6-57a5a570f1c2",
			Scope: "https://pdjppb2c01.onmicrosoft.com/1c6d5f69-fea6-4019-93df-899ae820b0f4/cv", Country: "JP"},
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

	b2c     B2CConfig
	country string // the account's country for sign-in (e.g. US, CA); empty: the region's default

	Keys                    Keys
	accessToken             string
	accessTokenExpirationTs int64

	httpClient        *http.Client
	debug             bool
	sensorDataBuilder *sensordata.SensorDataBuilder
	sleepFunc         func(context.Context, time.Duration) error
	refreshToken      string
	deviceSessionID   string
	refreshOnly       bool
	signature         string // certificate hash for key derivation (empty: MD5)
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
		signature:         SignatureSHA256,
		baseAPIDeviceID:   GenerateUUIDFromSeed(email),
		usherAPIDeviceID:  GenerateUsherDeviceID(email),
		b2c:               config.B2C,
		httpClient:        &http.Client{Timeout: 30 * time.Second},
		debug:             false,
		sensorDataBuilder: sensordata.NewSensorDataBuilder(),
		sleepFunc:         sleepWithContext,
	}, nil
}

// SetCountry sets the account's country for sign-in (two letters, e.g. "CA"); Mazda
// refuses a sign-in for a country the account isn't registered in.
func (c *Client) SetCountry(country string) {
	c.country = strings.ToUpper(strings.TrimSpace(country))
}

// SetRefreshOnly prevents unattended clients from replacing another device's session.
func (c *Client) SetRefreshOnly(enabled bool) { c.refreshOnly = enabled }

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
		return fmt.Errorf("encryption key request refused (code %v)", response.ErrorCode)
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
	ctx, cancel := context.WithTimeout(ctx, 3*AuthRequestTimeout)
	defer cancel()

	var tok *b2cToken
	var err error
	if c.refreshToken != "" {
		tok, err = c.b2cRefresh(ctx)
		// Never silently replace the phone's session after a failed refresh.
		if err != nil {
			return fmt.Errorf("refresh failed; explicit sign-in required: %w", err)
		}
	} else {
		if c.refreshOnly {
			return errors.New("explicit sign-in required; password login disabled")
		}
		tok, err = c.b2cLogin(ctx)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}
	}
	c.accessToken = tok.AccessToken
	c.accessTokenExpirationTs = time.Now().Unix() + tok.ExpiresIn - 60
	if tok.RefreshToken != "" {
		c.refreshToken = tok.RefreshToken
	}
	if c.deviceSessionID == "" {
		return c.attachSession(ctx)
	}

	return nil
}

// IsTokenValid checks if the access token is present and not expired.
func (c *Client) IsTokenValid() bool {
	return cache.IsTokenValid(c.accessToken, c.accessTokenExpirationTs)
}

// Helper functions

// certSignature is the app certificate hash used in key derivation: MD5 on the
// legacy host, SHA-256 on the MC host.
func (c *Client) certSignature() string {
	if c.signature != "" {
		return c.signature
	}

	return SignatureMD5
}

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
	val2 := strings.ToLower(SignWithMD5(val1 + c.certSignature()))

	return val2[20:32] + val2[0:10] + val2[4:6]
}

func (c *Client) getDecryptionKeyFromAppCode() string {
	val1 := SignWithMD5(c.appCode + AppPackageID)
	val2 := strings.ToLower(SignWithMD5(val1 + c.certSignature()))

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

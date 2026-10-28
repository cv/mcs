package api

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/cv/mcs/internal/cache"
)

const sessionProtocol = "mc-b2c-v1"

// attachSession establishes the CV session following B2C sign-in. This may sign
// the phone out: never call it in a retry loop after another device took over.
func (c *Client) attachSession(ctx context.Context) error {
	if err := c.ensureKeysPresent(ctx); err != nil {
		return err
	}
	token, _, err := pkcePair()
	if err != nil {
		return err
	}
	body := map[string]any{"serviceType": 0, "deviceInfo": map[string]any{"deviceToken": token, "deviceType": 1}, "locale": "en-" + c.accountCountry(), "country": c.accountCountry()}
	// Direct send avoids recursive ensureTokenValid -> attachSession.
	b, err := c.sendAPIRequestJSON(ctx, "POST", "remoteServices/attach/v4", nil, body, true, true)
	if err != nil {
		return err
	}
	var att struct {
		ResultCode string `json:"resultCode"`
		Data       struct {
			ServiceRegistStatus int `json:"serviceRegistStatus"`
			UserInfo            struct {
				SessionID string `json:"sessionId"`
			} `json:"userinfo"`
		} `json:"data"`
	}
	if json.Unmarshal(b, &att) != nil || att.ResultCode != "B2C200000" || att.Data.ServiceRegistStatus != 1 || att.Data.UserInfo.SessionID == "" {
		return errors.New("mazda session attachment refused; check MyMazda registration")
	}
	c.deviceSessionID = att.Data.UserInfo.SessionID

	return nil
}

// Scope the cache to this account, country and protocol. Legacy keys/tokens must
// not be reused with the MC service. Expired tokens are still useful for refresh.
func (c *Client) cacheAccount() string {
	return GenerateUUIDFromSeed(c.email + "|" + c.region.String() + "|" + c.accountCountry())
}
func (c *Client) RestoreSession(s *cache.TokenCache) {
	if s == nil || s.Protocol != sessionProtocol || s.Account != c.cacheAccount() {
		return
	}
	c.SetCachedCredentials(s.AccessToken, s.AccessTokenExpirationTs, s.EncKey, s.SignKey)
	c.refreshToken = s.RefreshToken
	c.deviceSessionID = s.DeviceSessionID
}
func (c *Client) CachedSession() *cache.TokenCache {
	return &cache.TokenCache{Protocol: sessionProtocol, Account: c.cacheAccount(), AccessToken: c.accessToken, AccessTokenExpirationTs: c.accessTokenExpirationTs, EncKey: c.Keys.EncKey, SignKey: c.Keys.SignKey, RefreshToken: c.refreshToken, DeviceSessionID: c.deviceSessionID}
}

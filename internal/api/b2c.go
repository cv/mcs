package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
)

// B2CConfig comes from MyMazda 9.5's regional MSAL configuration.
type B2CConfig struct {
	Host, Tenant, ClientID, Scope, Country string
}

const (
	b2cPolicy        = "B2C_1A_SIGNIN"
	b2cClientIDField = "client_id"
	b2cScopeField    = "scope"
	// This is the registered URI, including the literal percent escapes.
	// #nosec G101 -- public registered OAuth redirect, not a credential.
	b2cRedirect   = "msauth://com.interrait.mymazda/%2FnKMu1%2BlCjy5%2Be7OF9vfp4eFBks%3D"
	b2cAgent      = "Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0 Mobile Safari/537.36"
	authBodyLimit = 2 << 20
)

var b2cSettingsRe = regexp.MustCompile(`(?s)var SETTINGS = (\{.*?\});`)

type b2cSettings struct {
	TransID string                          `json:"transId"`
	CSRF    string                          `json:"csrf"`
	API     string                          `json:"api"`
	Hosts   struct{ Tenant, Policy string } `json:"hosts"`
}

type b2cToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func (c *Client) b2cBase() string {
	return "https://" + c.b2c.Host + "/" + c.b2c.Tenant + ".onmicrosoft.com/" + b2cPolicy
}
func (c *Client) accountCountry() string {
	if c.country != "" {
		return c.country
	}

	return c.b2c.Country
}
func (c *Client) authHTTPClient() *http.Client {
	jar, _ := cookiejar.New(nil)

	return &http.Client{Jar: jar, Transport: c.httpClient.Transport, Timeout: AuthRequestTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Scheme == "msauth" {
				return http.ErrUseLastResponse
			}
			if len(via) >= 10 || req.URL.Scheme != "https" || req.URL.Host != c.b2c.Host {
				return errors.New("unexpected sign-in redirect")
			}

			return nil
		},
	}
}

// authRequest never includes server bodies or URLs in errors: they can contain
// credentials, OAuth codes, CSRF values or tokens. Redirects keep passwords on-host.
func authRequest(ctx context.Context, hc *http.Client, method, u string, form url.Values, headers map[string]string) ([]byte, string, error) {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, "", errors.New("invalid sign-in request")
	}
	req.Header.Set("User-Agent", b2cAgent)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := hc.Do(req)
	if err != nil {
		return nil, "", errors.New("sign-in transport failed")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, "", fmt.Errorf("sign-in HTTP %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, authBodyLimit+1))
	if err != nil || len(b) > authBodyLimit {
		return nil, "", errors.New("invalid sign-in response")
	}

	return b, resp.Header.Get("Location"), nil
}

func (c *Client) b2cLogin(ctx context.Context) (*b2cToken, error) {
	if c.b2c.Host == "" {
		return nil, errors.New("sign-in is not configured for this region")
	}
	verifier, challenge, err := pkcePair()
	if err != nil {
		return nil, err
	}
	state, _, err := pkcePair()
	if err != nil {
		return nil, err
	}
	hc := c.authHTTPClient()
	scope := c.b2c.Scope + " openid offline_access profile"
	q := url.Values{b2cClientIDField: {c.b2c.ClientID}, "response_type": {"code"}, "redirect_uri": {b2cRedirect}, b2cScopeField: {scope}, "code_challenge": {challenge}, "code_challenge_method": {"S256"}, "state": {state}, "prompt": {"login"}, "ui_locales": {"en-" + c.accountCountry()}, "country": {c.accountCountry()}}
	st, err := c.b2cSettings(ctx, hc, q)
	if err != nil {
		return nil, err
	}
	if err = c.b2cPassword(ctx, hc, st); err != nil {
		return nil, err
	}
	code, err := c.b2cConfirm(ctx, hc, st, state)
	if err != nil {
		return nil, err
	}

	return c.b2cTokenRequest(ctx, hc, url.Values{"grant_type": {"authorization_code"}, b2cClientIDField: {c.b2c.ClientID}, "code": {code}, "redirect_uri": {b2cRedirect}, "code_verifier": {verifier}, b2cScopeField: {scope}})
}

func (c *Client) b2cSettings(ctx context.Context, hc *http.Client, q url.Values) (*b2cSettings, error) {
	page, _, err := authRequest(ctx, hc, http.MethodGet, c.b2cBase()+"/oauth2/v2.0/authorize?"+q.Encode(), nil, nil)
	if err != nil {
		return nil, err
	}
	m := b2cSettingsRe.FindSubmatch(page)
	if m == nil {
		return nil, errors.New("sign-in page: no settings found")
	}
	var st b2cSettings
	if json.Unmarshal(m[1], &st) != nil || st.CSRF == "" || st.TransID == "" || !strings.EqualFold(st.Hosts.Policy, b2cPolicy) || !strings.EqualFold(strings.Trim(st.Hosts.Tenant, "/"), c.b2c.Tenant+".onmicrosoft.com/"+b2cPolicy) {
		return nil, errors.New("invalid sign-in page settings")
	}

	return &st, nil
}
func (c *Client) b2cPassword(ctx context.Context, hc *http.Client, st *b2cSettings) error {
	tenant := "https://" + c.b2c.Host + "/" + strings.Trim(st.Hosts.Tenant, "/")
	body, _, err := authRequest(ctx, hc, http.MethodPost, tenant+"/SelfAsserted?"+url.Values{"tx": {st.TransID}, "p": {st.Hosts.Policy}}.Encode(), url.Values{"request_type": {"RESPONSE"}, "signInName": {c.email}, "password": {c.password}}, map[string]string{"X-CSRF-TOKEN": st.CSRF, "X-Requested-With": "XMLHttpRequest"})
	if err != nil {
		return err
	}
	var sa struct {
		Status string `json:"status"`
	}
	if json.Unmarshal(body, &sa) != nil || sa.Status != "200" {
		return errors.New("sign-in refused; check credentials and account requirements in MyMazda")
	}

	return nil
}
func (c *Client) b2cConfirm(ctx context.Context, hc *http.Client, st *b2cSettings, state string) (string, error) {
	tenant := "https://" + c.b2c.Host + "/" + strings.Trim(st.Hosts.Tenant, "/")
	if st.API != "CombinedSigninAndSignup" {
		return "", errors.New("unsupported sign-in confirmation flow")
	}
	_, loc, err := authRequest(ctx, hc, http.MethodGet, tenant+"/api/"+st.API+"/confirmed?"+url.Values{"rememberMe": {"false"}, "csrf_token": {st.CSRF}, "tx": {st.TransID}, "p": {st.Hosts.Policy}}.Encode(), nil, nil)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(loc, b2cRedirect+"?") {
		return "", errors.New("sign-in did not finish at the registered redirect")
	}
	ru, err := url.Parse(loc)
	if err != nil {
		return "", errors.New("invalid sign-in redirect")
	}
	if ru.Query().Get("state") != state {
		return "", errors.New("sign-in state mismatch")
	}
	code := ru.Query().Get("code")
	if code == "" {
		return "", errors.New("sign-in: no authorization code")
	}

	return code, nil
}

func (c *Client) b2cRefresh(ctx context.Context) (*b2cToken, error) {
	return c.b2cTokenRequest(ctx, c.authHTTPClient(), url.Values{"grant_type": {"refresh_token"}, b2cClientIDField: {c.b2c.ClientID}, "refresh_token": {c.refreshToken}, b2cScopeField: {c.b2c.Scope + " openid offline_access profile"}})
}
func (c *Client) b2cTokenRequest(ctx context.Context, hc *http.Client, form url.Values) (*b2cToken, error) {
	body, _, err := authRequest(ctx, hc, http.MethodPost, c.b2cBase()+"/oauth2/v2.0/token", form, nil)
	if err != nil {
		return nil, err
	}
	var tok b2cToken
	if json.Unmarshal(body, &tok) != nil || tok.AccessToken == "" || tok.ExpiresIn <= 60 {
		return nil, errors.New("sign-in token refused")
	}

	return &tok, nil
}
func pkcePair() (verifier, challenge string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(verifier))

	return verifier, base64.RawURLEncoding.EncodeToString(sum[:]), nil
}

package api

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:cyclop // End-to-end protocol mock and explicit failure scenarios.
func TestB2CSession(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"success", "state mismatch", "no session", "refresh denied"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			c, err := NewClient("test@example.com", "do-not-log-this", RegionMNAO)
			require.NoError(t, err)
			c.SetCountry("CA")
			var state, challenge string
			var logins, attaches, refreshes int
			reply := func(w http.ResponseWriter, payload any, key string) {
				b, e := json.Marshal(payload)
				require.NoError(t, e)
				enc, e := EncryptAES128CBC(b, key, IV)
				require.NoError(t, e)
				_ = json.NewEncoder(w).Encode(map[string]any{"state": "S", "payload": enc})
			}
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.HasSuffix(r.URL.Path, "/authorize"):
					logins++
					q := r.URL.Query()
					state = q.Get("state")
					challenge = q.Get("code_challenge")
					assert.NotEmpty(t, state)
					assert.Equal(t, "CA", q.Get("country"))
					assert.Equal(t, "en-CA", q.Get("ui_locales"))
					assert.Equal(t, b2cRedirect, q.Get("redirect_uri"))
					assert.Equal(t, "S256", q.Get("code_challenge_method"))
					http.SetCookie(w, &http.Cookie{Name: "flow", Value: "cookie", Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode})
					_, _ = fmt.Fprintf(w, `var SETTINGS = {"transId":"transaction","csrf":"csrf","api":"CombinedSigninAndSignup","hosts":{"tenant":"/%s.onmicrosoft.com/B2C_1A_signin","policy":"B2C_1A_signin"}};`, c.b2c.Tenant)
				case strings.HasSuffix(r.URL.Path, "/SelfAsserted"):
					assert.NoError(t, r.ParseForm())
					assert.Equal(t, "do-not-log-this", r.Form.Get("password"))
					assert.Equal(t, c.email, r.Form.Get("signInName"))
					assert.Equal(t, "csrf", r.Header.Get("X-Csrf-Token"))
					cookie, e := r.Cookie("flow")
					if !assert.NoError(t, e) {
						return
					}
					assert.Equal(t, "cookie", cookie.Value)
					_, _ = fmt.Fprint(w, `{"status":"200"}`)
				case strings.HasSuffix(r.URL.Path, "/confirmed"):
					st := state
					if mode == "state mismatch" {
						st = "wrong"
					}
					w.Header().Set("Location", b2cRedirect+"?"+url.Values{"code": {"code"}, "state": {st}}.Encode())
					w.WriteHeader(http.StatusFound)
				case strings.HasSuffix(r.URL.Path, "/token"):
					assert.NoError(t, r.ParseForm())
					if r.Form.Get("grant_type") == "authorization_code" {
						assert.Equal(t, "code", r.Form.Get("code"))
						assert.Equal(t, b2cRedirect, r.Form.Get("redirect_uri"))
						hash := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
						assert.Equal(t, challenge, base64.RawURLEncoding.EncodeToString(hash[:]))
					} else {
						refreshes++
						assert.Equal(t, "refresh_token", r.Form.Get("grant_type"))
						assert.Equal(t, "refresh", r.Form.Get("refresh_token"))
						if mode == "refresh denied" {
							w.WriteHeader(http.StatusBadRequest)
							_, _ = fmt.Fprint(w, "do-not-log-this")

							return
						}
					}
					_, _ = fmt.Fprint(w, `{"access_token":"token","expires_in":7200,"refresh_token":"refresh"}`)
				case r.URL.Path == "/service/checkVersion":
					assert.Equal(t, c.getSignFromTimestamp(r.Header.Get("Timestamp")), r.Header.Get("Sign"))
					reply(w, map[string]string{"encKey": "1234567890abcdef", "signKey": "sign"}, c.getDecryptionKeyFromAppCode())
				case r.URL.Path == "/remoteServices/attach/v4":
					attaches++
					assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
					assert.Equal(t, "token", r.Header.Get("Access-Token"))
					b, _ := io.ReadAll(r.Body)
					plain, e := DecryptAES128CBC(string(b), c.Keys.EncKey, IV)
					if !assert.NoError(t, e) {
						return
					}
					var req map[string]any
					assert.NoError(t, json.Unmarshal(plain, &req))
					assert.InDelta(t, 0, req["serviceType"], 0)
					assert.Equal(t, "CA", req["country"])
					session := "session"
					if mode == "no session" {
						session = ""
					}
					reply(w, map[string]any{"resultCode": "B2C200000", "data": map[string]any{"serviceRegistStatus": 1, "userinfo": map[string]string{"sessionId": session, "mazdaId": "unused"}}}, c.Keys.EncKey)
				default:
					t.Errorf("unexpected endpoint: %s", r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			defer server.Close()
			c.httpClient = server.Client()
			c.b2c.Host = strings.TrimPrefix(server.URL, "https://")
			c.baseURL = server.URL + "/"
			err = c.Login(t.Context())
			if mode == "state mismatch" || mode == "no session" {
				require.Error(t, err)
				assert.NotContains(t, err.Error(), "do-not-log-this")

				return
			}
			require.NoError(t, err)
			assert.Equal(t, "session", c.deviceSessionID)
			assert.True(t, c.IsTokenValid())
			saved := c.CachedSession()
			saved.AccessTokenExpirationTs = time.Now().Unix() - 1
			c2, e := NewClient(c.email, c.password, RegionMNAO)
			require.NoError(t, e)
			c2.SetCountry("CA")
			c2.httpClient = server.Client()
			c2.b2c.Host = c.b2c.Host
			c2.baseURL = c.baseURL
			c2.RestoreSession(saved)
			err = c2.ensureTokenValid(t.Context())
			if mode == "refresh denied" {
				require.Error(t, err)
				assert.NotContains(t, err.Error(), "do-not-log-this")
			} else {
				require.NoError(t, err)
				assert.True(t, c2.IsTokenValid())
			}
			assert.Equal(t, 1, logins)
			assert.Equal(t, 1, attaches)
			assert.Equal(t, 1, refreshes)
			assert.Equal(t, "session", c2.deviceSessionID)
			// Neither another account/country nor the old protocol imports this session.
			c3, e := NewClient("other@example.com", c.password, RegionMNAO)
			require.NoError(t, e)
			c3.SetCountry("CA")
			c3.RestoreSession(saved)
			assert.Empty(t, c3.accessToken)
			c3.email = c.email
			c3.SetCountry("US")
			c3.RestoreSession(saved)
			assert.Empty(t, c3.accessToken)
			c3.SetCountry("CA")
			saved.Protocol = ""
			c3.RestoreSession(saved)
			assert.Empty(t, c3.accessToken)
		})
	}
}

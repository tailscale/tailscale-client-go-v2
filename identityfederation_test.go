// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateIDToken(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		futureExp := time.Now().Add(1 * time.Hour).Unix()

		err := validateIDToken(createIDToken(futureExp))

		require.NoError(t, err)
	})

	t.Run("expired token", func(t *testing.T) {
		pastExp := time.Now().Add(-1 * time.Hour).Unix()

		err := validateIDToken(createIDToken(pastExp))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("missing exp claim", func(t *testing.T) {
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
		payload := base64.RawURLEncoding.EncodeToString([]byte(`{}`))
		signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))
		token := fmt.Sprintf("%s.%s.%s", header, payload, signature)

		err := validateIDToken(token)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing 'exp'")
	})

	t.Run("invalid JWT format - too few parts", func(t *testing.T) {
		err := validateIDToken("invalid.token")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JWT format")
	})

	t.Run("invalid JWT format - too many parts", func(t *testing.T) {
		err := validateIDToken("part1.part2.part3.part4")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JWT format")
	})

	t.Run("invalid base64 in payload", func(t *testing.T) {
		err := validateIDToken("header.invalid-base64!@#.signature")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode JWT payload")
	})

	t.Run("invalid JSON in payload", func(t *testing.T) {
		header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256"}`))
		payload := base64.RawURLEncoding.EncodeToString([]byte(`{invalid json`))
		signature := base64.RawURLEncoding.EncodeToString([]byte("sig"))
		token := fmt.Sprintf("%s.%s.%s", header, payload, signature)

		err := validateIDToken(token)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse JWT claims")
	})
}

func TestHTTPClient(t *testing.T) {
	validToken := createIDToken(time.Now().Add(1 * time.Hour).Unix())

	t.Run("success with static ID token", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v2/oauth/token-exchange" {
				json.NewEncoder(w).Encode(TokenExchangeResponse{
					AccessToken: "ts-api-test-token",
					TokenType:   "Bearer",
					ExpiresIn:   3600,
				})
			}
		}))
		defer srv.Close()

		client, err := IdentityFederationConfig{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				return validToken, nil
			},
			BaseURL: srv.URL,
		}.HTTPClient()

		require.NoError(t, err)
		require.NotNil(t, client)
	})

	t.Run("success with token generator", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v2/oauth/token-exchange" {
				json.NewEncoder(w).Encode(TokenExchangeResponse{
					AccessToken: "ts-api-test-token",
					TokenType:   "Bearer",
					ExpiresIn:   3600,
				})
			}
		}))
		defer srv.Close()

		generatorCalled := false
		client, err := IdentityFederationConfig{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				generatorCalled = true
				return validToken, nil
			},
			BaseURL: srv.URL,
		}.HTTPClient()

		require.NoError(t, err)
		require.NotNil(t, client)
		assert.True(t, generatorCalled, "generator should be called during initialization")
	})

	t.Run("missing client ID", func(t *testing.T) {
		_, err := IdentityFederationConfig{
			IDTokenFunc: func() (string, error) {
				return validToken, nil
			},
		}.HTTPClient()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "ClientID is required")
	})

	t.Run("missing IDTokenFunc", func(t *testing.T) {
		_, err := IdentityFederationConfig{
			ClientID: "test-client-id",
		}.HTTPClient()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "IDTokenFunc is required")
	})

	t.Run("expired ID token", func(t *testing.T) {
		expiredToken := createIDToken(time.Now().Add(-1 * time.Hour).Unix())

		_, err := IdentityFederationConfig{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				return expiredToken, nil
			},
		}.HTTPClient()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("generator returns error", func(t *testing.T) {
		_, err := IdentityFederationConfig{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				return "", fmt.Errorf("generator error")
			},
		}.HTTPClient()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch ID token")
		assert.Contains(t, err.Error(), "generator error")
	})

	t.Run("invalid JWT format", func(t *testing.T) {
		_, err := IdentityFederationConfig{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				return "not.a.valid.jwt", nil
			},
		}.HTTPClient()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JWT format")
	})
}

func TestTokenTransportRoundTrip(t *testing.T) {
	validToken := createIDToken(time.Now().Add(1 * time.Hour).Unix())

	t.Run("adds token to request using bearer token", func(t *testing.T) {
		var capturedAuthHeader string
		apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedAuthHeader = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
		}))
		defer apiSrv.Close()

		tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v2/oauth/token-exchange" {
				json.NewEncoder(w).Encode(TokenExchangeResponse{
					AccessToken: "test-access-token",
					TokenType:   "Bearer",
					ExpiresIn:   3600,
				})
			}
		}))
		defer tokenSrv.Close()

		httpClient, err := IdentityFederationConfig{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				return validToken, nil
			},
			BaseURL: tokenSrv.URL,
		}.HTTPClient()
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", apiSrv.URL, nil)
		_, err = httpClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, "Bearer test-access-token", capturedAuthHeader)
	})

	t.Run("generator called on expired token", func(t *testing.T) {
		freshToken := createIDToken(time.Now().Add(1 * time.Hour).Unix())

		generatorCallCount := 0
		tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v2/oauth/token-exchange" {
				json.NewEncoder(w).Encode(TokenExchangeResponse{
					AccessToken: "test-access-token",
					TokenType:   "Bearer",
					ExpiresIn:   3600,
				})
			}
		}))
		defer tokenSrv.Close()

		httpClient, err := IdentityFederationConfig{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				generatorCallCount++
				return freshToken, nil
			},
			BaseURL: tokenSrv.URL,
		}.HTTPClient()
		require.NoError(t, err)

		// Initial call should use generator
		assert.Equal(t, 1, generatorCallCount)

		// Manually expire the access token to trigger refresh
		transport := httpClient.Transport.(*tokenTransport)
		transport.expiresAt = time.Now().Add(-1 * time.Minute)

		// Make a request - should trigger refresh but reuse cached ID token
		apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer apiSrv.Close()

		req, _ := http.NewRequest("GET", apiSrv.URL, nil)
		_, err = httpClient.Do(req)
		require.NoError(t, err)

		// Generator should still be 1 because ID token is cached and still valid
		assert.Equal(t, 1, generatorCallCount)
	})

	t.Run("generator called when cached ID token expires", func(t *testing.T) {
		expiredToken := createIDToken(time.Now().Add(-1 * time.Hour).Unix())
		freshToken := createIDToken(time.Now().Add(1 * time.Hour).Unix())

		generatorCallCount := 0
		tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v2/oauth/token-exchange" {
				json.NewEncoder(w).Encode(TokenExchangeResponse{
					AccessToken: "test-access-token",
					TokenType:   "Bearer",
					ExpiresIn:   3600,
				})
			}
		}))
		defer tokenSrv.Close()

		httpClient, err := IdentityFederationConfig{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				generatorCallCount++
				if generatorCallCount == 1 {
					return expiredToken, nil
				}
				return freshToken, nil
			},
			BaseURL: tokenSrv.URL,
		}.HTTPClient()

		// First call should fail due to expired ID token
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ID token has expired")
		assert.Equal(t, 1, generatorCallCount)

		// Now test with a valid initial token that we manually expire
		generatorCallCount = 0
		httpClient, err = IdentityFederationConfig{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				generatorCallCount++
				return freshToken, nil
			},
			BaseURL: tokenSrv.URL,
		}.HTTPClient()
		require.NoError(t, err)
		assert.Equal(t, 1, generatorCallCount)

		// Manually expire both the access token and ID token
		transport := httpClient.Transport.(*tokenTransport)
		transport.expiresAt = time.Now().Add(-1 * time.Minute)
		transport.idToken = expiredToken

		// Make a request - should call generator because cached ID token is expired
		apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer apiSrv.Close()

		req, _ := http.NewRequest("GET", apiSrv.URL, nil)
		_, err = httpClient.Do(req)
		require.NoError(t, err)

		// Generator should now be 2 because expired ID token was refreshed
		assert.Equal(t, 2, generatorCallCount)
	})
}

func createIDToken(exp int64) string {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf(`{"exp":%d}`, exp)))
	signature := base64.RawURLEncoding.EncodeToString([]byte("fake-signature"))
	return fmt.Sprintf("%s.%s.%s", header, payload, signature)
}

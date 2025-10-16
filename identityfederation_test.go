// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestClientWithIdentityFederation(t *testing.T) {
	validToken := createIDToken(time.Now().Add(1 * time.Hour).Unix())

	t.Run("success with static ID token", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v2/oauth/token-exchange" {
				json.NewEncoder(w).Encode(TokenExchangeResponse{
					AccessToken: "ts-api-test-token",
					TokenType:   "Bearer",
					ExpiresIn:   3600,
				})
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer srv.Close()

		baseURL, _ := url.Parse(srv.URL)
		client := &Client{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				return validToken, nil
			},
			BaseURL: baseURL,
		}

		// Make a request to trigger transport initialization
		req, _ := http.NewRequest("GET", srv.URL+"/test", nil)
		client.init()
		_, err := client.HTTP.Do(req)

		require.NoError(t, err)
	})

	t.Run("success with token generator", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v2/oauth/token-exchange" {
				json.NewEncoder(w).Encode(TokenExchangeResponse{
					AccessToken: "ts-api-test-token",
					TokenType:   "Bearer",
					ExpiresIn:   3600,
				})
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))
		defer srv.Close()

		generatorCalled := false
		baseURL, _ := url.Parse(srv.URL)
		client := &Client{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				generatorCalled = true
				return validToken, nil
			},
			BaseURL: baseURL,
		}

		// Make a request to trigger token exchange
		req, _ := http.NewRequest("GET", srv.URL+"/test", nil)
		client.init()
		_, err := client.HTTP.Do(req)

		require.NoError(t, err)
		assert.True(t, generatorCalled, "generator should be called during first request")
	})

	t.Run("expired ID token on first request", func(t *testing.T) {
		expiredToken := createIDToken(time.Now().Add(-1 * time.Hour).Unix())

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		baseURL, _ := url.Parse(srv.URL)
		client := &Client{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				return expiredToken, nil
			},
			BaseURL: baseURL,
		}

		// Make a request to trigger token exchange
		req, _ := http.NewRequest("GET", srv.URL+"/test", nil)
		client.init()
		_, err := client.HTTP.Do(req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "expired")
	})

	t.Run("generator returns error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		baseURL, _ := url.Parse(srv.URL)
		client := &Client{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				return "", fmt.Errorf("generator error")
			},
			BaseURL: baseURL,
		}

		// Make a request to trigger token exchange
		req, _ := http.NewRequest("GET", srv.URL+"/test", nil)
		client.init()
		_, err := client.HTTP.Do(req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to fetch ID token")
		assert.Contains(t, err.Error(), "generator error")
	})

	t.Run("invalid JWT format", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		baseURL, _ := url.Parse(srv.URL)
		client := &Client{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				return "not.a.valid.jwt", nil
			},
			BaseURL: baseURL,
		}

		// Make a request to trigger token exchange
		req, _ := http.NewRequest("GET", srv.URL+"/test", nil)
		client.init()
		_, err := client.HTTP.Do(req)

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

		baseURL, _ := url.Parse(tokenSrv.URL)
		client := &Client{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				return validToken, nil
			},
			BaseURL: baseURL,
		}
		client.init()

		req, _ := http.NewRequest("GET", apiSrv.URL, nil)
		_, err := client.HTTP.Do(req)
		require.NoError(t, err)

		assert.Equal(t, "Bearer test-access-token", capturedAuthHeader)
	})

	t.Run("generator called on expired token", func(t *testing.T) {
		freshToken := createIDToken(time.Now().Add(1 * time.Hour).Unix())

		generatorCallCount := 0
		exchangeCount := 0
		tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v2/oauth/token-exchange" {
				exchangeCount++
				// First exchange returns a token that expires in 1 second
				// Second exchange returns a token that expires in 1 hour
				expiresIn := 1
				if exchangeCount > 1 {
					expiresIn = 3600
				}
				json.NewEncoder(w).Encode(TokenExchangeResponse{
					AccessToken: "test-access-token",
					TokenType:   "Bearer",
					ExpiresIn:   expiresIn,
				})
			}
		}))
		defer tokenSrv.Close()

		baseURL, _ := url.Parse(tokenSrv.URL)
		client := &Client{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				generatorCallCount++
				return freshToken, nil
			},
			BaseURL: baseURL,
		}
		client.init()

		// Make initial request to trigger token exchange
		apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer apiSrv.Close()

		req, _ := http.NewRequest("GET", apiSrv.URL, nil)
		_, err := client.HTTP.Do(req)
		require.NoError(t, err)

		// Initial call should use generator
		assert.Equal(t, 1, generatorCallCount)
		assert.Equal(t, 1, exchangeCount)

		// Wait for token to expire (it was set to expire in 1 second)
		time.Sleep(2 * time.Second)

		// Make a request - should trigger refresh but reuse cached ID token
		req, _ = http.NewRequest("GET", apiSrv.URL, nil)
		_, err = client.HTTP.Do(req)
		require.NoError(t, err)

		// Generator should still be 1 because ID token is cached and still valid
		assert.Equal(t, 1, generatorCallCount)
		// But token exchange should have happened twice
		assert.Equal(t, 2, exchangeCount)
	})

	t.Run("generator called when cached ID token expires", func(t *testing.T) {
		// Create two different tokens - one that's short-lived, one that's long-lived
		shortLivedToken := createIDToken(time.Now().Add(2 * time.Second).Unix())
		longLivedToken := createIDToken(time.Now().Add(1 * time.Hour).Unix())

		generatorCallCount := 0
		exchangeCount := 0
		tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v2/oauth/token-exchange" {
				exchangeCount++
				// First exchange returns a token that expires in 1 second
				expiresIn := 1
				if exchangeCount > 1 {
					expiresIn = 3600
				}
				json.NewEncoder(w).Encode(TokenExchangeResponse{
					AccessToken: "test-access-token",
					TokenType:   "Bearer",
					ExpiresIn:   expiresIn,
				})
			}
		}))
		defer tokenSrv.Close()

		baseURL, _ := url.Parse(tokenSrv.URL)
		client := &Client{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				generatorCallCount++
				// First call returns short-lived token, second call returns long-lived token
				if generatorCallCount == 1 {
					return shortLivedToken, nil
				}
				return longLivedToken, nil
			},
			BaseURL: baseURL,
		}
		client.init()

		// Make initial request
		apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer apiSrv.Close()

		req, _ := http.NewRequest("GET", apiSrv.URL, nil)
		_, err := client.HTTP.Do(req)
		require.NoError(t, err)
		assert.Equal(t, 1, generatorCallCount)

		// Wait for both access token AND cached ID token to expire
		time.Sleep(3 * time.Second)

		// Make a request - should call generator again because cached ID token is expired
		req, _ = http.NewRequest("GET", apiSrv.URL, nil)
		_, err = client.HTTP.Do(req)
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

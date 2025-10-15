// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// IdentityFederationConfig encapsulates the configuration needed to obtain a Tailscale API token via workload identity federation.
type IdentityFederationConfig struct {
	// ClientID is the client ID of the federated identity Tailscale OAuth client.
	ClientID string
	// IDTokenFunc returns an identity token from the IdP to exchange for a Tailscale API token.
	// The client calls this function to obtain a fresh ID token and reauthenticate when the API token
	// and cached ID token have expired. For static tokens, return the token directly. If a static token
	// expires, the client cannot automatically refresh the API token; the consumer is responsible to create a new client
	// with a fresh ID token.
	IDTokenFunc func() (string, error)
	// BaseURL is an optional base URL for the API server to which we'll connect. Defaults to https://api.tailscale.com.
	BaseURL string
}

// TokenExchangeResponse represents the response from the Tailscale token exchange endpoint.
type TokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // in seconds
	Scope       string `json:"scope"`
}

// jwtClaims represents the claims in a JWT token (minimal set for validation).
type jwtClaims struct {
	Exp int64 `json:"exp"`
}

// tokenTransport implements http.RoundTripper and handles token exchange and caching.
type tokenTransport struct {
	config    IdentityFederationConfig
	transport http.RoundTripper

	idToken        string
	apiAccessToken string
	expiresAt      time.Time

	tokenRefreshMu sync.Mutex
}

// HTTPClient constructs an HTTP client that authenticates using identity federation.
// The client automatically handles token exchange and caching.
func (c IdentityFederationConfig) HTTPClient() (*http.Client, error) {
	if c.ClientID == "" {
		return nil, fmt.Errorf("ClientID is required")
	}
	if c.IDTokenFunc == nil {
		return nil, fmt.Errorf("IDTokenFunc is required")
	}
	if c.BaseURL == "" {
		c.BaseURL = defaultBaseURL.String()
	}

	transport := &tokenTransport{
		config:    c,
		transport: http.DefaultTransport,
	}

	// Perform initial token exchange to validate configuration early
	if err := transport.refreshToken(context.Background()); err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: transport,
		Timeout:   defaultHttpClientTimeout,
	}, nil
}

// RoundTrip executes a single HTTP transaction, refreshing the API access token if necessary.
func (t *tokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.apiAccessToken == "" || time.Now().After(t.expiresAt) {
		if err := t.refreshToken(req.Context()); err != nil {
			return nil, err
		}
	}

	reqClone := req.Clone(req.Context())
	reqClone.Header.Set("Authorization", "Bearer "+t.apiAccessToken)

	return t.transport.RoundTrip(reqClone)
}

// refreshToken performs the token exchange and updates the cached token.
func (t *tokenTransport) refreshToken(ctx context.Context) error {
	t.tokenRefreshMu.Lock()
	defer t.tokenRefreshMu.Unlock()

	if t.idToken == "" || validateIDToken(t.idToken) != nil {
		idToken, err := t.config.IDTokenFunc()
		if err != nil {
			return fmt.Errorf("failed to fetch ID token: %w", err)
		}
		if err := validateIDToken(idToken); err != nil {
			return fmt.Errorf("fetched ID token is invalid: %w", err)
		}
		t.idToken = idToken
	}

	exchangeURL := fmt.Sprintf("%s/api/v2/oauth/token-exchange", t.config.BaseURL)
	values := url.Values{
		"client_id": {t.config.ClientID},
		"jwt":       {t.idToken},
	}.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, exchangeURL, strings.NewReader(values))
	if err != nil {
		return fmt.Errorf("failed to create token exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Transport: t.transport,
		Timeout:   10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unexpected token exchange request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(b))
	}

	var tokenResp TokenExchangeResponse
	if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to decode token exchange response: %w", err)
	}

	t.apiAccessToken = tokenResp.AccessToken
	// Set expiration with a 5-minute buffer for safety
	t.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn)*time.Second - 5*time.Minute)

	return nil
}

// validateIDToken decodes and validates the ID token's expiration claim
// to give a more helpful error if the token is expired or malformed.
func validateIDToken(idToken string) error {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid JWT format: expected 3 parts separated by '.', got %d", len(parts))
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	if claims.Exp == 0 {
		return fmt.Errorf("JWT is missing 'exp' (expiration) claim")
	}

	expirationTime := time.Unix(claims.Exp, 0)
	if time.Now().After(expirationTime) {
		return fmt.Errorf("ID token has expired (expired at %s)", expirationTime.Format(time.RFC3339))
	}

	return nil
}

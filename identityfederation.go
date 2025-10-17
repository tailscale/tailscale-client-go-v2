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

	"golang.org/x/oauth2"
)

var _ Auth = &IdentityFederation{}

// tokenExchangeResponse represents the response from the Tailscale token exchange endpoint.
type tokenExchangeResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // in seconds
	Scope       string `json:"scope"`
}

// jwtClaims represents the claims in a JWT token (minimal set for validation).
type jwtClaims struct {
	Exp int64 `json:"exp"`
}

// IdentityFederation configures identity federation authentication.
type IdentityFederation struct {
	// ClientID is the ID of the Tailscale OAuth client.
	ClientID string
	// IDTokenFunc returns an identity token from the IdP to exchange for a Tailscale API token.
	// The client calls this function to obtain a fresh ID token and reauthenticate when the API token
	// and cached ID token have expired. For static tokens, return the token directly. If a static token
	// expires, the client cannot automatically refresh the API token; the consumer is responsible to create a new client
	// with a fresh ID token.
	IDTokenFunc func() (string, error)
}

// identityFederationTokenSource implements oauth2.TokenSource using identity federation.
type identityFederationTokenSource struct {
	http        *http.Client
	baseURL     string
	clientID    string
	idTokenFunc func() (string, error)

	mu      sync.Mutex // protects the below fields
	idToken string
}

// HTTPClient implements the [Auth] interface.
func (i *IdentityFederation) HTTPClient(orig *http.Client, baseURL string) *http.Client {
	s := &identityFederationTokenSource{
		http:        orig,
		baseURL:     baseURL,
		clientID:    i.ClientID,
		idTokenFunc: i.IDTokenFunc,
	}

	return &http.Client{
		Transport: &oauth2.Transport{
			Base:   orig.Transport,
			Source: oauth2.ReuseTokenSource(nil, s),
		},
		CheckRedirect: orig.CheckRedirect,
		Jar:           orig.Jar,
		Timeout:       orig.Timeout,
	}
}

// Token implements oauth2.TokenSource by exchanging an ID token for an API access token.
func (i *identityFederationTokenSource) Token() (*oauth2.Token, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.idToken == "" || validateIDToken(i.idToken) != nil {
		idToken, err := i.idTokenFunc()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch ID token: %w", err)
		}
		if err := validateIDToken(idToken); err != nil {
			return nil, fmt.Errorf("fetched ID token is invalid: %w", err)
		}
		i.idToken = idToken
	}

	exchangeURL := fmt.Sprintf("%s/api/v2/oauth/token-exchange", i.baseURL)
	values := url.Values{
		"client_id": {i.clientID},
		"jwt":       {i.idToken},
	}.Encode()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, exchangeURL, strings.NewReader(values))
	if err != nil {
		return nil, fmt.Errorf("failed to create token exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := i.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unexpected token exchange request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(b))
	}

	var tokenResp tokenExchangeResponse
	if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token exchange response: %w", err)
	}

	return &oauth2.Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		Expiry:      time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}, nil
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

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

// identityFederationTokenSource implements oauth2.TokenSource using identity federation.
type identityFederationTokenSource struct {
	transport   http.RoundTripper
	baseURL     string
	clientID    string
	idTokenFunc func() (string, error)

	mu      sync.Mutex // protects the below fields
	idToken string
	tokenSource oauth2.TokenSource
}

// Token implements oauth2.TokenSource by exchanging an ID token for an API access token.
func (s *identityFederationTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.tokenSource != nil {
		token , err := s.tokenSource.Token()
		if err == nil && token.Valid() {
			return token, nil
		}
	}

	if s.idToken == "" || validateIDToken(s.idToken) != nil {
		idToken, err := s.idTokenFunc()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch ID token: %w", err)
		}
		if err := validateIDToken(idToken); err != nil {
			return nil, fmt.Errorf("fetched ID token is invalid: %w", err)
		}
		s.idToken = idToken
	}

	exchangeURL := fmt.Sprintf("%s/api/v2/oauth/token-exchange", s.baseURL)
	values := url.Values{
		"client_id": {s.clientID},
		"jwt":       {s.idToken},
	}.Encode()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, exchangeURL, strings.NewReader(values))
	if err != nil {
		return nil, fmt.Errorf("failed to create token exchange request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.transport.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("unexpected token exchange request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(b))
	}

	var tokenResp TokenExchangeResponse
	if err = json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token exchange response: %w", err)
	}

	s.tokenSource = oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: tokenResp.AccessToken,
		TokenType:   tokenResp.TokenType,
		Expiry:      time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	})

	return s.tokenSource.Token()
}

// newIdentityFederationTransport creates an http.RoundTripper that handles identity federation authentication.
func newIdentityFederationTransport(baseTransport http.RoundTripper, baseURL, clientID string, idTokenFunc func() (string, error)) http.RoundTripper {
	tokenSource := &identityFederationTokenSource{
		transport:   baseTransport,
		baseURL:     baseURL,
		clientID:    clientID,
		idTokenFunc: idTokenFunc,
	}

	return &oauth2.Transport{
		Source: oauth2.ReuseTokenSource(nil, tokenSource),
		Base:   baseTransport,
	}
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

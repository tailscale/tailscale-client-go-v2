// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Ensure that [OAuth] implements the [Auth] interface.
var _ Auth = &OAuth{}

// OAuth configures OAuth authentication.
type OAuth struct {
	// ClientID is the client ID of the OAuth client.
	ClientID string
	// ClientSecret is the client secret of the OAuth client.
	ClientSecret string
	// Scopes are the scopes to request when generating tokens for this OAuth client.
	Scopes []string
}

// HTTPClient implements the [Auth] interface.
func (o *OAuth) HTTPClient(orig *http.Client, baseURL string) *http.Client {
	oauthConfig := clientcredentials.Config{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		Scopes:       o.Scopes,
		TokenURL:     baseURL + "/api/v2/oauth/token",
	}

	// Use context.Background() here, since this is used to refresh the token in the future.
	tokenSource := oauthConfig.TokenSource(context.Background())

	return &http.Client{
		Transport: &oauth2.Transport{
			Base:   orig.Transport,
			Source: oauth2.ReuseTokenSource(nil, tokenSource),
		},
		CheckRedirect: orig.CheckRedirect,
		Jar:           orig.Jar,
		Timeout:       orig.Timeout,
	}
}

// OAuthConfig provides a mechanism for configuring OAuth authentication.
// Deprecated: use [OAuth] instead.
type OAuthConfig struct {
	// ClientID is the client ID of the OAuth client.
	ClientID string
	// ClientSecret is the client secret of the OAuth client.
	ClientSecret string
	// Scopes are the scopes to request when generating tokens for this OAuth client.
	Scopes []string
	// BaseURL is an optional base URL for the API server to which we'll connect. Defaults to https://api.tailscale.com.
	BaseURL string
}

// HTTPClient constructs an HTTP client that authenticates using OAuth.
// Deprecated: use [OAuth] instead.
func (ocfg OAuthConfig) HTTPClient() *http.Client {
	baseURL := ocfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL.String()
	}
	oauthConfig := clientcredentials.Config{
		ClientID:     ocfg.ClientID,
		ClientSecret: ocfg.ClientSecret,
		Scopes:       ocfg.Scopes,
		TokenURL:     baseURL + "/api/v2/oauth/token",
	}

	// Use context.Background() here, since this is used to refresh the token in the future.
	client := oauthConfig.Client(context.Background())
	client.Timeout = defaultHttpClientTimeout
	return client
}

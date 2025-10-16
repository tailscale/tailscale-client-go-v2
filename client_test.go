// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	_ "embed"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestErrorData(t *testing.T) {
	t.Parallel()

	t.Run("It should return the data element from a valid error", func(t *testing.T) {
		expected := APIError{
			Data: []APIErrorData{
				{
					User: "user1@example.com",
					Errors: []string{
						"address \"user2@example.com:400\": want: Accept, got: Drop",
					},
				},
			},
		}

		actual := ErrorData(expected)
		assert.EqualValues(t, expected.Data, actual)
	})

	t.Run("It should return an empty slice for any other error", func(t *testing.T) {
		assert.Empty(t, ErrorData(io.EOF))
	})
}

func Test_BuildTailnetURL(t *testing.T) {
	t.Parallel()

	base, err := url.Parse("http://example.com")
	require.NoError(t, err)

	c := &Client{
		BaseURL: base,
		Tailnet: "tn/with/slashes",
	}
	actual := c.buildTailnetURL("component/with/slashes")
	expected, err := url.Parse("http://example.com/api/v2/tailnet/tn%2Fwith%2Fslashes/component%2Fwith%2Fslashes")
	require.NoError(t, err)
	assert.EqualValues(t, expected.String(), actual.String())
}

func Test_BuildTailnetURLDefault(t *testing.T) {
	t.Parallel()

	base, err := url.Parse("http://example.com")
	require.NoError(t, err)

	c := &Client{
		BaseURL: base,
	}
	c.init()
	actual := c.buildTailnetURL("path")
	expected, err := url.Parse("http://example.com/api/v2/tailnet/-/path")
	require.NoError(t, err)
	assert.EqualValues(t, expected.String(), actual.String())
}

func Test_ClientAuthentication(t *testing.T) {
	t.Parallel()

	t.Run("OAuth transport is set when ClientSecret are provided", func(t *testing.T) {
		c := &Client{
			ClientSecret: "tskey-client-abc123-xyz789",
			Scopes:       []string{"all:read"},
		}
		c.init()

		assert.NotNil(t, c.HTTP)
		assert.NotNil(t, c.HTTP.Transport)
		_, ok := c.HTTP.Transport.(*oauth2.Transport)
		assert.True(t, ok, "expected transport to be *oauth2.Transport")
	})

	t.Run("OAuth transport is set when Federated Identity config is provided", func(t *testing.T) {
		c := &Client{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				return "test-token", nil
			},
		}
		c.init()

		assert.NotNil(t, c.HTTP)
		assert.NotNil(t, c.HTTP.Transport)
		_, ok := c.HTTP.Transport.(*oauth2.Transport)
		assert.True(t, ok, "expected transport to be *oauth2.Transport")
	})

	t.Run("OAuth wraps custom transport preserving proxy settings", func(t *testing.T) {
		proxyURL, _ := url.Parse("http://proxy.example.com:8080")
		customTransport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		c := &Client{
			ClientSecret: "tskey-client-abc123-xyz789",
			HTTP: &http.Client{
				Transport: customTransport,
			},
		}
		c.init()

		assert.NotNil(t, c.HTTP)
		assert.NotNil(t, c.HTTP.Transport)
		oauthTransport, ok := c.HTTP.Transport.(*oauth2.Transport)
		assert.True(t, ok, "expected transport to be *oauth2.Transport")

		// Verify the custom transport with proxy is wrapped
		wrappedTransport, ok := oauthTransport.Base.(*http.Transport)
		assert.True(t, ok, "underlying transport should be *http.Transport")
		assert.NotNil(t, wrappedTransport.Proxy, "proxy setting should be preserved")
	})

	t.Run("Identity federation wraps custom transport preserving proxy settings", func(t *testing.T) {
		proxyURL, _ := url.Parse("http://proxy.example.com:8080")
		customTransport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
		c := &Client{
			ClientID: "test-client-id",
			IDTokenFunc: func() (string, error) {
				return "test-token", nil
			},
			HTTP: &http.Client{
				Transport: customTransport,
			},
		}
		c.init()

		assert.NotNil(t, c.HTTP)
		assert.NotNil(t, c.HTTP.Transport)
		tokenTransport, ok := c.HTTP.Transport.(*oauth2.Transport)
		assert.True(t, ok, "expected transport to be *oauth2.Transport")

		// Verify the custom transport with proxy is wrapped
		wrappedTransport, ok := tokenTransport.Base.(*http.Transport)
		assert.True(t, ok, "underlying transport should be *http.Transport")
		assert.NotNil(t, wrappedTransport.Proxy, "proxy setting should be preserved")
	})
}

func Test_DeriveClientID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		clientSecret string
		want         string
	}{
		{
			name:         "Valid client secret with standard format",
			clientSecret: "tskey-client-abc123-xyz789",
			want:         "abc123",
		},
		{
			name:         "Client secret with unexpected shape",
			clientSecret: "plaintext",
			want:         defaultClientID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := deriveClientID(tt.clientSecret)
			assert.Equal(t, tt.want, got)
		})
	}
}

func ptrTo[T any](v T) *T {
	return &v
}

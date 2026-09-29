// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	_ "embed"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func ptrTo[T any](v T) *T {
	return &v
}

func TestIsNotFound(t *testing.T) {
	t.Parallel()

	e := APIError{Status: http.StatusNotFound}
	assert.True(t, IsNotFound(e))
}

func TestNonJSONErrorResponse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		status      int
		body        []byte
		wantMessage string
	}{
		{name: "empty 404", status: http.StatusNotFound, wantMessage: "Not Found"},
		{name: "plain text 429", status: http.StatusTooManyRequests, body: []byte("slow down\n"), wantMessage: "slow down"},
		{name: "html 502", status: http.StatusBadGateway, body: []byte("<html><body>Bad Gateway</body></html>"), wantMessage: "<html><body>Bad Gateway</body></html>"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, server := NewTestHarness(t)
			server.ResponseCode = tc.status
			if tc.body != nil {
				server.ResponseBody = tc.body
			}

			_, err := client.Devices().Get(context.Background(), "12345")

			var apiErr APIError
			require.ErrorAs(t, err, &apiErr)
			assert.Equal(t, tc.status, apiErr.Status)
			assert.Equal(t, tc.wantMessage, apiErr.Message)
			assert.Equal(t, tc.status == http.StatusNotFound, IsNotFound(err))
		})
	}
}

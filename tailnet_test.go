// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_Tailnets_Create(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	createRequest := CreateTailnetRequest{
		DisplayName: "Example Tailnet",
	}
	expected := &Tailnet{
		ID:          "T123456CNTRL",
		DisplayName: createRequest.DisplayName,
		OrgID:       "o123456CNTRL",
		DNSName:     "tail1234.ts.net",
		CreatedAt:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
		OAuthClient: &TailnetOAuthClient{
			ID:     "k123456CNTRL",
			Secret: "tskey-client-xxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		},
	}
	server.ResponseBody = expected

	actual, err := client.Tailnets().Create(context.Background(), createRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.MethodPost, server.Method)
	assert.Equal(t, "/api/v2/organizations/-/tailnets", server.Path)
	assert.Equal(t, expected, actual)
	assert.Equal(t, expected.OAuthClient.Secret, actual.OAuthClient.Secret)

	var receivedRequest CreateTailnetRequest
	err = json.Unmarshal(server.Body.Bytes(), &receivedRequest)
	assert.NoError(t, err)
	assert.Equal(t, createRequest, receivedRequest)
}

func TestClient_Tailnets_List(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expectedTailnets := map[string][]Tailnet{
		"tailnets": {
			{
				ID:          "T123456CNTRL",
				DisplayName: "Example Tailnet",
				OrgID:       "o123456CNTRL",
				CreatedAt:   time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
			},
			{
				ID:          "T654321CNTRL",
				DisplayName: "Another Tailnet",
				OrgID:       "o123456CNTRL",
				CreatedAt:   time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC),
			},
		},
	}
	server.ResponseBody = expectedTailnets

	actual, err := client.Tailnets().List(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/organizations/-/tailnets", server.Path)
	assert.Equal(t, expectedTailnets["tailnets"], actual)
}

func TestClient_Tailnets_Delete(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	err := client.Tailnets().Delete(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, http.MethodDelete, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com", server.Path)
}

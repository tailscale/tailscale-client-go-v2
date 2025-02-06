// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_Users_List(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expectedUsers := map[string][]User{
		"users": {
			{
				ID:                 "12345",
				DisplayName:        "Jane Doe",
				LoginName:          "janedoe",
				ProfilePicURL:      "http://example.com/users/janedoe",
				TailnetID:          "1",
				Created:            time.Date(2022, 2, 10, 11, 50, 23, 0, time.UTC),
				Type:               UserTypeMember,
				Role:               UserRoleOwner,
				Status:             UserStatusActive,
				DeviceCount:        2,
				LastSeen:           time.Date(2022, 2, 10, 12, 50, 23, 0, time.UTC),
				CurrentlyConnected: true,
			},
			{
				ID:                 "12346",
				DisplayName:        "John Doe",
				LoginName:          "johndoe",
				ProfilePicURL:      "http://example.com/users/johndoe",
				TailnetID:          "2",
				Created:            time.Date(2022, 2, 10, 11, 50, 23, 12, time.UTC),
				Type:               UserTypeShared,
				Role:               UserRoleMember,
				Status:             UserStatusIdle,
				DeviceCount:        2,
				LastSeen:           time.Date(2022, 2, 10, 12, 50, 23, 12, time.UTC),
				CurrentlyConnected: true,
			},
		},
	}
	server.ResponseBody = expectedUsers

	actualUsers, err := client.Users().List(
		context.Background(),
		PointerTo(UserTypeMember),
		PointerTo(UserRoleAdmin))
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/users", server.Path)
	assert.Equal(t, url.Values{"type": []string{"member"}, "role": []string{"admin"}}, server.Query)
	assert.Equal(t, expectedUsers["users"], actualUsers)
}

func TestClient_Users_Get(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expectedUser := &User{
		ID:                 "12345",
		DisplayName:        "Jane Doe",
		LoginName:          "janedoe",
		ProfilePicURL:      "http://example.com/users/janedoe",
		TailnetID:          "1",
		Created:            time.Date(2022, 2, 10, 11, 50, 23, 0, time.UTC),
		Type:               UserTypeMember,
		Role:               UserRoleOwner,
		Status:             UserStatusActive,
		DeviceCount:        2,
		LastSeen:           time.Date(2022, 2, 10, 12, 50, 23, 0, time.UTC),
		CurrentlyConnected: true,
	}
	server.ResponseBody = expectedUser

	actualUser, err := client.Users().Get(context.Background(), "12345")
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/users/12345", server.Path)
	assert.Equal(t, expectedUser, actualUser)
}

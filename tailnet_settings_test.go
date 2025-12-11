// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_TailnetSettings_Get(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expected := TailnetSettings{
		ACLsExternallyManagedOn:                true,
		ACLsExternalLink:                       "https://foo.com",
		DevicesApprovalOn:                      true,
		DevicesAutoUpdatesOn:                   true,
		DevicesKeyDurationDays:                 5,
		UsersApprovalOn:                        true,
		UsersRoleAllowedToJoinExternalTailnets: RoleAllowedToJoinExternalTailnetsMember,
		NetworkFlowLoggingOn:                   true,
		RegionalRoutingOn:                      true,
		PostureIdentityCollectionOn:            true,
		HTTPSEnabled:                           true,
	}
	server.ResponseBody = expected

	actual, err := client.TailnetSettings().Get(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/settings", server.Path)
	assert.Equal(t, &expected, actual)
}

func TestClient_TailnetSettings_Update(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK
	server.ResponseBody = nil

	updateRequest := UpdateTailnetSettingsRequest{
		ACLsExternallyManagedOn:                PointerTo(true),
		ACLsExternalLink:                       PointerTo("https://foo.com"),
		DevicesApprovalOn:                      PointerTo(true),
		DevicesAutoUpdatesOn:                   PointerTo(true),
		DevicesKeyDurationDays:                 PointerTo(5),
		UsersApprovalOn:                        PointerTo(true),
		UsersRoleAllowedToJoinExternalTailnets: PointerTo(RoleAllowedToJoinExternalTailnetsMember),
		NetworkFlowLoggingOn:                   PointerTo(true),
		RegionalRoutingOn:                      PointerTo(true),
		PostureIdentityCollectionOn:            PointerTo(true),
		HTTPSEnabled:                           PointerTo(true),
	}
	err := client.TailnetSettings().Update(context.Background(), updateRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.MethodPatch, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/settings", server.Path)
	var receivedRequest UpdateTailnetSettingsRequest
	err = json.Unmarshal(server.Body.Bytes(), &receivedRequest)
	assert.NoError(t, err)
	assert.EqualValues(t, updateRequest, receivedRequest)
}

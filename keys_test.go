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

func TestClient_CreateKey(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	capabilities := KeyCapabilities{}
	capabilities.Devices.Create.Ephemeral = true
	capabilities.Devices.Create.Reusable = true
	capabilities.Devices.Create.Preauthorized = true
	capabilities.Devices.Create.Tags = []string{"test:test"}

	expected := &Key{
		ID:           "test",
		Key:          "thisisatestkey",
		Created:      time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Expires:      time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Capabilities: capabilities,
		Description:  "",
	}

	server.ResponseBody = expected

	actual, err := client.Keys().Create(context.Background(), CreateKeyRequest{
		Capabilities: capabilities,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, expected, actual)
	assert.Equal(t, http.MethodPost, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/keys", server.Path)

	var actualReq CreateKeyRequest
	assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &actualReq))
	assert.EqualValues(t, capabilities, actualReq.Capabilities)
	assert.EqualValues(t, 0, actualReq.ExpirySeconds)
	assert.EqualValues(t, "", actualReq.Description)
}

func TestClient_CreateKeyWithExpirySeconds(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	capabilities := KeyCapabilities{}
	capabilities.Devices.Create.Ephemeral = true
	capabilities.Devices.Create.Reusable = true
	capabilities.Devices.Create.Preauthorized = true
	capabilities.Devices.Create.Tags = []string{"test:test"}

	expected := &Key{
		ID:           "test",
		Key:          "thisisatestkey",
		Created:      time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Expires:      time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Capabilities: capabilities,
		Description:  "",
	}

	server.ResponseBody = expected

	actual, err := client.Keys().Create(context.Background(), CreateKeyRequest{
		Capabilities:  capabilities,
		ExpirySeconds: 1440,
	})
	assert.NoError(t, err)
	assert.EqualValues(t, expected, actual)
	assert.Equal(t, http.MethodPost, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/keys", server.Path)

	var actualReq CreateKeyRequest
	assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &actualReq))
	assert.EqualValues(t, capabilities, actualReq.Capabilities)
	assert.EqualValues(t, 1440, actualReq.ExpirySeconds)
	assert.EqualValues(t, "", actualReq.Description)
}

func TestClient_CreateKeyWithDescription(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	capabilities := KeyCapabilities{}
	capabilities.Devices.Create.Ephemeral = true
	capabilities.Devices.Create.Reusable = true
	capabilities.Devices.Create.Preauthorized = true
	capabilities.Devices.Create.Tags = []string{"test:test"}

	expected := &Key{
		ID:           "test",
		Key:          "thisisatestkey",
		Created:      time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Expires:      time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Capabilities: capabilities,
		Description:  "key description",
	}

	server.ResponseBody = expected

	actual, err := client.Keys().Create(context.Background(), CreateKeyRequest{
		Capabilities: capabilities,
		Description:  "key description",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, expected, actual)
	assert.Equal(t, http.MethodPost, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/keys", server.Path)

	var actualReq CreateKeyRequest
	assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &actualReq))
	assert.EqualValues(t, capabilities, actualReq.Capabilities)
	assert.EqualValues(t, 0, actualReq.ExpirySeconds)
	assert.EqualValues(t, "key description", actualReq.Description)
}

func TestClient_GetKey(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	capabilities := KeyCapabilities{}
	capabilities.Devices.Create.Ephemeral = true
	capabilities.Devices.Create.Reusable = true
	capabilities.Devices.Create.Preauthorized = true
	capabilities.Devices.Create.Tags = []string{"test:test"}

	expected := &Key{
		ID:           "test",
		Created:      time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Expires:      time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		Capabilities: capabilities,
		Description:  "",
	}

	server.ResponseBody = expected

	actual, err := client.Keys().Get(context.Background(), expected.ID)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, actual)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/keys/"+expected.ID, server.Path)
}

func TestClient_Keys(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expected := []Key{
		{ID: "key-a"},
		{ID: "key-b"},
	}

	server.ResponseBody = map[string][]Key{
		"keys": expected,
	}

	actual, err := client.Keys().List(context.Background(), false)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, actual)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/keys", server.Path)
}

func TestClient_DeleteKey(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	const keyID = "test"

	assert.NoError(t, client.Keys().Delete(context.Background(), keyID))
	assert.Equal(t, http.MethodDelete, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/keys/"+keyID, server.Path)
}

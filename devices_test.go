// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	//go:embed testdata/devices.json
	jsonDevices []byte
)

func TestClient_SetDeviceSubnetRoutes(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	const deviceID = "test"
	routes := []string{"127.0.0.1"}

	assert.NoError(t, client.Devices().SetSubnetRoutes(context.Background(), deviceID, routes))
	assert.Equal(t, http.MethodPost, server.Method)
	assert.Equal(t, "/api/v2/device/test/routes", server.Path)

	body := make(map[string][]string)
	assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &body))
	assert.EqualValues(t, routes, body["routes"])
}

func TestClient_Devices_Get(t *testing.T) {
	t.Parallel()

	expectedDevice := &Device{
		Addresses:         []string{"127.0.0.1"},
		Name:              "test",
		ID:                "testid",
		Authorized:        true,
		KeyExpiryDisabled: true,
		User:              "test@example.com",
		Tags: []string{
			"tag:value",
		},
		BlocksIncomingConnections: false,
		ClientVersion:             "1.22.1",
		Created:                   Time{time.Date(2022, 2, 10, 11, 50, 23, 0, time.UTC)},
		Expires:                   Time{time.Date(2022, 8, 9, 11, 50, 23, 0, time.UTC)},
		Hostname:                  "test",
		IsExternal:                false,
		LastSeen:                  Time{time.Date(2022, 3, 9, 20, 3, 42, 0, time.UTC)},
		MachineKey:                "mkey:test",
		NodeKey:                   "nodekey:test",
		OS:                        "windows",
		TailnetLockError:          "test error",
		TailnetLockKey:            "tlpub:test",
		UpdateAvailable:           true,
		AdvertisedRoutes:          []string{"127.0.0.1", "127.0.0.2"},
		EnabledRoutes:             []string{"127.0.0.1"},
		ClientConnectivity: &ClientConnectivity{
			Endpoints: []string{"199.9.14.201:59128", "192.68.0.21:59128"},
			DERP:      "New York City",
			DERPLatency: map[string]DERPRegion{
				"Dallas": {
					LatencyMilliseconds: 60.463043,
				},
				"New York City": {
					Preferred:           true,
					LatencyMilliseconds: 31.323811,
				},
			},
			MappingVariesByDestIP: true,
			ClientSupports: map[string]bool{
				"hairPinning": false,
				"ipv6":        false,
				"pcp":         false,
				"pmp":         false,
				"udp":         false,
				"upnp":        false,
			},
		},
	}

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK
	server.ResponseBody = expectedDevice

	actualDevice, err := client.Devices().Get(context.Background(), "testid")
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/device/testid", server.Path)
	assert.EqualValues(t, expectedDevice, actualDevice)
}

func TestClient_Devices_GetPostureAttributes(t *testing.T) {
	t.Parallel()

	expectedAttributes := &DevicePostureAttributes{
		Attributes: map[string]interface{}{
			"custom:key":          "value",
			"node:os":             "linux",
			"node:osVersion":      "5.19.0-42-generic",
			"node:tsReleaseTrack": "stable",
			"node:tsVersion":      "1.40.0",
			"node:tsAutoUpdate":   false,
		},
		Expiries: map[string]Time{
			"custom:key": {time.Date(2022, 2, 10, 11, 50, 23, 0, time.UTC)},
		},
	}

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK
	server.ResponseBody = expectedAttributes

	actualAttributes, err := client.Devices().GetPostureAttributes(context.Background(), "testid")
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/device/testid/attributes", server.Path)

	assert.EqualValues(t, expectedAttributes, actualAttributes)
}

func TestClient_Devices_List(t *testing.T) {
	t.Parallel()

	expectedDevices := map[string][]Device{
		"devices": {
			{
				Addresses:         []string{"127.0.0.1"},
				Name:              "test",
				ID:                "test",
				Authorized:        true,
				KeyExpiryDisabled: true,
				User:              "test@example.com",
				Tags: []string{
					"tag:value",
				},
				BlocksIncomingConnections: false,
				ClientVersion:             "1.22.1",
				Created:                   Time{time.Date(2022, 2, 10, 11, 50, 23, 0, time.UTC)},
				Expires:                   Time{time.Date(2022, 8, 9, 11, 50, 23, 0, time.UTC)},
				Hostname:                  "test",
				IsExternal:                false,
				LastSeen:                  Time{time.Date(2022, 3, 9, 20, 3, 42, 0, time.UTC)},
				MachineKey:                "mkey:test",
				NodeKey:                   "nodekey:test",
				OS:                        "windows",
				UpdateAvailable:           true,
				AdvertisedRoutes:          []string{"127.0.0.1", "127.0.0.2"},
				EnabledRoutes:             []string{"127.0.0.1"},
				ClientConnectivity: &ClientConnectivity{
					Endpoints: []string{"199.9.14.201:59128", "192.68.0.21:59128"},
					DERP:      "New York City",
					DERPLatency: map[string]DERPRegion{
						"Dallas": {
							LatencyMilliseconds: 60.463043,
						},
						"New York City": {
							Preferred:           true,
							LatencyMilliseconds: 31.323811,
						},
					},
					MappingVariesByDestIP: true,
					ClientSupports: map[string]bool{
						"hairPinning": false,
						"ipv6":        false,
						"pcp":         false,
						"pmp":         false,
						"udp":         false,
						"upnp":        false,
					},
				},
			},
		},
	}

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK
	server.ResponseBody = expectedDevices

	actualDevices, err := client.Devices().ListWithAllFields(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/devices", server.Path)
	assert.Equal(t, "all", server.Query.Get("fields"))
	assert.EqualValues(t, expectedDevices["devices"], actualDevices)
}

func TestDevices_Unmarshal(t *testing.T) {
	t.Parallel()

	tt := []struct {
		Name           string
		DevicesContent []byte
		Expected       []Device
		UnmarshalFunc  func(data []byte, v interface{}) error
	}{
		{
			Name:           "It should handle badly formed devices",
			DevicesContent: jsonDevices,
			UnmarshalFunc:  json.Unmarshal,
			Expected: []Device{
				{
					Addresses:                 []string{"100.101.102.103", "fd7a:115c:a1e0:ab12:4843:cd96:6265:6667"},
					Authorized:                true,
					BlocksIncomingConnections: false,
					ClientVersion:             "",
					Created:                   Time{},
					Expires: Time{
						time.Date(1, 1, 1, 00, 00, 00, 0, time.UTC),
					},
					Hostname:          "hello",
					ID:                "50052",
					IsExternal:        true,
					KeyExpiryDisabled: true,
					LastSeen: Time{
						time.Date(2022, 4, 15, 13, 24, 40, 0, time.UTC),
					},
					MachineKey:      "",
					Name:            "hello.example.com",
					NodeKey:         "nodekey:30dc3c061ac8b33fdc6d88a4a67b053b01b56930d78cae0cf7a164411d424c0d",
					OS:              "linux",
					UpdateAvailable: false,
					User:            "services@example.com",
				},
				{
					Addresses:                 []string{"100.121.200.21", "fd7a:115c:a1e0:ab12:4843:cd96:6265:e618"},
					Authorized:                true,
					BlocksIncomingConnections: false,
					ClientVersion:             "1.22.2-t60b671955-gecc5d9846",
					Created: Time{
						time.Date(2022, 3, 5, 17, 10, 27, 0, time.UTC),
					},
					Expires: Time{
						time.Date(2022, 9, 1, 17, 10, 27, 0, time.UTC),
					},
					Hostname:          "foo",
					ID:                "50053",
					IsExternal:        false,
					KeyExpiryDisabled: true,
					LastSeen: Time{
						time.Date(2022, 4, 15, 13, 25, 21, 0, time.UTC),
					},
					MachineKey:      "mkey:30dc3c061ac8b33fdc6d88a4a67b053b01b56930d78cae0cf7a164411d424c0d",
					Name:            "foo.example.com",
					NodeKey:         "nodekey:30dc3c061ac8b33fdc6d88a4a67b053b01b56930d78cae0cf7a164411d424c0d",
					OS:              "linux",
					UpdateAvailable: false,
					User:            "foo@example.com",
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			actual := make(map[string][]Device)

			assert.NoError(t, tc.UnmarshalFunc(tc.DevicesContent, &actual))
			assert.EqualValues(t, tc.Expected, actual["devices"])
		})
	}
}

func TestClient_DeleteDevice(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK
	ctx := context.Background()

	deviceID := "deviceTestId"
	assert.NoError(t, client.Devices().Delete(ctx, deviceID))
	assert.Equal(t, http.MethodDelete, server.Method)
	assert.Equal(t, "/api/v2/device/deviceTestId", server.Path)
}

func TestClient_DeviceSubnetRoutes(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK
	server.ResponseBody = &DeviceRoutes{
		Advertised: []string{"127.0.0.1"},
		Enabled:    []string{"127.0.0.1"},
	}

	const deviceID = "test"

	routes, err := client.Devices().SubnetRoutes(context.Background(), deviceID)
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/device/test/routes", server.Path)
	assert.Equal(t, server.ResponseBody, routes)
}

func TestClient_SetDeviceAuthorized(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	const deviceID = "test"

	for _, value := range []bool{true, false} {
		assert.NoError(t, client.Devices().SetAuthorized(context.Background(), deviceID, value))
		assert.Equal(t, http.MethodPost, server.Method)
		assert.Equal(t, "/api/v2/device/test/authorized", server.Path)

		body := make(map[string]bool)
		assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &body))
		assert.EqualValues(t, value, body["authorized"])
	}
}

func TestClient_SetDeviceName(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	const deviceID = "test"
	name := "test"

	assert.NoError(t, client.Devices().SetName(context.Background(), deviceID, name))
	assert.EqualValues(t, http.MethodPost, server.Method)
	assert.EqualValues(t, "/api/v2/device/"+deviceID+"/name", server.Path)

	body := make(map[string]string)
	assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &body))
	assert.EqualValues(t, name, body["name"])
}

func TestClient_SetDeviceTags(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	const deviceID = "test"
	tags := []string{"a:b", "b:c"}

	assert.NoError(t, client.Devices().SetTags(context.Background(), deviceID, tags))
	assert.EqualValues(t, http.MethodPost, server.Method)
	assert.EqualValues(t, "/api/v2/device/"+deviceID+"/tags", server.Path)

	body := make(map[string][]string)
	assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &body))
	assert.EqualValues(t, tags, body["tags"])
}

func TestClient_SetDevicePostureAttributes(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK
	server.ResponseBody = nil

	const deviceID = "test"
	const attributeKey = "custom:test"

	setRequest := DevicePostureAttributeRequest{
		Value:   "value",
		Expiry:  Time{time.Date(2022, 2, 10, 11, 50, 23, 0, time.UTC)},
		Comment: "test",
	}

	assert.NoError(t, client.Devices().SetPostureAttribute(context.Background(), deviceID, attributeKey, setRequest))
	assert.EqualValues(t, http.MethodPost, server.Method)
	assert.EqualValues(t, "/api/v2/device/"+deviceID+"/attributes/"+attributeKey, server.Path)

	var receivedRequest DevicePostureAttributeRequest
	err := json.Unmarshal(server.Body.Bytes(), &receivedRequest)
	assert.NoError(t, err)
	assert.EqualValues(t, setRequest, receivedRequest)
}

func TestClient_SetDeviceKey(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	const deviceID = "test"
	expected := DeviceKey{
		KeyExpiryDisabled: true,
	}

	assert.NoError(t, client.Devices().SetKey(context.Background(), deviceID, expected))

	assert.EqualValues(t, http.MethodPost, server.Method)
	assert.EqualValues(t, "/api/v2/device/"+deviceID+"/key", server.Path)

	var actual DeviceKey
	assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &actual))
	assert.EqualValues(t, expected, actual)
}

func TestClient_SetDeviceIPv4Address(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	const deviceID = "test"
	address := "100.64.0.1"

	assert.NoError(t, client.Devices().SetIPv4Address(context.Background(), deviceID, address))
	assert.Equal(t, http.MethodPost, server.Method)
	assert.EqualValues(t, "/api/v2/device/"+deviceID+"/ip", server.Path)
}

func TestClient_UserAgent(t *testing.T) {
	t.Parallel()
	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	// Check the default user-agent.
	assert.NoError(t, client.Devices().SetAuthorized(context.Background(), "test", true))
	assert.Equal(t, "tailscale-client-go", server.Header.Get("User-Agent"))

	// Check a custom user-agent.
	client = &Client{
		APIKey:    "fake key",
		BaseURL:   server.BaseURL,
		UserAgent: "custom-user-agent",
	}
	assert.NoError(t, client.Devices().SetAuthorized(context.Background(), "test", true))
	assert.Equal(t, "custom-user-agent", server.Header.Get("User-Agent"))
}

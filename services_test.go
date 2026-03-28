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

func TestClient_ListServices(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expected := []Service{
		{
			Name:    "svc:my-service",
			Addrs:   []string{"100.64.0.1", "fd7a:115c:a1e0::1"},
			Comment: "test service",
			Ports:   []string{"443"},
			Tags:    []string{"tag:web"},
		},
	}
	server.ResponseBody = serviceList{Services: expected}

	actual, err := client.Services().List(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/vip-services", server.Path)
	assert.Equal(t, expected, actual)
}

func TestClient_GetService(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expected := &Service{
		Name:    "svc:my-service",
		Addrs:   []string{"100.64.0.1", "fd7a:115c:a1e0::1"},
		Comment: "test service",
		Ports:   []string{"443"},
		Tags:    []string{"tag:web"},
	}
	server.ResponseBody = expected

	actual, err := client.Services().Get(context.Background(), "svc:my-service")
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/vip-services/svc:my-service", server.Path)
	assert.Equal(t, expected, actual)
}

func TestClient_CreateOrUpdateService(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	svc := Service{
		Name:    "svc:my-service",
		Comment: "new service",
		Ports:   []string{"443"},
		Tags:    []string{"tag:web"},
	}

	err := client.Services().CreateOrUpdate(context.Background(), svc)
	assert.NoError(t, err)
	assert.Equal(t, http.MethodPut, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/vip-services/svc:my-service", server.Path)

	var received Service
	err = json.Unmarshal(server.Body.Bytes(), &received)
	assert.NoError(t, err)
	assert.Equal(t, svc, received)
}

func TestClient_DeleteService(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	err := client.Services().Delete(context.Background(), "svc:my-service")
	assert.NoError(t, err)
	assert.Equal(t, http.MethodDelete, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/vip-services/svc:my-service", server.Path)
}

func TestClient_VIPServices_Deprecated(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expected := []Service{
		{
			Name:  "svc:my-service",
			Addrs: []string{"100.64.0.1", "fd7a:115c:a1e0::1"},
		},
	}
	server.ResponseBody = serviceList{Services: expected}

	actual, err := client.VIPServices().List(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestClient_GetService_NotFound(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusNotFound
	server.ResponseBody = APIError{Message: "not found"}

	_, err := client.Services().Get(context.Background(), "svc:nonexistent")
	assert.Error(t, err)
	assert.True(t, IsNotFound(err))
}

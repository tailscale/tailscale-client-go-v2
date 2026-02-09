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

func TestClient_CreateService(t *testing.T) {
	t.Parallel()
	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expected := &Service{
		Name:    "testservice",
		Comment: "a test comment",
		Addrs:   []string{"127.0.0.1"},
		Ports:   []string{"tcp:512"},
		Tags:    []string{"tag:thisisatest"},
	}
	server.ResponseBody = expected
	actual, err := client.Services().Update(context.Background(), *expected)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, actual)
	assert.Equal(t, http.MethodPut, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/services/testservice", server.Path)

	var actualReq Service
	assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &actualReq))
	assert.EqualValues(t, "a test comment", actualReq.Comment)
}

func TestClient_RenameService(t *testing.T) {
	t.Parallel()
	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expected := &Service{
		Name:    "testservice2",
		Comment: "a test comment",
		Addrs:   []string{"127.0.0.1"},
		Ports:   []string{"tcp:512"},
		Tags:    []string{"tag:thisisatest"},
	}
	server.ResponseBody = expected
	actual, err := client.Services().Update(context.Background(), *expected, WithServiceName("testservice"))
	assert.NoError(t, err)
	assert.EqualValues(t, expected, actual)
	assert.Equal(t, http.MethodPut, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/services/testservice", server.Path)

	var actualReq Service
	assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &actualReq))
	assert.EqualValues(t, "a test comment", actualReq.Comment)
}

func TestClient_GetService(t *testing.T) {
	t.Parallel()
	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expected := &Service{
		Name:    "testservice",
		Comment: "a test comment",
		Addrs:   []string{"127.0.0.1"},
		Ports:   []string{"tcp:512"},
		Tags:    []string{"tag:thisisatest"},
	}
	server.ResponseBody = expected
	actual, err := client.Services().Get(context.Background(), expected.Name)
	assert.NoError(t, err)
	assert.EqualValues(t, expected, actual)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/services/testservice", server.Path)
}

func TestClient_ListServices(t *testing.T) {
	t.Parallel()
	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expected := []Service{
		Service{
			Name:    "testservice1",
			Comment: "a test comment",
			Addrs:   []string{"127.0.0.1"},
			Ports:   []string{"tcp:512"},
			Tags:    []string{"tag:thisisatest"},
		},
		Service{
			Name:    "testservice2",
			Comment: "another test comment",
			Addrs:   []string{"127.0.0.2"},
			Ports:   []string{"tcp:513"},
			Tags:    []string{"tag:thisisatest2"},
		},
	}
	server.ResponseBody = map[string][]Service{
		"vipServices": expected,
	}
	actual, err := client.Services().List(context.Background())
	assert.NoError(t, err)
	assert.EqualValues(t, expected, actual)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/services", server.Path)
}

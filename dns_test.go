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

func TestClient_DNSNameservers(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expectedNameservers := map[string][]string{
		"dns": {"127.0.0.1"},
	}

	server.ResponseBody = expectedNameservers
	nameservers, err := client.DNS().Nameservers(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/dns/nameservers", server.Path)
	assert.Equal(t, expectedNameservers["dns"], nameservers)
}

func TestClient_DNSPreferences(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK
	server.ResponseBody = &DNSPreferences{
		MagicDNS: true,
	}

	preferences, err := client.DNS().Preferences(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/dns/preferences", server.Path)
	assert.Equal(t, server.ResponseBody, preferences)
}

func TestClient_DNSSearchPaths(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expectedPaths := map[string][]string{
		"searchPaths": {"test"},
	}

	server.ResponseBody = expectedPaths

	paths, err := client.DNS().SearchPaths(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/dns/searchpaths", server.Path)
	assert.Equal(t, expectedPaths["searchPaths"], paths)
}

func TestClient_SplitDNS(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expectedNameservers := SplitDNSResponse{
		"example.com": {"1.1.1.1", "1.2.3.4"},
	}

	server.ResponseBody = expectedNameservers
	nameservers, err := client.DNS().SplitDNS(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/dns/split-dns", server.Path)
	assert.EqualValues(t, expectedNameservers, nameservers)
}

func TestClient_SetDNSNameservers(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	nameservers := []string{"127.0.0.1"}

	assert.NoError(t, client.DNS().SetNameservers(context.Background(), nameservers))
	assert.Equal(t, http.MethodPost, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/dns/nameservers", server.Path)

	body := make(map[string][]string)
	assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &body))
	assert.EqualValues(t, nameservers, body["dns"])
}

func TestClient_SetDNSPreferences(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	preferences := DNSPreferences{
		MagicDNS: true,
	}

	assert.NoError(t, client.DNS().SetPreferences(context.Background(), preferences))
	assert.Equal(t, http.MethodPost, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/dns/preferences", server.Path)

	var body DNSPreferences
	assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &body))
	assert.EqualValues(t, preferences, body)
}

func TestClient_SetDNSSearchPaths(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	paths := []string{"test"}

	assert.NoError(t, client.DNS().SetSearchPaths(context.Background(), paths))
	assert.Equal(t, http.MethodPost, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/dns/searchpaths", server.Path)

	body := make(map[string][]string)
	assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &body))
	assert.EqualValues(t, paths, body["searchPaths"])
}

func TestClient_UpdateSplitDNS(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	nameservers := []string{"1.1.2.1", "3.3.3.4"}
	request := SplitDNSRequest{
		"example.com": nameservers,
	}

	expectedNameservers := SplitDNSResponse{
		"example.com": nameservers,
	}
	server.ResponseBody = expectedNameservers

	resp, err := client.DNS().UpdateSplitDNS(context.Background(), request)
	assert.NoError(t, err)
	assert.Equal(t, http.MethodPatch, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/dns/split-dns", server.Path)

	body := make(SplitDNSResponse)
	assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &body))
	assert.EqualValues(t, nameservers, body["example.com"])
	assert.EqualValues(t, expectedNameservers, resp)
}

func TestClient_SetSplitDNS(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	nameservers := []string{"1.1.2.1", "3.3.3.4"}
	request := SplitDNSRequest{
		"example.com": nameservers,
	}

	assert.NoError(t, client.DNS().SetSplitDNS(context.Background(), request))
	assert.Equal(t, http.MethodPut, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/dns/split-dns", server.Path)

	body := make(SplitDNSResponse)
	assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &body))
	assert.EqualValues(t, nameservers, body["example.com"])
}

func TestClient_DNSConfiguration(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK
	server.ResponseBody = &DNSConfiguration{
		Nameservers: []DNSConfigurationResolver{
			{
				Address:         "8.8.8.8",
				UseWithExitNode: true,
			},
		},
		SplitDNS: map[string][]DNSConfigurationResolver{
			"example.com": []DNSConfigurationResolver{
				{
					Address:         "4.4.4.4",
					UseWithExitNode: true,
				},
			},
		},
		Preferences: DNSConfigurationPreferences{
			OverrideLocalDNS: true,
			MagicDNS:         true,
		},
	}

	configuration, err := client.DNS().Configuration(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/dns/configuration", server.Path)
	assert.Equal(t, server.ResponseBody, configuration)
}

func TestClient_SetDNSConfiguration(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	configuration := DNSConfiguration{
		Nameservers: []DNSConfigurationResolver{
			{
				Address:         "8.8.8.8",
				UseWithExitNode: true,
			},
		},
		SplitDNS: map[string][]DNSConfigurationResolver{
			"example.com": {
				{
					Address:         "4.4.4.4",
					UseWithExitNode: true,
				},
			},
		},
		Preferences: DNSConfigurationPreferences{
			OverrideLocalDNS: true,
			MagicDNS:         true,
		},
	}

	assert.NoError(t, client.DNS().SetConfiguration(context.Background(), configuration))
	assert.Equal(t, http.MethodPost, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/dns/configuration", server.Path)

	var body DNSConfiguration
	assert.NoError(t, json.Unmarshal(server.Body.Bytes(), &body))
	assert.EqualValues(t, configuration, body)
}

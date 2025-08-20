// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_LogstreamConfiguration(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	expectedLogstream := &LogstreamConfiguration{
		DestinationType:      LogstreamCriblEndpoint,
		URL:                  "http://example.com",
		User:                 "my-user",
		UploadPeriodMinutes:  5,
		CompressionFormat:    CompressionFormatZstd,
		S3Bucket:             "my-bucket",
		S3Region:             "us-west-2",
		S3KeyPrefix:          "logs/",
		S3AuthenticationType: S3AccessKeyAuthentication,
		S3AccessKeyID:        "my-access-key-id",
		S3RoleARN:            "my-role-arn",
		S3ExternalID:         "my-external-id",
	}
	server.ResponseBody = expectedLogstream

	actualLogstream, err := client.Logging().LogstreamConfiguration(context.Background(), LogTypeConfig)
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/logging/configuration/stream", server.Path)
	assert.Equal(t, expectedLogstream, actualLogstream)
}

func TestClient_SetLogstreamConfiguration(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	logstreamRequest := SetLogstreamConfigurationRequest{
		DestinationType:      LogstreamCriblEndpoint,
		URL:                  "http://example.com",
		User:                 "my-user",
		Token:                "my-token",
		UploadPeriodMinutes:  5,
		CompressionFormat:    CompressionFormatZstd,
		S3Bucket:             "my-bucket",
		S3Region:             "us-west-2",
		S3KeyPrefix:          "logs/",
		S3AuthenticationType: S3AccessKeyAuthentication,
		S3AccessKeyID:        "my-access-key-id",
		S3SecretAccessKey:    "my-secret-access-key",
		S3RoleARN:            "my-role-arn",
		S3ExternalID:         "my-external-id",
	}
	server.ResponseBody = nil

	err := client.Logging().SetLogstreamConfiguration(context.Background(), LogTypeNetwork, logstreamRequest)
	assert.NoError(t, err)
	assert.Equal(t, http.MethodPut, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/logging/network/stream", server.Path)
	var receivedRequest SetLogstreamConfigurationRequest
	err = json.Unmarshal(server.Body.Bytes(), &receivedRequest)
	assert.NoError(t, err)
	assert.EqualValues(t, logstreamRequest, receivedRequest)
}

func TestClient_DeleteLogstream(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	err := client.Logging().DeleteLogstreamConfiguration(context.Background(), LogTypeConfig)
	assert.NoError(t, err)
	assert.Equal(t, http.MethodDelete, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/logging/configuration/stream", server.Path)
}

func TestClient_CreateOrGetAwsExternalId(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	wantExternalID := &AWSExternalID{
		ExternalID:            "external-id",
		TailscaleAWSAccountID: "account-id",
	}
	server.ResponseBody = wantExternalID

	gotExternalID, err := client.Logging().CreateOrGetAwsExternalId(context.Background(), true)
	assert.NoError(t, err)
	assert.Equal(t, server.Method, http.MethodPost)
	assert.Equal(t, server.Path, "/api/v2/tailnet/example.com/aws-external-id")
	assert.Equal(t, gotExternalID, wantExternalID)

	gotRequest := make(map[string]bool)
	err = json.Unmarshal(server.Body.Bytes(), &gotRequest)
	assert.NoError(t, err)
	assert.EqualValues(t, gotRequest, map[string]bool{"reusable": true})
}

func TestClient_ValidateAWSTrustPolicy(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	roleARN := "arn:aws:iam::123456789012:role/example-role"

	err := client.Logging().ValidateAWSTrustPolicy(context.Background(), "external-id-0000-0000", roleARN)
	assert.NoError(t, err)
	assert.Equal(t, server.Method, http.MethodPost)
	assert.Equal(t, server.Path, "/api/v2/tailnet/example.com/aws-external-id/external-id-0000-0000/validate-aws-trust-policy")

	gotRequest := make(map[string]string)
	err = json.Unmarshal(server.Body.Bytes(), &gotRequest)
	assert.NoError(t, err)
	assert.EqualValues(t, gotRequest, map[string]string{"roleArn": roleARN})
}

func TestClient_GetNetworkFlowLogs(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	now := time.Now().UTC().Truncate(time.Second)
	expectedLogs := []NetworkFlowLog{
		{
			Logged: now,
			NodeID: "node1",
			Start:  now.Add(-5 * time.Minute),
			End:    now,
			VirtualTraffic: []TrafficStats{
				{Proto: 6, Src: "10.0.0.1:80", Dst: "10.0.0.2:1234", TxPkts: 10, TxBytes: 1000},
			},
		},
		{
			Logged: now.Add(1 * time.Second),
			NodeID: "node2",
			Start:  now.Add(-4 * time.Minute),
			End:    now.Add(1 * time.Second),
			PhysicalTraffic: []TrafficStats{
				{Proto: 17, Src: "192.168.1.1:53", Dst: "8.8.8.8:53", RxPkts: 5, RxBytes: 500},
			},
		},
	}

	server.ResponseBody = map[string]any{"logs": expectedLogs}

	params := NetworkFlowLogsRequest{
		Start: now.Add(-1 * time.Hour),
		End:   now,
	}

	var actualLogs []NetworkFlowLog
	handler := func(log NetworkFlowLog) error {
		actualLogs = append(actualLogs, log)
		return nil
	}

	err := client.Logging().GetNetworkFlowLogs(context.Background(), params, handler)
	assert.NoError(t, err)
	assert.Equal(t, http.MethodGet, server.Method)
	assert.Equal(t, "/api/v2/tailnet/example.com/logging/network", server.Path)
	
	assert.Len(t, actualLogs, 2)
	assert.Equal(t, expectedLogs, actualLogs)
}

func TestClient_GetNetworkFlowLogs_HandlerError(t *testing.T) {
	t.Parallel()

	client, server := NewTestHarness(t)
	server.ResponseCode = http.StatusOK

	now := time.Now().UTC()
	server.ResponseBody = map[string]any{
		"logs": []NetworkFlowLog{{
			Logged: now, NodeID: "test-node", Start: now.Add(-5 * time.Minute), End: now,
		}},
	}

	params := NetworkFlowLogsRequest{Start: now.Add(-1 * time.Hour), End: now}

	handler := func(log NetworkFlowLog) error {
		return fmt.Errorf("test handler error")
	}

	err := client.Logging().GetNetworkFlowLogs(context.Background(), params, handler)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "handler error")
	assert.Contains(t, err.Error(), "test handler error")
}



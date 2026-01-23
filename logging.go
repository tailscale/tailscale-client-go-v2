// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// LoggingResource provides access to https://tailscale.com/api#tag/logging.
type LoggingResource struct {
	*Client
}

const (
	LogstreamSplunkEndpoint  LogstreamEndpointType = "splunk"
	LogstreamElasticEndpoint LogstreamEndpointType = "elastic"
	LogstreamPantherEndpoint LogstreamEndpointType = "panther"
	LogstreamCriblEndpoint   LogstreamEndpointType = "cribl"
	LogstreamDatadogEndpoint LogstreamEndpointType = "datadog"
	LogstreamAxiomEndpoint   LogstreamEndpointType = "axiom"
	LogstreamS3Endpoint      LogstreamEndpointType = "s3"
	LogstreamGCSEndpoint     LogstreamEndpointType = "gcs"
)

const (
	LogTypeConfig  LogType = "configuration"
	LogTypeNetwork LogType = "network"
)

const (
	CompressionFormatNone CompressionFormat = "none"
	CompressionFormatZstd CompressionFormat = "zstd"
	CompressionFormatGzip CompressionFormat = "gzip"
)

const (
	S3AccessKeyAuthentication S3AuthenticationType = "accesskey"
	S3RoleARNAuthentication   S3AuthenticationType = "rolearn"
)

// LogstreamConfiguration type defines a log stream entity in tailscale.
type LogstreamConfiguration struct {
	LogType              LogType               `json:"logType,omitempty"`
	DestinationType      LogstreamEndpointType `json:"destinationType,omitempty"`
	URL                  string                `json:"url,omitempty"`
	User                 string                `json:"user,omitempty"`
	UploadPeriodMinutes  int                   `json:"uploadPeriodMinutes,omitempty"`
	CompressionFormat    CompressionFormat     `json:"compressionFormat,omitempty"`
	S3Bucket             string                `json:"s3Bucket,omitempty"`
	S3Region             string                `json:"s3Region,omitempty"`
	S3KeyPrefix          string                `json:"s3KeyPrefix,omitempty"`
	S3AuthenticationType S3AuthenticationType  `json:"s3AuthenticationType,omitempty"`
	S3AccessKeyID        string                `json:"s3AccessKeyId,omitempty"`
	S3RoleARN            string                `json:"s3RoleArn,omitempty"`
	S3ExternalID         string                `json:"s3ExternalId,omitempty"`
	GCSBucket            string                `json:"gcsBucket,omitempty"`
	GCSKeyPrefix         string                `json:"gcsKeyPrefix,omitempty"`
	GCSScopes            []string              `json:"gcsScopes,omitzero"`
	GCSCredentials       string                `json:"gcsCredentials,omitempty"`
}

// SetLogstreamConfigurationRequest type defines a request for setting a LogstreamConfiguration.
type SetLogstreamConfigurationRequest struct {
	DestinationType      LogstreamEndpointType `json:"destinationType,omitempty"`
	URL                  string                `json:"url,omitempty"`
	User                 string                `json:"user,omitempty"`
	Token                string                `json:"token,omitempty"`
	UploadPeriodMinutes  int                   `json:"uploadPeriodMinutes,omitempty"`
	CompressionFormat    CompressionFormat     `json:"compressionFormat,omitempty"`
	S3Bucket             string                `json:"s3Bucket,omitempty"`
	S3Region             string                `json:"s3Region,omitempty"`
	S3KeyPrefix          string                `json:"s3KeyPrefix,omitempty"`
	S3AuthenticationType S3AuthenticationType  `json:"s3AuthenticationType,omitempty"`
	S3AccessKeyID        string                `json:"s3AccessKeyId,omitempty"`
	S3SecretAccessKey    string                `json:"s3SecretAccessKey,omitempty"`
	S3RoleARN            string                `json:"s3RoleArn,omitempty"`
	S3ExternalID         string                `json:"s3ExternalId,omitempty"`
	GCSBucket            string                `json:"gcsBucket,omitempty"`
	GCSKeyPrefix         string                `json:"gcsKeyPrefix,omitempty"`
	GCSScopes            []string              `json:"gcsScopes,omitzero"`
	GCSCredentials       string                `json:"gcsCredentials,omitempty"`
}

// LogstreamEndpointType describes the type of the endpoint.
type LogstreamEndpointType string

// LogType describes the type of logging.
type LogType string

// CompressionFormat specifies what kind of compression to use on logs.
type CompressionFormat string

// S3AuthenticationType describes the type of authentication used to stream logs to a LogstreamS3Endpoint.
type S3AuthenticationType string

// LogstreamConfiguration retrieves the tailnet's [LogstreamConfiguration] for the given [LogType].
func (lr *LoggingResource) LogstreamConfiguration(ctx context.Context, logType LogType) (*LogstreamConfiguration, error) {
	req, err := lr.buildRequest(ctx, http.MethodGet, lr.buildTailnetURL("logging", logType, "stream"))
	if err != nil {
		return nil, err
	}

	return body[LogstreamConfiguration](lr, req)
}

// SetLogstreamConfiguration sets the tailnet's [LogstreamConfiguration] for the given [LogType].
func (lr *LoggingResource) SetLogstreamConfiguration(ctx context.Context, logType LogType, request SetLogstreamConfigurationRequest) error {
	req, err := lr.buildRequest(ctx, http.MethodPut, lr.buildTailnetURL("logging", logType, "stream"), requestBody(request))
	if err != nil {
		return err
	}

	return lr.do(req, nil)
}

// DeleteLogstreamConfiguration deletes the tailnet's [LogstreamConfiguration] for the given [LogType].
func (lr *LoggingResource) DeleteLogstreamConfiguration(ctx context.Context, logType LogType) error {
	req, err := lr.buildRequest(ctx, http.MethodDelete, lr.buildTailnetURL("logging", logType, "stream"))
	if err != nil {
		return err
	}

	return lr.do(req, nil)
}

// AWSExternalID represents an AWS External ID that Tailscale can use to stream logs from a
// particular Tailscale AWS account to a LogstreamS3Endpoint that uses S3RoleARNAuthentication.
type AWSExternalID struct {
	ExternalID            string `json:"externalId,omitempty"`
	TailscaleAWSAccountID string `json:"tailscaleAwsAccountId,omitempty"`
}

// CreateOrGetAwsExternalId gets an AWS External ID that Tailscale can use to stream logs to
// a LogstreamS3Endpoint using S3RoleARNAuthentication, creating a new one for this tailnet
// when necessary.
func (lr *LoggingResource) CreateOrGetAwsExternalId(ctx context.Context, reusable bool) (*AWSExternalID, error) {
	req, err := lr.buildRequest(ctx, http.MethodPost, lr.buildTailnetURL("aws-external-id"), requestBody(map[string]bool{
		"reusable": reusable,
	}))
	if err != nil {
		return nil, err
	}
	return body[AWSExternalID](lr, req)
}

// ValidateAWSTrustPolicy validates that Tailscale can assume your AWS IAM role with (and only
// with) the given AWS External ID.
func (lr *LoggingResource) ValidateAWSTrustPolicy(ctx context.Context, awsExternalID string, roleARN string) error {
	req, err := lr.buildRequest(ctx, http.MethodPost, lr.buildTailnetURL("aws-external-id", awsExternalID, "validate-aws-trust-policy"), requestBody(map[string]string{
		"roleArn": roleARN,
	}))
	if err != nil {
		return err
	}
	return lr.do(req, nil)
}

// NetworkFlowLog represents a network flow log entry from the Tailscale API.
type NetworkFlowLog struct {
	Logged          time.Time      `json:"logged"`                    // the time at which this log was captured by the server
	NodeID          string         `json:"nodeId"`                    // the node ID for which the flow statistics apply
	Start           time.Time      `json:"start"`                     // the start of the sample period (node's local clock)
	End             time.Time      `json:"end"`                       // the end of the sample period (node's local clock)
	VirtualTraffic  []TrafficStats `json:"virtualTraffic,omitempty"`  // traffic between Tailscale nodes
	SubnetTraffic   []TrafficStats `json:"subnetTraffic,omitempty"`   // traffic involving subnet routes
	ExitTraffic     []TrafficStats `json:"exitTraffic,omitempty"`     // traffic via exit nodes
	PhysicalTraffic []TrafficStats `json:"physicalTraffic,omitempty"` // WireGuard transport-level statistics
}

// TrafficStats represents traffic flow statistics.
// This type is used for all traffic types: virtual, subnet, exit, and physical.
type TrafficStats struct {
	Proto   int    `json:"proto,omitempty"`   // IP protocol number (e.g., 6 for TCP, 17 for UDP)
	Src     string `json:"src,omitempty"`     // Source address and port
	Dst     string `json:"dst,omitempty"`     // Destination address and port
	TxPkts  uint64 `json:"txPkts,omitempty"`  // Transmitted packets
	TxBytes uint64 `json:"txBytes,omitempty"` // Transmitted bytes
	RxPkts  uint64 `json:"rxPkts,omitempty"`  // Received packets
	RxBytes uint64 `json:"rxBytes,omitempty"` // Received bytes
}

// NetworkFlowLogsRequest represents query parameters for fetching network flow logs.
type NetworkFlowLogsRequest struct {
	// Start must be set to a non-zero time within the log retention period (last 30 days).
	// The server may adjust times that are too old.
	Start time.Time
	// End must be set to a non-zero time after Start.
	End time.Time
}

// NetworkFlowLogHandler is a callback function for processing individual network flow log entries.
// It receives each log entry as it's parsed from the JSON stream.
// Return an error to stop processing and bubble up the error.
type NetworkFlowLogHandler func(log NetworkFlowLog) error

// GetNetworkFlowLogs streams network flow logs for the tailnet, calling the provided
// handler function for each log entry as it's parsed from the JSON response.
// This approach is memory-efficient and handles large datasets without loading all logs into memory.
//
// Both start and end parameters are required by the server.
// Times older than 30 days will be automatically adjusted by the server to the retention limit.
func (lr *LoggingResource) GetNetworkFlowLogs(ctx context.Context, params NetworkFlowLogsRequest, handler NetworkFlowLogHandler) error {

	u := lr.buildTailnetURL("logging", "network")
	u.RawQuery = url.Values{
		"start": {params.Start.Format(time.RFC3339)},
		"end":   {params.End.Format(time.RFC3339)},
	}.Encode()

	req, err := lr.buildRequest(ctx, http.MethodGet, u)
	if err != nil {
		return err
	}

	return lr.streamNetworkFlowLogs(req, handler)
}

// checkDelim reads and verifies the next JSON delimiter from the decoder
func checkDelim(dec *json.Decoder, want json.Delim, description string) error {
	token, err := dec.Token()
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", description, err)
	}
	if delim, ok := token.(json.Delim); !ok || delim != want {
		return fmt.Errorf("expected %c for %s, got %v", want, description, token)
	}
	return nil
}

// streamNetworkFlowLogs performs the streaming JSON parsing of network flow logs
func (lr *LoggingResource) streamNetworkFlowLogs(req *http.Request, handler NetworkFlowLogHandler) error {
	lr.init()
	resp, err := lr.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	decoder := json.NewDecoder(resp.Body)

	if err := checkDelim(decoder, '{', "opening brace"); err != nil {
		return err
	}

	token, err := decoder.Token()
	if err != nil {
		return fmt.Errorf("failed to read field name: %w", err)
	}
	if fieldName, ok := token.(string); !ok || fieldName != "logs" {
		return fmt.Errorf("expected 'logs' field, got %v", token)
	}

	if err := checkDelim(decoder, '[', "logs array start"); err != nil {
		return err
	}

	for decoder.More() {
		if err := req.Context().Err(); err != nil {
			return err
		}

		var log NetworkFlowLog
		if err := decoder.Decode(&log); err != nil {
			return fmt.Errorf("failed to decode log entry: %w", err)
		}

		if err := handler(log); err != nil {
			return fmt.Errorf("handler error: %w", err)
		}
	}

	if err := checkDelim(decoder, ']', "logs array end"); err != nil {
		return err
	}

	if err := checkDelim(decoder, '}', "closing brace"); err != nil {
		return err
	}

	return nil
}

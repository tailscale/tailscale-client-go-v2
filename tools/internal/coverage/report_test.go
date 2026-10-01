package coverage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tailscale.com/client/tailscale/v2/tools/internal/openapi"
	"tailscale.com/client/tailscale/v2/tools/internal/repoanalysis"
)

func TestBuildAndWriteMarkdown(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	specPath := filepath.Join(t.TempDir(), "spec.yaml")
	outputDir := filepath.Join(t.TempDir(), "coverage")

	const repoSource = `package tailscale

import (
	"context"
	"net/http"
	"net/url"
)

type Client struct{}

func (c *Client) buildRequest(ctx context.Context, method string, uri *url.URL, opts ...any) (*http.Request, error) {
	return nil, nil
}

func (c *Client) buildURL(pathElements ...any) *url.URL { return nil }
func (c *Client) buildTailnetURL(pathElements ...any) *url.URL { return nil }

type DevicesResource struct{ *Client }
type WebhooksResource struct{ *Client }

type Device struct {
	ID               string ` + "`json:\"id\"`" + `
	ClientVersion    string ` + "`json:\"clientVersion\"`" + `
}

type Webhook struct {
	EndpointID string ` + "`json:\"endpointId\"`" + `
	Secret     string ` + "`json:\"secret\"`" + `
}

type ExtraModel struct {
	Value string ` + "`json:\"value\"`" + `
}

func (dr *DevicesResource) Get(ctx context.Context, deviceID string) error {
	_, _ = dr.buildRequest(ctx, http.MethodGet, dr.buildURL("device", deviceID))
	return nil
}

func (wr *WebhooksResource) List(ctx context.Context) error {
	_, _ = wr.buildRequest(ctx, http.MethodGet, wr.buildTailnetURL("webhooks"))
	return nil
}
`

	const spec = `openapi: 3.1.0
paths:
  /device/{deviceId}:
    get:
      operationId: getDevice
      summary: Get a device
      tags:
        - Devices
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Device'
  /device/{deviceId}/expire:
    post:
      operationId: expireDeviceKey
      summary: Expire a device key
      tags:
        - Devices
      responses:
        '200':
          description: OK
  /tailnet/{tailnet}/webhooks:
    get:
      operationId: listWebhooks
      summary: List webhooks
      tags:
        - Webhooks
      responses:
        '200':
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Webhook'
components:
  schemas:
    Device:
      type: object
      properties:
        id:
          type: string
        clientVersion:
          type: string
        multipleConnections:
          type: boolean
    Webhook:
      type: object
      properties:
        endpointId:
          type: string
        created:
          type: string
          format: date-time
`

	if err := os.WriteFile(filepath.Join(repoRoot, "sample.go"), []byte(repoSource), 0o644); err != nil {
		t.Fatalf("WriteFile(sample.go): %v", err)
	}
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("WriteFile(spec): %v", err)
	}

	loadedSpec, err := openapi.Load(specPath)
	if err != nil {
		t.Fatalf("Load(spec): %v", err)
	}

	repo, err := repoanalysis.Analyze(repoRoot)
	if err != nil {
		t.Fatalf("Analyze(repo): %v", err)
	}

	report, err := Build(loadedSpec, repo, specPath, repoRoot)
	if err != nil {
		t.Fatalf("Build(report): %v", err)
	}

	if len(report.Endpoint.Missing) != 1 || report.Endpoint.Missing[0].OperationID != "expireDeviceKey" {
		t.Fatalf("missing operations = %#v, want expireDeviceKey", report.Endpoint.Missing)
	}

	if report.Model.TotalSpecModels != 2 {
		t.Fatalf("TotalSpecModels = %d, want 2", report.Model.TotalSpecModels)
	}
	if report.Model.CoveredModels != 2 {
		t.Fatalf("CoveredModels = %d, want 2", report.Model.CoveredModels)
	}
	if len(report.Model.ExtraModels) != 1 || report.Model.ExtraModels[0].Name != "ExtraModel" {
		t.Fatalf("ExtraModels = %#v, want ExtraModel", report.Model.ExtraModels)
	}

	var deviceMatch, webhookMatch *ModelMatch
	for i := range report.Model.Matches {
		match := &report.Model.Matches[i]
		switch match.Spec.Name {
		case "Device":
			deviceMatch = match
		case "Webhook":
			webhookMatch = match
		}
	}

	if deviceMatch == nil || webhookMatch == nil {
		t.Fatalf("matches = %#v, want Device and Webhook", report.Model.Matches)
	}
	if !strings.Contains(strings.Join(deviceMatch.MissingProperties, ","), "multipleConnections") {
		t.Fatalf("device missing properties = %#v, want multipleConnections", deviceMatch.MissingProperties)
	}
	if !strings.Contains(strings.Join(webhookMatch.MissingProperties, ","), "created") {
		t.Fatalf("webhook missing properties = %#v, want created", webhookMatch.MissingProperties)
	}
	if len(webhookMatch.ExtraProperties) != 1 || webhookMatch.ExtraProperties[0].Path != "secret" {
		t.Fatalf("webhook extra properties = %#v, want secret", webhookMatch.ExtraProperties)
	}

	if err := WriteMarkdown(report, outputDir); err != nil {
		t.Fatalf("WriteMarkdown: %v", err)
	}

	summary, err := os.ReadFile(filepath.Join(outputDir, "summary.md"))
	if err != nil {
		t.Fatalf("ReadFile(summary.md): %v", err)
	}
	if !strings.Contains(string(summary), "API models") {
		t.Fatalf("summary.md did not contain model summary: %s", string(summary))
	}

	modelCoverage, err := os.ReadFile(filepath.Join(outputDir, "model-coverage.md"))
	if err != nil {
		t.Fatalf("ReadFile(model-coverage.md): %v", err)
	}
	if !strings.Contains(string(modelCoverage), "ExtraModel") {
		t.Fatalf("model-coverage.md did not mention ExtraModel: %s", string(modelCoverage))
	}

	propertyCoverage, err := os.ReadFile(filepath.Join(outputDir, "model-property-coverage.md"))
	if err != nil {
		t.Fatalf("ReadFile(model-property-coverage.md): %v", err)
	}
	for _, want := range []string{"multipleConnections", "created", "secret"} {
		if !strings.Contains(string(propertyCoverage), want) {
			t.Fatalf("model-property-coverage.md did not mention %q: %s", want, string(propertyCoverage))
		}
	}
}

package openapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const DefaultSchemaURL = "https://api.tailscale.com/api/v2?outputOpenapiSchema=true"

// Fetch downloads the OpenAPI schema from schemaURL.
func Fetch(ctx context.Context, client *http.Client, schemaURL string) ([]byte, error) {
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, schemaURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download schema: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return nil, fmt.Errorf("download schema: unexpected status %s: %s", resp.Status, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read schema: %w", err)
	}

	return data, nil
}

// WriteFile writes the schema to path, creating parent directories when needed.
func WriteFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write schema: %w", err)
	}

	return nil
}

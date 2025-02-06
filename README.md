# tailscale.com/client/tailscale/v2

[![Go Reference](https://pkg.go.dev/badge/tailscale.com/client/tailscale/v2.svg)](https://pkg.go.dev/tailscale.com/client/tailscale/v2)
[![Github Actions](https://github.com/tailscale/tailscale-client-go-v2/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/tailscale/tailscale-client-go-v2/actions/workflows/ci.yml)

The official client implementation for the [Tailscale](https://tailscale.com) HTTP API.
For more details, please see [API documentation](https://tailscale.com/api).

## Example (Using API Key)

```go
package main

import (
	"context"
	"os"

	"tailscale.com/client/tailscale/v2"
)

func main() {
	client := &tailscale.Client{
		Tailnet: os.Getenv("TAILSCALE_TAILNET"),
		APIKey:  os.Getenv("TAILSCALE_API_KEY"),
	}

	devices, err := client.Devices().List(context.Background())
}
```

## Example (Using OAuth)

```go
package main

import (
	"context"
	"os"

	"tailscale.com/client/tailscale/v2"
)

func main() {
	client := &tailscale.Client{
		Tailnet: os.Getenv("TAILSCALE_TAILNET"),
		HTTP:    tailscale.OAuthConfig{
			ClientID:     os.Getenv("TAILSCALE_OAUTH_CLIENT_ID"),
			ClientSecret: os.Getenv("TAILSCALE_OAUTH_CLIENT_SECRET"),
			Scopes:       []string{"all:write"},
		}.HTTPClient(),
	}
	
	devices, err := client.Devices().List(context.Background())
}
```

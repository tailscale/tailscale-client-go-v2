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

## Example (Using Identity Federation)

### With a static ID token:

For static ID tokens, simply return the same token value each time. Note that if both the Tailscale API access token
and the ID token expire, the client must be recreated with a fresh ID token to reauthenticate.

```go
package main

import (
	"context"
	"log"
	"os"

	"tailscale.com/client/tailscale/v2"
)

func main() {
	httpClient, err := tailscale.IdentityFederationConfig{
		ClientID: os.Getenv("TAILSCALE_CLIENT_ID"),
		IDTokenFunc: func() (string, error) {
			return os.Getenv("ID_TOKEN"), nil
		},
	}.HTTPClient()
	if err != nil {
		log.Fatal(err)
	}

	client := &tailscale.Client{
		Tailnet: os.Getenv("TAILSCALE_TAILNET"),
		HTTP:    httpClient,
	}

	devices, err := client.Devices().List(context.Background())
}
```

### With a dynamic ID token generator

For long-running applications, instruct the client on how to fetch ID tokens from your IdP so the client can reauthenticate
automatically when the Tailscale API access token and ID token expire:

```go
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"tailscale.com/client/tailscale/v2"
)

func main() {
	httpClient, err := tailscale.IdentityFederationConfig{
		ClientID: os.Getenv("TAILSCALE_CLIENT_ID"),
		IDTokenFunc: func() (string, error) {
			resp, err := http.Get("https://my-idp.com/id-token")
			if err != nil {
				return "", fmt.Errorf("failed to fetch ID token: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return "", fmt.Errorf("failed to fetch ID token: status %d", resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", fmt.Errorf("failed to read ID token response: %w", err)
			}

			return strings.TrimSpace(string(body)), nil
		},
	}.HTTPClient()
	if err != nil {
		log.Fatal(err)
	}

	client := &tailscale.Client{
		Tailnet: os.Getenv("TAILSCALE_TAILNET"),
		HTTP:    httpClient,
	}

	devices, err := client.Devices().List(context.Background())
}
```

## Releasing

Pushing a tag of the format `vX.Y.Z` will trigger the [release workflow](./.github/workflows/release.yml) which uses
[goreleaser](https://github.com/goreleaser/goreleaser) to build and sign artifacts and generate a
[GitHub release](https://github.com/tailscale/tailscale-client-go-v2/releases).


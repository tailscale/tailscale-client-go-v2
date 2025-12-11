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
		Auth: &tailscale.OAuth{
			ClientID:     os.Getenv("TAILSCALE_OAUTH_CLIENT_ID"),
			ClientSecret: os.Getenv("TAILSCALE_OAUTH_CLIENT_SECRET"),
			Scopes:       []string{"all:write"},
		},
	}
	
	devices, err := client.Devices().List(context.Background())
}
```

## Example (Using Identity Federation)

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
		Auth: &tailscale.IdentityFederation{
			ClientID: os.Getenv("TAILSCALE_OAUTH_CLIENT_ID"),
			IDTokenFunc: func() (string, error) {
				return os.Getenv("IDENTITY_TOKEN"), nil
            },
		},
	}

	devices, err := client.Devices().List(context.Background())
}
```

## Example (Using Your Own Authentication Mechanism)

```go
package main

import (
	"context"
	"os"

	"tailscale.com/client/tailscale/v2"
)

type MyAuth struct {...}

func (a *MyAuth) HTTPClient(orig *http.Client, baseURL string) *http.Client {
	// build an HTTP client that adds authentication to outgoing requests
	// see tailscale.OAuth for an example.
}

func main() {
	client := &tailscale.Client{
		Tailnet: os.Getenv("TAILSCALE_TAILNET"),
		Auth: &MyAuth{...},
	}
	
	devices, err := client.Devices().List(context.Background())
}
```

## Releasing

Pushing a tag of the format `vX.Y.Z` will trigger the [release workflow](./.github/workflows/release.yml) which uses
[goreleaser](https://github.com/goreleaser/goreleaser) to build and sign artifacts and generate a
[GitHub release](https://github.com/tailscale/tailscale-client-go-v2/releases).


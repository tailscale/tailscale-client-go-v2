package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"tailscale.com/client/tailscale/v2/tools/internal/openapi"
)

func main() {
	var (
		schemaURL = flag.String("url", openapi.DefaultSchemaURL, "URL to fetch the OpenAPI schema from")
		output    = flag.String("out", "tools/openapi/spec/tailscale-v2-openapi.yaml", "path to write the downloaded schema to")
		timeout   = flag.Duration("timeout", 30*time.Second, "HTTP timeout")
	)
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	data, err := openapi.Fetch(ctx, &http.Client{Timeout: *timeout}, *schemaURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := openapi.WriteFile(*output, data); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("wrote %s\n", *output)
}

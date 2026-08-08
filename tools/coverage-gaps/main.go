package main

import (
	"flag"
	"fmt"
	"os"

	"tailscale.com/client/tailscale/v2/tools/internal/coverage"
	"tailscale.com/client/tailscale/v2/tools/internal/openapi"
	"tailscale.com/client/tailscale/v2/tools/internal/repoanalysis"
)

func main() {
	var (
		specPath = flag.String("spec", "tools/openapi/spec/tailscale-v2-openapi.yaml", "path to the downloaded OpenAPI schema")
		repoRoot = flag.String("repo", ".", "path to the repository root to inspect")
		output   = flag.String("out", "docs/coverage-gaps", "directory to write markdown reports to")
	)
	flag.Parse()

	spec, err := openapi.Load(*specPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	repo, err := repoanalysis.Analyze(*repoRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	report, err := coverage.Build(spec, repo, *specPath, *repoRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := coverage.WriteMarkdown(report, *output); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("wrote coverage reports to %s\n", *output)
}

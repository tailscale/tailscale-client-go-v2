OPENAPI_SPEC ?= tools/openapi/spec/tailscale-v2-openapi.yaml
COVERAGE_GAPS_DIR ?= docs/coverage-gaps
GOCACHE ?= $(CURDIR)/.gocache

export GOCACHE

.PHONY: openapi-refresh coverage-gaps

openapi-refresh:
	go run ./tools/fetch-openapi -out $(OPENAPI_SPEC)

coverage-gaps:
	go run ./tools/coverage-gaps -spec $(OPENAPI_SPEC) -out $(COVERAGE_GAPS_DIR)

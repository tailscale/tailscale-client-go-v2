// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"net/http"
	"time"
)

// TailnetsResource provides access to tailnet creation APIs.
type TailnetsResource struct {
	*Client
}

// CreateTailnetRequest describes the definition of an API-only tailnet to create.
type CreateTailnetRequest struct {
	DisplayName string `json:"displayName"`
}

// Tailnet describes a tailnet in an organization.
type Tailnet struct {
	ID          string              `json:"id"`
	DisplayName string              `json:"displayName"`
	OrgID       string              `json:"orgId"`
	DNSName     string              `json:"dnsName"`
	CreatedAt   time.Time           `json:"createdAt"`
	OAuthClient *TailnetOAuthClient `json:"oauthClient,omitempty"`
}

// TailnetOAuthClient describes the OAuth client returned when creating an API-only tailnet.
type TailnetOAuthClient struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

// Create creates a new API-only tailnet. Returns the created [Tailnet] if successful.
func (tr *TailnetsResource) Create(ctx context.Context, request CreateTailnetRequest) (*Tailnet, error) {
	req, err := tr.buildRequest(ctx, http.MethodPost, tr.buildURL("organizations", "-", "tailnets"), requestBody(request))
	if err != nil {
		return nil, err
	}

	return body[Tailnet](tr, req)
}

// List lists every tailnet in the organization.
func (tr *TailnetsResource) List(ctx context.Context) ([]Tailnet, error) {
	req, err := tr.buildRequest(ctx, http.MethodGet, tr.buildURL("organizations", "-", "tailnets"))
	if err != nil {
		return nil, err
	}

	resp := make(map[string][]Tailnet)
	if err = tr.do(req, &resp); err != nil {
		return nil, err
	}

	return resp["tailnets"], nil
}

// Delete deletes the tailnet associated with the current client.
func (tr *TailnetsResource) Delete(ctx context.Context) error {
	req, err := tr.buildRequest(ctx, http.MethodDelete, tr.buildTailnetURL())
	if err != nil {
		return err
	}

	return tr.do(req, nil)
}

// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"net/http"
	"time"
)

// KeysResource provides access to https://tailscale.com/api#tag/keys.
type KeysResource struct {
	*Client
}

// KeyCapabilities describes the capabilities of an authentication key.
type KeyCapabilities struct {
	Devices struct {
		Create struct {
			Reusable      bool     `json:"reusable"`
			Ephemeral     bool     `json:"ephemeral"`
			Tags          []string `json:"tags"`
			Preauthorized bool     `json:"preauthorized"`
		} `json:"create"`
	} `json:"devices"`
}

// CreateKeyRequest describes the definition of an authentication key to create.
type CreateKeyRequest struct {
	Capabilities  KeyCapabilities `json:"capabilities"`
	ExpirySeconds int64           `json:"expirySeconds"`
	Description   string          `json:"description"`
}

// CreateOAuthClientRequest describes the definition of an OAuth client to create.
type CreateOAuthClientRequest struct {
	Scopes      []string `json:"scopes"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
}

type createOAuthClientWithKeyTypeRequest struct {
	KeyType string `json:"keyType"`
	CreateOAuthClientRequest
}

// Key describes an authentication key within the tailnet.
type Key struct {
	ID            string          `json:"id"`
	KeyType       string          `json:"keyType"`
	Key           string          `json:"key"`
	Description   string          `json:"description"`
	ExpirySeconds *time.Duration  `json:"expirySeconds"`
	Created       time.Time       `json:"created"`
	Expires       time.Time       `json:"expires"`
	Revoked       time.Time       `json:"revoked"`
	Invalid       bool            `json:"invalid"`
	Capabilities  KeyCapabilities `json:"capabilities"`
	Scopes        []string        `json:"scopes,omitempty"`
	Tags          []string        `json:"tags,omitempty"`
	UserID        string          `json:"userId"`
}

// Create creates a new authentication key. Returns the generated [Key] if successful.
// Deprecated: Use CreateAuthKey instead.
func (kr *KeysResource) Create(ctx context.Context, ckr CreateKeyRequest) (*Key, error) {
	req, err := kr.buildRequest(ctx, http.MethodPost, kr.buildTailnetURL("keys"), requestBody(ckr))
	if err != nil {
		return nil, err
	}

	return body[Key](kr, req)
}

// CreateAuthKey creates a new authentication key. Returns the generated [Key] if successful.
func (kr *KeysResource) CreateAuthKey(ctx context.Context, ckr CreateKeyRequest) (*Key, error) {
	return kr.Create(ctx, ckr)
}

// CreateOAuthClient creates a new OAuth client. Returns the generated [Key] if successful.
func (kr *KeysResource) CreateOAuthClient(ctx context.Context, ckr CreateOAuthClientRequest) (*Key, error) {
	req, err := kr.buildRequest(ctx, http.MethodPost, kr.buildTailnetURL("keys"), requestBody(createOAuthClientWithKeyTypeRequest{
		KeyType:                  "client",
		CreateOAuthClientRequest: ckr,
	}))
	if err != nil {
		return nil, err
	}

	return body[Key](kr, req)
}

// Get returns all information on a [Key] whose identifier matches the one provided. This will not return the
// authentication key itself, just the metadata.
func (kr *KeysResource) Get(ctx context.Context, id string) (*Key, error) {
	req, err := kr.buildRequest(ctx, http.MethodGet, kr.buildTailnetURL("keys", id))
	if err != nil {
		return nil, err
	}

	return body[Key](kr, req)
}

// List returns every [Key] within the tailnet. The only fields set for each [Key] will be its identifier.
// The keys returned are relative to the user that owns the API key used to authenticate the client.
//
// Specify all to list both user and tailnet level keys.
func (kr *KeysResource) List(ctx context.Context, all bool) ([]Key, error) {
	url := kr.buildTailnetURL("keys")
	if all {
		url.RawQuery = "all=true"
	}
	req, err := kr.buildRequest(ctx, http.MethodGet, url)
	if err != nil {
		return nil, err
	}

	resp := make(map[string][]Key)
	if err = kr.do(req, &resp); err != nil {
		return nil, err
	}

	return resp["keys"], nil
}

// Delete removes an authentication key from the tailnet.
func (kr *KeysResource) Delete(ctx context.Context, id string) error {
	req, err := kr.buildRequest(ctx, http.MethodDelete, kr.buildTailnetURL("keys", id))
	if err != nil {
		return err
	}

	return kr.do(req, nil)
}

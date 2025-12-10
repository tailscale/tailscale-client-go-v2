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

// SetOAuthClientRequest describes the definition of an existing OAuth client to
// set (wholesale update) the configuration of.
type SetOAuthClientRequest struct {
	Scopes      []string `json:"scopes"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
}

type setOAuthClientWithKeyTypeRequest struct {
	KeyType string `json:"keyType"`
	SetOAuthClientRequest
}

// CreateFederatedIdentityRequest describes the definition of a federated identity to create.
type CreateFederatedIdentityRequest struct {
	Scopes           []string          `json:"scopes"`
	Tags             []string          `json:"tags"`
	Audience         string            `json:"audience"`
	Issuer           string            `json:"issuer"`
	Subject          string            `json:"subject"`
	CustomClaimRules map[string]string `json:"customClaimRules"`
	Description      string            `json:"description"`
}

type createFederatedIdentityWithKeyTypeRequest struct {
	KeyType string `json:"keyType"`
	CreateFederatedIdentityRequest
}

// SetFederatedIdentityRequest describes the definition of a federated identity to create.
type SetFederatedIdentityRequest struct {
	Scopes           []string          `json:"scopes"`
	Tags             []string          `json:"tags"`
	Audience         string            `json:"audience"`
	Issuer           string            `json:"issuer"`
	Subject          string            `json:"subject"`
	CustomClaimRules map[string]string `json:"customClaimRules"`
	Description      string            `json:"description"`
}

type setFederatedIdentityWithKeyTypeRequest struct {
	KeyType string `json:"keyType"`
	SetFederatedIdentityRequest
}

// Key describes an authentication key within the tailnet.
type Key struct {
	ID               string            `json:"id"`
	KeyType          string            `json:"keyType"`
	Key              string            `json:"key"`
	Description      string            `json:"description"`
	ExpirySeconds    *time.Duration    `json:"expirySeconds"`
	Created          time.Time         `json:"created"`
	Updated          time.Time         `json:"updated"`
	Expires          time.Time         `json:"expires"`
	Revoked          time.Time         `json:"revoked"`
	Invalid          bool              `json:"invalid"`
	Capabilities     KeyCapabilities   `json:"capabilities"`
	Scopes           []string          `json:"scopes,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	UserID           string            `json:"userId"`
	Audience         string            `json:"audience"`
	Issuer           string            `json:"issuer"`
	Subject          string            `json:"subject"`
	CustomClaimRules map[string]string `json:"customClaimRules"`
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

// SetOAuthClient sets the configuration for an existing OAuth client. Returns the generated [Key] if successful.
func (kr *KeysResource) SetOAuthClient(ctx context.Context, id string, skr SetOAuthClientRequest) (*Key, error) {
	req, err := kr.buildRequest(ctx, http.MethodPut, kr.buildTailnetURL("keys", id), requestBody(setOAuthClientWithKeyTypeRequest{
		KeyType:               "client",
		SetOAuthClientRequest: skr,
	}))
	if err != nil {
		return nil, err
	}

	return body[Key](kr, req)
}

// CreateFederatedIdentity creates a new federated identity. Returns the generated [Key] if successful.
func (kr *KeysResource) CreateFederatedIdentity(ctx context.Context, ckr CreateFederatedIdentityRequest) (*Key, error) {
	req, err := kr.buildRequest(ctx, http.MethodPost, kr.buildTailnetURL("keys"), requestBody(createFederatedIdentityWithKeyTypeRequest{
		KeyType:                        "federated",
		CreateFederatedIdentityRequest: ckr,
	}))
	if err != nil {
		return nil, err
	}

	return body[Key](kr, req)
}

// SetFederatedIdentity sets the configuration for an existing federated identity. Returns the generated [Key] if successful.
func (kr *KeysResource) SetFederatedIdentity(ctx context.Context, id string, skr SetFederatedIdentityRequest) (*Key, error) {
	req, err := kr.buildRequest(ctx, http.MethodPut, kr.buildTailnetURL("keys", id), requestBody(setFederatedIdentityWithKeyTypeRequest{
		KeyType:                     "federated",
		SetFederatedIdentityRequest: skr,
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

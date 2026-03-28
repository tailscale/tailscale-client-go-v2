// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"net/http"
)

// ServicesResource provides access to https://tailscale.com/api#tag/services.
type ServicesResource struct {
	*Client
}

// Service is a Tailscale service with a stable virtual IP address.
type Service struct {
	Name        string            `json:"name,omitempty"`
	Addrs       []string          `json:"addrs,omitempty"`
	Comment     string            `json:"comment,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Ports       []string          `json:"ports,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

type serviceList struct {
	Services []Service `json:"vipServices"`
}

// List lists every [Service] in the tailnet.
func (sr *ServicesResource) List(ctx context.Context) ([]Service, error) {
	req, err := sr.buildRequest(ctx, http.MethodGet, sr.buildTailnetURL("vip-services"))
	if err != nil {
		return nil, err
	}

	resp, err := body[serviceList](sr, req)
	if err != nil {
		return nil, err
	}
	return resp.Services, nil
}

// Get retrieves a specific [Service] by name.
func (sr *ServicesResource) Get(ctx context.Context, name string) (*Service, error) {
	req, err := sr.buildRequest(ctx, http.MethodGet, sr.buildTailnetURL("vip-services", name))
	if err != nil {
		return nil, err
	}

	return body[Service](sr, req)
}

// CreateOrUpdate creates or updates a [Service].
func (sr *ServicesResource) CreateOrUpdate(ctx context.Context, svc Service) error {
	req, err := sr.buildRequest(ctx, http.MethodPut, sr.buildTailnetURL("vip-services", svc.Name), requestBody(svc))
	if err != nil {
		return err
	}

	return sr.do(req, nil)
}

// Delete deletes a specific [Service].
func (sr *ServicesResource) Delete(ctx context.Context, name string) error {
	req, err := sr.buildRequest(ctx, http.MethodDelete, sr.buildTailnetURL("vip-services", name))
	if err != nil {
		return err
	}

	return sr.do(req, nil)
}

// VIPService is an alias for [Service].
// Deprecated: use [Service] instead.
type VIPService = Service

// VIPServicesResource is an alias for [ServicesResource].
// Deprecated: use [ServicesResource] instead.
type VIPServicesResource = ServicesResource

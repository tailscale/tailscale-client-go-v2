// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"net/http"
)

// VIPServicesResource provides access to https://tailscale.com/api#tag/vipservices.
type VIPServicesResource struct {
	*Client
}

// VIPService is a Tailscale VIP service with a stable virtual IP address.
type VIPService struct {
	Name        string            `json:"name,omitempty"`
	Addrs       []string          `json:"addrs,omitempty"`
	Comment     string            `json:"comment,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Ports       []string          `json:"ports,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

type vipServiceList struct {
	VIPServices []VIPService `json:"vipServices"`
}

// List lists every [VIPService] in the tailnet.
func (vr *VIPServicesResource) List(ctx context.Context) ([]VIPService, error) {
	req, err := vr.buildRequest(ctx, http.MethodGet, vr.buildTailnetURL("vip-services"))
	if err != nil {
		return nil, err
	}

	resp, err := body[vipServiceList](vr, req)
	if err != nil {
		return nil, err
	}
	return resp.VIPServices, nil
}

// Get retrieves a specific [VIPService] by name.
func (vr *VIPServicesResource) Get(ctx context.Context, name string) (*VIPService, error) {
	req, err := vr.buildRequest(ctx, http.MethodGet, vr.buildTailnetURL("vip-services", name))
	if err != nil {
		return nil, err
	}

	return body[VIPService](vr, req)
}

// CreateOrUpdate creates or updates a [VIPService].
func (vr *VIPServicesResource) CreateOrUpdate(ctx context.Context, svc VIPService) error {
	req, err := vr.buildRequest(ctx, http.MethodPut, vr.buildTailnetURL("vip-services", svc.Name), requestBody(svc))
	if err != nil {
		return err
	}

	return vr.do(req, nil)
}

// Delete deletes a specific [VIPService].
func (vr *VIPServicesResource) Delete(ctx context.Context, name string) error {
	req, err := vr.buildRequest(ctx, http.MethodDelete, vr.buildTailnetURL("vip-services", name))
	if err != nil {
		return err
	}

	return vr.do(req, nil)
}

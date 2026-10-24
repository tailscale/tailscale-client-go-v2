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
	DisplayName string            `json:"displayName,omitempty"`
	Addrs       []string          `json:"addrs,omitempty"`
	Comment     string            `json:"comment,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Ports       []string          `json:"ports,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

type serviceList struct {
	Services []Service `json:"vipServices"`
}

// ServiceHostInfo is an information summary for a device hosting a [Service].
type ServiceHostInfo struct {
	// note the open-api spec says the key is "stableNodeID", but the actual api returns "nodeId"
	StableNodeID  string `json:"nodeId,omitempty"`
	ApprovalLevel string `json:"approvalLevel,omitempty"`
	Configured    string `json:"configured,omitempty"`
}

type serviceHostList struct {
	Hosts []ServiceHostInfo `json:"hosts"`
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

// ServiceDevices lists all devices hosting the specified [Service].
func (sr *ServicesResource) ServiceDevices(ctx context.Context, name string) ([]ServiceHostInfo, error) {
	req, err := sr.buildRequest(ctx, http.MethodGet, sr.buildTailnetURL("services", name, "devices"))
	if err != nil {
		return nil, err
	}
	resp, err := body[serviceHostList](sr, req)
	if err != nil {
		return nil, err
	}
	return resp.Hosts, nil
}

// ServiceApproval is the approval status of a [Service] on a specific device.
type ServiceApproval struct {
	Approved     bool `json:"approved"`
	AutoApproved bool `json:"autoApproved,omitempty"`
}

type setServiceApprovalRequest struct {
	Approved bool `json:"approved"`
}

// SetDeviceApproval sets the approval status of the named [Service] on the specified device.
func (sr *ServicesResource) SetDeviceApproval(ctx context.Context, name, deviceID string, approved bool) (*ServiceApproval, error) {
	req, err := sr.buildRequest(ctx, http.MethodPost, sr.buildTailnetURL("services", name, "device", deviceID, "approved"), requestBody(setServiceApprovalRequest{Approved: approved}))
	if err != nil {
		return nil, err
	}
	return body[ServiceApproval](sr, req)
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

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

type Service struct {
	Name    string   `json:"name"`
	Addrs   []string `json:"addrs"`
	Comment string   `json:"comment"`
	Ports   []string `json:"ports"`
	Tags    []string `json:"tags"`
}

type serviceUpdateOp func(Service) (svcName string, service Service)

// WithServiceName is a functional option for [ServicesResource]'s Update method that allows specifying
// the old name of the service that is to be renamed.
func WithServiceName(name string) serviceUpdateOp {
	return func(svc Service) (string, Service) {
		return name, svc
	}
}

// Update creates or updates a given [Service]. To rename a service, pass the [WithServiceName] option.
func (sr *ServicesResource) Update(ctx context.Context, service Service, opts ...serviceUpdateOp) (*Service, error) {
	name := service.Name
	for _, opt := range opts {
		name, service = opt(service)
	}
	req, err := sr.buildRequest(ctx, http.MethodPut, sr.buildTailnetURL("services", name), requestBody(service))
	if err != nil {
		return nil, err
	}

	return body[Service](sr, req)
}

// List lists every [Service] in the tailnet.
func (sr *ServicesResource) List(ctx context.Context) ([]Service, error) {
	u := sr.buildTailnetURL("services")
	req, err := sr.buildRequest(ctx, http.MethodGet, u)
	if err != nil {
		return nil, err
	}

	resp := make(map[string][]Service)
	if err = sr.do(req, &resp); err != nil {
		return nil, err
	}
	return resp["vipServices"], nil
}

// Get retrieves a specific [Service].
func (sr *ServicesResource) Get(ctx context.Context, name string) (*Service, error) {
	req, err := sr.buildRequest(ctx, http.MethodGet, sr.buildTailnetURL("services", name))
	if err != nil {
		return nil, err
	}
	return body[Service](sr, req)
}

// Delete deletes a specific service.
func (sr *ServicesResource) Delete(ctx context.Context, name string) error {
	req, err := sr.buildRequest(ctx, http.MethodDelete, sr.buildTailnetURL("services", name))
	if err != nil {
		return err
	}

	return sr.do(req, nil)
}

// ServiceDeviceApprovalStatus defines the approval level of a device that desires to serve a [Service].
type ServiceDeviceApprovalStatus struct {
	Approved     bool `json:"approved"`
	AutoApproved bool `json:"autoApproved"`
}

// GetDeviceApprovalStatus returns the approval status indicating whether the given device may provide the given service.
func (sr *ServicesResource) GetDeviceApprovalStatus(ctx context.Context, service string, device string) (*ServiceDeviceApprovalStatus, error) {
	req, err := sr.buildRequest(ctx, http.MethodGet, sr.buildTailnetURL("services", service, "device", device))
	if err != nil {
		return nil, err
	}

	return body[ServiceDeviceApprovalStatus](sr, req)
}

type deviceApprovalStatusRequest struct {
	Approved bool `json:"approved"`
}

// UpdateDeviceApprovalStatus sets whether a given device is approved to serve the given service.
func (sr *ServicesResource) UpdateDeviceApprovalStatus(ctx context.Context, service string, device string, approved bool) (*ServiceDeviceApprovalStatus, error) {
	req, err := sr.buildRequest(ctx, http.MethodPost, sr.buildTailnetURL("services", service, "device", device, "approved"), requestBody(deviceApprovalStatusRequest{Approved: approved}))
	if err != nil {
		return nil, err
	}

	return body[ServiceDeviceApprovalStatus](sr, req)
}

// ServiceDeviceApprovalLevel defines whether (and how) a tailnet device is allowed to serve a service.
type ServiceDeviceApprovalLevel string

const (
	ServiceDeviceNotApproved      ServiceDeviceApprovalLevel = "not-approved"
	ServiceDeviceAutoApproved     ServiceDeviceApprovalLevel = "approved:auto"
	ServiceDeviceManuallyApproved ServiceDeviceApprovalLevel = "approved:manual"
)

// IsApproved returns true if and only if the device is approved for serving a service.
func (dsal ServiceDeviceApprovalLevel) IsApproved() bool {
	return dsal != ServiceDeviceNotApproved
}

// ServiceDeviceStatus indicates the status of a tailnet node that is serving a service.
type ServiceDeviceStatus struct {
	NodeID        string                     `json:"stableNodeID"`
	ApprovalLevel ServiceDeviceApprovalLevel `json:"approvalLevel"`
	Configured    string                     `json:"configured"`
}

// ListDevices returns a list of tailnet devices currently serving a service, with their approval status.
func (sr *ServicesResource) ListDevices(ctx context.Context, service string) ([]ServiceDeviceStatus, error) {
	req, err := sr.buildRequest(ctx, http.MethodGet, sr.buildTailnetURL("services", service, "devices"))
	if err != nil {
		return nil, err
	}

	resp := make(map[string][]ServiceDeviceStatus)
	if err = sr.do(req, &resp); err != nil {
		return nil, err
	}
	return resp["hosts"], nil
}

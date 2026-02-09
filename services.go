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

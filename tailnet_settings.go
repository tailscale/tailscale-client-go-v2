// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"net/http"
)

// TailnetSettingsResource provides access to https://tailscale.com/api#tag/tailnetsettings.
type TailnetSettingsResource struct {
	*Client
}

// TailnetSettings represents the current settings of a tailnet.
// See https://tailscale.com/api#model/tailnetsettings.
type TailnetSettings struct {
	ACLsExternallyManagedOn bool   `json:"aclsExternallyManagedOn"`
	ACLsExternalLink        string `json:"aclsExternalLink"`

	DevicesApprovalOn      bool `json:"devicesApprovalOn"`
	DevicesAutoUpdatesOn   bool `json:"devicesAutoUpdatesOn"`
	DevicesKeyDurationDays int  `json:"devicesKeyDurationDays"` // days before device key expiry

	UsersApprovalOn                        bool                              `json:"usersApprovalOn"`
	UsersRoleAllowedToJoinExternalTailnets RoleAllowedToJoinExternalTailnets `json:"usersRoleAllowedToJoinExternalTailnets"`

	NetworkFlowLoggingOn bool `json:"networkFlowLoggingOn"`

	// RegionalRoutingOn conveys whether a tailnet's [RouteSelection] is set to [RegionalRouting].
	//
	// Deprecated: This field has been superseded by [TailnetSettings.RouteSelection]
	// as of 2026-10-09, but will be returned in GET responses for backwards compatibility.
	RegionalRoutingOn bool `json:"regionalRoutingOn"`

	PostureIdentityCollectionOn bool           `json:"postureIdentityCollectionOn"`
	HTTPSEnabled                bool           `json:"httpsEnabled"`
	RouteSelection              RouteSelection `json:"routeSelection"`
}

// UpdateTailnetSettingsRequest is a request to update the settings of a tailnet.
// Nil values indicate that the existing setting should be left unchanged.
type UpdateTailnetSettingsRequest struct {
	ACLsExternallyManagedOn *bool   `json:"aclsExternallyManagedOn"`
	ACLsExternalLink        *string `json:"aclsExternalLink"`

	DevicesApprovalOn      *bool `json:"devicesApprovalOn,omitempty"`
	DevicesAutoUpdatesOn   *bool `json:"devicesAutoUpdatesOn,omitempty"`
	DevicesKeyDurationDays *int  `json:"devicesKeyDurationDays,omitempty"` // days before device key expiry

	UsersApprovalOn                        *bool                              `json:"usersApprovalOn,omitempty"`
	UsersRoleAllowedToJoinExternalTailnets *RoleAllowedToJoinExternalTailnets `json:"usersRoleAllowedToJoinExternalTailnets,omitempty"`

	NetworkFlowLoggingOn *bool `json:"networkFlowLoggingOn,omitempty"`

	// RegionalRoutingOn is whether a tailnet's [RouteSelection] should be set to [RouteSelectionRegionalRouting].
	//
	// Deprecated: This field has been superseded by [UpdateTailnetSettingsRequest.RouteSelection]
	// as of 2026-10-09, but can still be updated. Requests must not specify both the
	// [UpdateTailnetSettingsRequest.RegionalRoutingOn] and [UpdateTailnetSettingsRequest.RouteSelection] fields.
	RegionalRoutingOn *bool `json:"regionalRoutingOn,omitempty"`

	PostureIdentityCollectionOn *bool `json:"postureIdentityCollectionOn,omitempty"`
	HTTPSEnabled                *bool `json:"httpsEnabled,omitempty"`

	// RouteSelection is the [RouteSelection] to set for a tailnet.
	// Requests must not specify both the [UpdateTailnetSettingsRequest.RegionalRoutingOn] and
	// [UpdateTailnetSettingsRequest.RouteSelection] fields.
	RouteSelection *RouteSelection `json:"routeSelection,omitempty"`
}

// RoleAllowedToJoinExternalTailnets constrains which users are allowed to join external tailnets
// based on their role.
type RoleAllowedToJoinExternalTailnets string

const (
	RoleAllowedToJoinExternalTailnetsNone   RoleAllowedToJoinExternalTailnets = "none"
	RoleAllowedToJoinExternalTailnetsAdmin  RoleAllowedToJoinExternalTailnets = "admin"
	RoleAllowedToJoinExternalTailnetsMember RoleAllowedToJoinExternalTailnets = "member"
)

// RouteSelection is the route selection algorithm for a tailnet.
// See https://tailscale.com/docs/features/route-selection.
type RouteSelection string

const (
	RouteSelectionActivePassiveFailover           RouteSelection = "active-passive-failover"
	RouteSelectionRegionalRouting                 RouteSelection = "regional-routing"
	RouteSelectionRegionalRoutingInRegionFailover RouteSelection = "regional-routing-failover"
	RouteSelectionMagicRoute                      RouteSelection = "magicroute"
)

// Get retrieves the current [TailnetSettings].
// See https://tailscale.com/api#tag/tailnetsettings/GET/tailnet/{tailnet}/settings.
func (tsr *TailnetSettingsResource) Get(ctx context.Context) (*TailnetSettings, error) {
	req, err := tsr.buildRequest(ctx, http.MethodGet, tsr.buildTailnetURL("settings"))
	if err != nil {
		return nil, err
	}

	return body[TailnetSettings](tsr, req)
}

// Update updates the tailnet settings.
// See https://tailscale.com/api#tag/tailnetsettings/PATCH/tailnet/{tailnet}/settings.
func (tsr *TailnetSettingsResource) Update(ctx context.Context, request UpdateTailnetSettingsRequest) error {
	req, err := tsr.buildRequest(ctx, http.MethodPatch, tsr.buildTailnetURL("settings"), requestBody(request))
	if err != nil {
		return err
	}

	return tsr.do(req, nil)
}

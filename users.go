// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"net/http"
	"time"
)

// UsersResource provides access to https://tailscale.com/api#tag/users.
type UsersResource struct {
	*Client
}

const (
	UserTypeMember UserType = "member"
	UserTypeShared UserType = "shared"
)

const (
	UserRoleOwner        UserRole = "owner"
	UserRoleMember       UserRole = "member"
	UserRoleAdmin        UserRole = "admin"
	UserRoleITAdmin      UserRole = "it-admin"
	UserRoleNetworkAdmin UserRole = "network-admin"
	UserRoleBillingAdmin UserRole = "billing-admin"
	UserRoleAuditor      UserRole = "auditor"
)

const (
	UserStatusActive           UserStatus = "active"
	UserStatusIdle             UserStatus = "idle"
	UserStatusSuspended        UserStatus = "suspended"
	UserStatusNeedsApproval    UserStatus = "needs-approval"
	UserStatusOverBillingLimit UserStatus = "over-billing-limit"
)

// UserType is the type of relation this user has to the tailnet associated with the request.
type UserType string

// UserRole is the role of the user.
type UserRole string

// UserStatus is the status of the user.
type UserStatus string

// User is a representation of a user within a tailnet.
type User struct {
	ID                 string     `json:"id"`
	DisplayName        string     `json:"displayName"`
	LoginName          string     `json:"loginName"`
	ProfilePicURL      string     `json:"profilePicUrl"`
	TailnetID          string     `json:"tailnetId"`
	Created            time.Time  `json:"created"`
	Type               UserType   `json:"type"`
	Role               UserRole   `json:"role"`
	Status             UserStatus `json:"status"`
	DeviceCount        int        `json:"deviceCount"`
	LastSeen           time.Time  `json:"lastSeen"`
	CurrentlyConnected bool       `json:"currentlyConnected"`
}

// List lists every [User] of the tailnet. If userType and/or role are provided,
// the list of users will be filtered by those.
func (ur *UsersResource) List(ctx context.Context, userType *UserType, role *UserRole) ([]User, error) {
	u := ur.buildTailnetURL("users")
	q := u.Query()
	if userType != nil {
		q.Add("type", string(*userType))
	}
	if role != nil {
		q.Add("role", string(*role))
	}
	u.RawQuery = q.Encode()

	req, err := ur.buildRequest(ctx, http.MethodGet, u)
	if err != nil {
		return nil, err
	}

	resp := make(map[string][]User)
	if err = ur.do(req, &resp); err != nil {
		return nil, err
	}

	return resp["users"], nil
}

// Get retrieves the [User] identified by the given id.
func (ur *UsersResource) Get(ctx context.Context, id string) (*User, error) {
	req, err := ur.buildRequest(ctx, http.MethodGet, ur.buildURL("users", id))
	if err != nil {
		return nil, err
	}

	return body[User](ur, req)
}

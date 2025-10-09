// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

package tailscale

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// DevicesResource provides access to https://tailscale.com/api#tag/devices.
type DevicesResource struct {
	*Client
}

type DeviceRoutes struct {
	Advertised []string `json:"advertisedRoutes"`
	Enabled    []string `json:"enabledRoutes"`
}

// Time wraps a time and allows for unmarshalling timestamps that represent an empty time as an empty string (e.g "")
// this is used by the tailscale API when it returns devices that have no created date, such as its hello service.
type Time struct {
	time.Time
}

// MarshalJSON is an implementation of json.Marshal.
func (t Time) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Time)
}

// UnmarshalJSON unmarshals the content of data as a time.Time, a blank string will keep the time at its zero value.
func (t *Time) UnmarshalJSON(data []byte) error {
	if string(data) == `""` {
		return nil
	}

	if err := json.Unmarshal(data, &t.Time); err != nil {
		return err
	}

	return nil
}

type DERPRegion struct {
	Preferred           bool    `json:"preferred,omitempty"`
	LatencyMilliseconds float64 `json:"latencyMs"`
}

type ClientSupports struct {
	HairPinning bool `json:"hairPinning"`
	IPV6        bool `json:"ipv6"`
	PCP         bool `json:"pcp"`
	PMP         bool `json:"pmp"`
	UDP         bool `json:"udp"`
	UPNP        bool `json:"upnp"`
}

type ClientConnectivity struct {
	Endpoints             []string `json:"endpoints"`
	DERP                  string   `json:"derp"`
	MappingVariesByDestIP bool     `json:"mappingVariesByDestIP"`
	// DERPLatency is mapped by region name (e.g. "New York City", "Seattle").
	DERPLatency    map[string]DERPRegion `json:"latency"`
	ClientSupports ClientSupports        `json:"clientSupports"`
}

type Device struct {
	Addresses                 []string `json:"addresses"`
	Name                      string   `json:"name"`
	ID                        string   `json:"id"`     // The legacy identifier for a device. Use NodeId instead.
	NodeID                    string   `json:"nodeId"` // The preferred identifier for a device.
	Authorized                bool     `json:"authorized"`
	User                      string   `json:"user"`
	Tags                      []string `json:"tags"`
	KeyExpiryDisabled         bool     `json:"keyExpiryDisabled"`
	BlocksIncomingConnections bool     `json:"blocksIncomingConnections"`
	ClientVersion             string   `json:"clientVersion"`
	Created                   Time     `json:"created"`
	Expires                   Time     `json:"expires"`
	Hostname                  string   `json:"hostname"`
	IsEphemeral               bool     `json:"isEphemeral"`
	IsExternal                bool     `json:"isExternal"`
	ConnectedToControl        bool     `json:"connectedToControl"`
	LastSeen                  *Time    `json:"lastSeen"` // Will be nil if ConnectedToControl is true.
	MachineKey                string   `json:"machineKey"`
	NodeKey                   string   `json:"nodeKey"`
	OS                        string   `json:"os"`
	TailnetLockError          string   `json:"tailnetLockError"`
	TailnetLockKey            string   `json:"tailnetLockKey"`
	UpdateAvailable           bool     `json:"updateAvailable"`

	// The below are only included in listings when querying `all` fields.
	AdvertisedRoutes   []string            `json:"AdvertisedRoutes"`
	EnabledRoutes      []string            `json:"enabledRoutes"`
	ClientConnectivity *ClientConnectivity `json:"clientConnectivity"`
}

type DevicePostureAttributes struct {
	Attributes map[string]any  `json:"attributes"`
	Expiries   map[string]Time `json:"expiries"`
}

type DevicePostureAttributeRequest struct {
	Value   any    `json:"value"`
	Expiry  Time   `json:"expiry"`
	Comment string `json:"comment"`
}

// GetWithAllFields gets the [Device] identified by `deviceID`.
// All fields will be populated.
//
// Using the device `NodeID` is preferred, but its numeric `ID` value can also be used.
func (dr *DevicesResource) GetWithAllFields(ctx context.Context, deviceID string) (*Device, error) {
	return dr.get(ctx, deviceID, true)
}

// Get gets the [Device] identified by `deviceID`.
//
// Using the device `NodeID` is preferred, but its numeric `ID` value can also be used.
func (dr *DevicesResource) Get(ctx context.Context, deviceID string) (*Device, error) {
	return dr.get(ctx, deviceID, false)
}

func (dr *DevicesResource) get(ctx context.Context, deviceID string, allFields bool) (*Device, error) {
	req, err := dr.buildRequest(ctx, http.MethodGet, dr.buildURL("device", deviceID))
	if err != nil {
		return nil, err
	}

	if allFields {
		q := req.URL.Query()
		q.Set("fields", "all")
		req.URL.RawQuery = q.Encode()
	}

	return body[Device](dr, req)
}

// GetPostureAttributes retrieves the posture attributes of the device identified by deviceID.
//
// Using the device `NodeID` is preferred, but its numeric `ID` value can also be used.
func (dr *DevicesResource) GetPostureAttributes(ctx context.Context, deviceID string) (*DevicePostureAttributes, error) {
	req, err := dr.buildRequest(ctx, http.MethodGet, dr.buildURL("device", deviceID, "attributes"))
	if err != nil {
		return nil, err
	}

	return body[DevicePostureAttributes](dr, req)
}

// SetPostureAttribute sets the posture attribute of the device identified by deviceID.
//
// Using the device `NodeID` is preferred, but its numeric `ID` value can also be used.
func (dr *DevicesResource) SetPostureAttribute(ctx context.Context, deviceID, attributeKey string, request DevicePostureAttributeRequest) error {
	req, err := dr.buildRequest(ctx, http.MethodPost, dr.buildURL("device", deviceID, "attributes", attributeKey), requestBody(request))
	if err != nil {
		return err
	}

	return dr.do(req, nil)
}

// DeletePostureAttribute deletes the posture attribute of the device identified by deviceID.
//
// Using the device `NodeID` is preferred, but its numeric `ID` value can also be used.
func (dr *DevicesResource) DeletePostureAttribute(ctx context.Context, deviceID, attributeKey string) error {
	req, err := dr.buildRequest(ctx, http.MethodDelete, dr.buildURL("device", deviceID, "attributes", attributeKey))
	if err != nil {
		return err
	}

	return dr.do(req, nil)
}

// ListWithAllFields lists every [Device] in the tailnet. Each [Device] in
// the response will have all fields populated.
func (dr *DevicesResource) ListWithAllFields(ctx context.Context) ([]Device, error) {
	return dr.list(ctx, true)
}

// List lists every [Device] in the tailnet. The fields `EnabledRoutes`,
// `AdvertisedRoutes` and `ClientConnectivity` will be omitted from the resulting
// [Devices]. To get these fields, use `ListWithAllFields`.
func (dr *DevicesResource) List(ctx context.Context) ([]Device, error) {
	return dr.list(ctx, false)
}

func (dr *DevicesResource) list(ctx context.Context, allFields bool) ([]Device, error) {
	req, err := dr.buildRequest(ctx, http.MethodGet, dr.buildTailnetURL("devices"))
	if err != nil {
		return nil, err
	}

	if allFields {
		q := req.URL.Query()
		q.Set("fields", "all")
		req.URL.RawQuery = q.Encode()
	}

	m := make(map[string][]Device)
	err = dr.do(req, &m)
	if err != nil {
		return nil, err
	}

	return m["devices"], nil
}

// SetAuthorized marks the specified device as authorized or not.
//
// Using the device `NodeID` is preferred, but its numeric `ID` value can also be used.
func (dr *DevicesResource) SetAuthorized(ctx context.Context, deviceID string, authorized bool) error {
	req, err := dr.buildRequest(ctx, http.MethodPost, dr.buildURL("device", deviceID, "authorized"), requestBody(map[string]bool{
		"authorized": authorized,
	}))
	if err != nil {
		return err
	}

	return dr.do(req, nil)
}

// Delete deletes the device identified by deviceID.
//
// Using the device `NodeID` is preferred, but its numeric `ID` value can also be used.
func (dr *DevicesResource) Delete(ctx context.Context, deviceID string) error {
	req, err := dr.buildRequest(ctx, http.MethodDelete, dr.buildURL("device", deviceID))
	if err != nil {
		return err
	}

	return dr.do(req, nil)
}

// SetName updates the name of the device identified by deviceID.
//
// Using the device `NodeID` is preferred, but its numeric `ID` value can also be used.
func (dr *DevicesResource) SetName(ctx context.Context, deviceID, name string) error {
	req, err := dr.buildRequest(ctx, http.MethodPost, dr.buildURL("device", deviceID, "name"), requestBody(map[string]string{
		"name": name,
	}))
	if err != nil {
		return err
	}

	return dr.do(req, nil)
}

// SetTags updates the tags of the device identified by deviceID.
//
// Using the device `NodeID` is preferred, but its numeric `ID` value can also be used.
func (dr *DevicesResource) SetTags(ctx context.Context, deviceID string, tags []string) error {
	req, err := dr.buildRequest(ctx, http.MethodPost, dr.buildURL("device", deviceID, "tags"), requestBody(map[string][]string{
		"tags": tags,
	}))
	if err != nil {
		return err
	}

	return dr.do(req, nil)
}

// DeviceKey type represents the properties of the key of an individual device within
// the tailnet.
type DeviceKey struct {
	KeyExpiryDisabled bool `json:"keyExpiryDisabled"` // Whether or not this device's key will ever expire.
}

// SetKey updates the properties of a device's key.
//
// Using the device `NodeID` is preferred, but its numeric `ID` value can also be used.
func (dr *DevicesResource) SetKey(ctx context.Context, deviceID string, key DeviceKey) error {
	req, err := dr.buildRequest(ctx, http.MethodPost, dr.buildURL("device", deviceID, "key"), requestBody(key))
	if err != nil {
		return err
	}

	return dr.do(req, nil)
}

// SetDeviceIPv4Address sets the Tailscale IPv4 address of the device.
//
// Using the device `NodeID` is preferred, but its numeric `ID` value can also be used.
func (dr *DevicesResource) SetIPv4Address(ctx context.Context, deviceID string, ipv4Address string) error {
	req, err := dr.buildRequest(ctx, http.MethodPost, dr.buildURL("device", deviceID, "ip"), requestBody(map[string]string{
		"ipv4": ipv4Address,
	}))
	if err != nil {
		return err
	}

	return dr.do(req, nil)
}

// SetSubnetRoutes sets which subnet routes are enabled to be routed by a device by replacing the existing list
// of subnet routes with the supplied routes. Routes can be enabled without a device advertising them (e.g. for preauth).
//
// Using the device `NodeID` is preferred, but its numeric `ID` value can also be used.
func (dr *DevicesResource) SetSubnetRoutes(ctx context.Context, deviceID string, routes []string) error {
	req, err := dr.buildRequest(ctx, http.MethodPost, dr.buildURL("device", deviceID, "routes"), requestBody(map[string][]string{
		"routes": routes,
	}))
	if err != nil {
		return err
	}

	return dr.do(req, nil)
}

// SubnetRoutes Retrieves the list of subnet routes that a device is advertising, as well as those that are
// enabled for it. Enabled routes are not necessarily advertised (e.g. for pre-enabling), and likewise, advertised
// routes are not necessarily enabled.
//
// Using the device `NodeID` is preferred, but its numeric `ID` value can also be used.
func (dr *DevicesResource) SubnetRoutes(ctx context.Context, deviceID string) (*DeviceRoutes, error) {
	req, err := dr.buildRequest(ctx, http.MethodGet, dr.buildURL("device", deviceID, "routes"))
	if err != nil {
		return nil, err
	}

	return body[DeviceRoutes](dr, req)
}

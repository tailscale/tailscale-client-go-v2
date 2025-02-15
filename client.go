// Copyright (c) David Bond, Tailscale Inc, & Contributors
// SPDX-License-Identifier: MIT

// package tailscale contains a basic implementation of a client for the Tailscale HTTP API.
//
// Documentation is at https://tailscale.com/api
package tailscale

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/tailscale/hujson"
)

// Client is used to perform actions against the Tailscale API.
type Client struct {
	// BaseURL is the base URL for accessing the Tailscale API server. Defaults to https://api.tailscale.com.
	BaseURL *url.URL
	// UserAgent configures the User-Agent HTTP header for requests. Defaults to "tailscale-client-go".
	UserAgent string
	// APIKey allows specifying an APIKey to use for authentication.
	// To use OAuth Client credentials, construct an [http.Client] using [OAuthConfig] and specify that below.
	APIKey string
	// Tailnet allows specifying a specific Tailnet by name, to which this Client will connect by default.
	Tailnet string

	// HTTP is the [http.Client] to use for requests to the API server.
	// If not specified, a new [http.Client] with a Timeout of 1 minute will be used.
	HTTP *http.Client

	initOnce sync.Once

	// Specific resources
	contacts        *ContactsResource
	devicePosture   *DevicePostureResource
	devices         *DevicesResource
	dns             *DNSResource
	keys            *KeysResource
	logging         *LoggingResource
	policyFile      *PolicyFileResource
	tailnetSettings *TailnetSettingsResource
	users           *UsersResource
	webhooks        *WebhooksResource
}

// APIError type describes an error as returned by the Tailscale API.
type APIError struct {
	Message string         `json:"message"`
	Data    []APIErrorData `json:"data"`
	status  int
}

// APIErrorData type describes elements of the data field within errors returned by the Tailscale API.
type APIErrorData struct {
	User   string   `json:"user"`
	Errors []string `json:"errors"`
}

const defaultContentType = "application/json"
const defaultHttpClientTimeout = time.Minute
const defaultUserAgent = "tailscale-client-go"

var defaultBaseURL *url.URL

func init() {
	var err error
	defaultBaseURL, err = url.Parse("https://api.tailscale.com")
	if err != nil {
		panic(fmt.Errorf("failed to parse defaultBaseURL: %w", err))
	}
}

// init returns a new instance of the Client type that will perform operations against a chosen tailnet and will
// provide the apiKey for authorization.
func (c *Client) init() {
	c.initOnce.Do(func() {
		if c.BaseURL == nil {
			c.BaseURL = defaultBaseURL
		}
		if c.UserAgent == "" {
			c.UserAgent = defaultUserAgent
		}
		if c.HTTP == nil {
			c.HTTP = &http.Client{Timeout: defaultHttpClientTimeout}
		}
		c.contacts = &ContactsResource{c}
		c.devicePosture = &DevicePostureResource{c}
		c.devices = &DevicesResource{c}
		c.dns = &DNSResource{c}
		c.keys = &KeysResource{c}
		c.logging = &LoggingResource{c}
		c.policyFile = &PolicyFileResource{c}
		c.tailnetSettings = &TailnetSettingsResource{c}
		c.users = &UsersResource{c}
		c.webhooks = &WebhooksResource{c}
	})
}

// Contacts() provides access to https://tailscale.com/api#tag/contacts.
func (c *Client) Contacts() *ContactsResource {
	c.init()
	return c.contacts
}

// DevicePosture provides access to https://tailscale.com/api#tag/deviceposture.
func (c *Client) DevicePosture() *DevicePostureResource {
	c.init()
	return c.devicePosture
}

// Devices provides access to https://tailscale.com/api#tag/devices.
func (c *Client) Devices() *DevicesResource {
	c.init()
	return c.devices
}

// DNS provides access to https://tailscale.com/api#tag/dns.
func (c *Client) DNS() *DNSResource {
	c.init()
	return c.dns
}

// Keys provides access to https://tailscale.com/api#tag/keys.
func (c *Client) Keys() *KeysResource {
	c.init()
	return c.keys
}

// Logging provides access to https://tailscale.com/api#tag/logging.
func (c *Client) Logging() *LoggingResource {
	c.init()
	return c.logging
}

// PolicyFile provides access to https://tailscale.com/api#tag/policyfile.
func (c *Client) PolicyFile() *PolicyFileResource {
	c.init()
	return c.policyFile
}

// TailnetSettings provides access to https://tailscale.com/api#tag/tailnetsettings.
func (c *Client) TailnetSettings() *TailnetSettingsResource {
	c.init()
	return c.tailnetSettings
}

// Users provides access to https://tailscale.com/api#tag/users.
func (c *Client) Users() *UsersResource {
	c.init()
	return c.users
}

// Webhooks provides access to https://tailscale.com/api#tag/webhooks.
func (c *Client) Webhooks() *WebhooksResource {
	c.init()
	return c.webhooks
}

type requestParams struct {
	headers     map[string]string
	body        any
	contentType string
}

type requestOption func(*requestParams)

func requestBody(body any) requestOption {
	return func(rof *requestParams) {
		rof.body = body
	}
}

func requestHeaders(headers map[string]string) requestOption {
	return func(rof *requestParams) {
		rof.headers = headers
	}
}

func requestContentType(ct string) requestOption {
	return func(rof *requestParams) {
		rof.contentType = ct
	}
}

// buildURL builds a url to /api/v2/... using the given pathElements.
// It url escapes each path element, so the caller doesn't need to worry about that.
func (c *Client) buildURL(pathElements ...any) *url.URL {
	elem := make([]string, 1, len(pathElements)+1)
	elem[0] = "/api/v2"
	for _, pathElement := range pathElements {
		elem = append(elem, url.PathEscape(fmt.Sprint(pathElement)))
	}
	return c.BaseURL.JoinPath(elem...)
}

// buildTailnetURL builds a url to /api/v2/tailnet/<tailnet>/... using the given pathElements.
// It url escapes each path element, so the caller doesn't need to worry about that.
func (c *Client) buildTailnetURL(pathElements ...any) *url.URL {
	allElements := make([]any, 2, len(pathElements)+2)
	allElements[0] = "tailnet"
	allElements[1] = c.Tailnet
	allElements = append(allElements, pathElements...)
	return c.buildURL(allElements...)
}

func (c *Client) buildRequest(ctx context.Context, method string, uri *url.URL, opts ...requestOption) (*http.Request, error) {
	rof := &requestParams{
		contentType: defaultContentType,
	}
	for _, opt := range opts {
		opt(rof)
	}

	var bodyBytes []byte
	if rof.body != nil {
		switch body := rof.body.(type) {
		case string:
			bodyBytes = []byte(body)
		case []byte:
			bodyBytes = body
		default:
			var err error
			bodyBytes, err = json.Marshal(rof.body)
			if err != nil {
				return nil, err
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, uri.String(), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", c.UserAgent)
	}

	for k, v := range rof.headers {
		req.Header.Set(k, v)
	}

	switch {
	case rof.body == nil:
		req.Header.Set("Accept", rof.contentType)
	default:
		req.Header.Set("Content-Type", rof.contentType)
	}

	if c.APIKey != "" {
		req.SetBasicAuth(c.APIKey, "")
	}

	return req, nil
}

// doer is a resource type (such as *ContactsResource) with a doWithResponseHeaders
// method that sends an HTTP request and decodes its body into out.
//
// Concretely, the doWithResponseHeaders method will usually be (*Client).doWithResponseHeaders,
// as all the Resource types embed a *Client.
type doer interface {
	doWithResponseHeaders(req *http.Request, out any) (http.Header, error)
}

// body calls resource.do, passing a *T to do, and returns
// exactly one non-zero value depending on the result of do.
func body[T any](resource doer, req *http.Request) (*T, error) {
	t, _, err := bodyWithResponseHeader[T](resource, req)
	return t, err
}

// bodyWithResponseHeader is like [body] but also returns the response header.
func bodyWithResponseHeader[T any](resource doer, req *http.Request) (*T, http.Header, error) {
	var v T
	header, err := resource.doWithResponseHeaders(req, &v)
	if err != nil {
		return nil, nil, err
	}
	return &v, header, nil
}

func (c *Client) do(req *http.Request, out any) error {
	_, err := c.doWithResponseHeaders(req, out)
	return err
}

func (c *Client) doWithResponseHeaders(req *http.Request, out any) (http.Header, error) {
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices {
		// If we don't care about the response body, leave. This check is required as some
		// API responses have empty bodies, so we don't want to try and standardize them for
		// parsing.
		if out == nil {
			return res.Header, nil
		}

		// If we're expected to write result into a []byte, do not attempt to parse it.
		if o, ok := out.(*[]byte); ok {
			*o = bytes.Clone(body)
			return res.Header, nil
		}

		// If we've got hujson back, convert it to JSON, so we can natively parse it.
		if !json.Valid(body) {
			body, err = hujson.Standardize(body)
			if err != nil {
				return res.Header, err
			}
		}

		return res.Header, json.Unmarshal(body, out)
	}

	if res.StatusCode >= http.StatusBadRequest {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return res.Header, err
		}

		apiErr.status = res.StatusCode
		return res.Header, apiErr
	}

	return res.Header, nil
}

func (err APIError) Error() string {
	return fmt.Sprintf("%s (%v)", err.Message, err.status)
}

// IsNotFound returns true if the provided error implementation is an APIError with a status of 404.
func IsNotFound(err error) bool {
	var apiErr APIError
	if errors.As(err, &apiErr) {
		return apiErr.status == http.StatusNotFound
	}

	return false
}

// ErrorData returns the contents of the [APIError].Data field from the provided error if it is of type [APIError].
// Returns a nil slice if the given error is not of type [APIError].
func ErrorData(err error) []APIErrorData {
	var apiErr APIError
	if errors.As(err, &apiErr) {
		return apiErr.Data
	}

	return nil
}

// PointerTo returns a pointer to the given value.
// Pointers are used in PATCH requests to distinguish between specified and unspecified values.
func PointerTo[T any](value T) *T {
	return &value
}

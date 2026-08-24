// Package sendgrid is a minimal REST client for the SendGrid v3 API,
// covering only the endpoints this provider needs.
package sendgrid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const DefaultHost = "https://api.sendgrid.com"

// Client is a SendGrid v3 API client.
type Client struct {
	apiKey     string
	host       string
	onBehalfOf string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithHost overrides the API host, e.g. for EU data residency
// ("https://api.eu.sendgrid.com").
func WithHost(host string) Option {
	return func(c *Client) {
		if host != "" {
			c.host = strings.TrimRight(host, "/")
		}
	}
}

// WithOnBehalfOf sets the On-Behalf-Of header, causing requests to act as
// the given subuser or account.
func WithOnBehalfOf(username string) Option {
	return func(c *Client) {
		c.onBehalfOf = username
	}
}

// WithHTTPClient overrides the underlying http.Client (used in tests).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// NewClient creates a new SendGrid API client authenticated with apiKey.
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey: apiKey,
		host:   DefaultHost,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// APIError represents an error response from the SendGrid API.
type APIError struct {
	StatusCode int
	Errors     []struct {
		Message string `json:"message"`
		Field   string `json:"field"`
		Help    string `json:"help"`
	} `json:"errors"`
}

func (e *APIError) Error() string {
	if len(e.Errors) == 0 {
		return fmt.Sprintf("sendgrid: request failed with status %d", e.StatusCode)
	}
	msgs := make([]string, 0, len(e.Errors))
	for _, apiErr := range e.Errors {
		if apiErr.Field != "" {
			msgs = append(msgs, fmt.Sprintf("%s (field: %s)", apiErr.Message, apiErr.Field))
		} else {
			msgs = append(msgs, apiErr.Message)
		}
	}
	return fmt.Sprintf("sendgrid: request failed with status %d: %s", e.StatusCode, strings.Join(msgs, "; "))
}

// IsNotFound reports whether err represents a 404 response from the API.
func IsNotFound(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.StatusCode == http.StatusNotFound
}

// requestOptions carries per-request overrides.
type requestOptions struct {
	onBehalfOf string
	query      map[string]string
}

type RequestOption func(*requestOptions)

// OnBehalfOf overrides the client-level On-Behalf-Of header for a single request.
func OnBehalfOf(username string) RequestOption {
	return func(o *requestOptions) { o.onBehalfOf = username }
}

// Query adds a query string parameter to a single request.
func Query(key, value string) RequestOption {
	return func(o *requestOptions) {
		if o.query == nil {
			o.query = map[string]string{}
		}
		o.query[key] = value
	}
}

// do performs an HTTP request against the SendGrid API. body and out may be
// nil. A non-2xx response is returned as *APIError.
func (c *Client) do(ctx context.Context, method, path string, body, out any, opts ...RequestOption) error {
	var reqOpts requestOptions
	for _, opt := range opts {
		opt(&reqOpts)
	}

	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("sendgrid: encoding request body: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	url := c.host + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return fmt.Errorf("sendgrid: building request: %w", err)
	}

	if len(reqOpts.query) > 0 {
		q := req.URL.Query()
		for k, v := range reqOpts.query {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	onBehalfOf := c.onBehalfOf
	if reqOpts.onBehalfOf != "" {
		onBehalfOf = reqOpts.onBehalfOf
	}
	if onBehalfOf != "" {
		req.Header.Set("On-Behalf-Of", onBehalfOf)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("sendgrid: performing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("sendgrid: reading response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		apiErr := &APIError{StatusCode: resp.StatusCode}
		// Best-effort decode; some error responses aren't JSON.
		_ = json.Unmarshal(respBody, apiErr)
		return apiErr
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("sendgrid: decoding response body: %w", err)
		}
	}

	return nil
}

func (c *Client) get(ctx context.Context, path string, out any, opts ...RequestOption) error {
	return c.do(ctx, http.MethodGet, path, nil, out, opts...)
}

func (c *Client) post(ctx context.Context, path string, body, out any, opts ...RequestOption) error {
	return c.do(ctx, http.MethodPost, path, body, out, opts...)
}

func (c *Client) patch(ctx context.Context, path string, body, out any, opts ...RequestOption) error {
	return c.do(ctx, http.MethodPatch, path, body, out, opts...)
}

func (c *Client) put(ctx context.Context, path string, body, out any, opts ...RequestOption) error {
	return c.do(ctx, http.MethodPut, path, body, out, opts...)
}

func (c *Client) delete(ctx context.Context, path string, opts ...RequestOption) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil, opts...)
}

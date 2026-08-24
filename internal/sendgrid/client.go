// Package sendgrid is a minimal REST client for the SendGrid v3 API,
// covering only the endpoints this provider needs.
package sendgrid

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// esc percent-encodes a caller-supplied value (a resource ID, name, or IP)
// before it's concatenated into a request path, so a value containing "/",
// "?", "#", or other URL-meaningful characters can't redirect the request
// to a different path or inject query parameters.
func esc(s string) string {
	return url.PathEscape(s)
}

const DefaultHost = "https://api.sendgrid.com"

// defaultMaxRetries and defaultRetryWait control retry behavior for
// requests that fail with a retryable status (429 or 5xx). The wait
// doubles each attempt (capped) unless the response includes a
// Retry-After header, which takes precedence.
const (
	defaultMaxRetries = 4
	defaultRetryWait  = 500 * time.Millisecond
	maxRetryWait      = 30 * time.Second
)

// Client is a SendGrid v3 API client.
type Client struct {
	apiKey     string
	host       string
	onBehalfOf string
	userAgent  string
	httpClient *http.Client
	maxRetries int
	retryWait  time.Duration
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

// WithUserAgent sets the User-Agent header sent with every request, e.g.
// "terraform-provider-sendgrid/1.2.3".
func WithUserAgent(userAgent string) Option {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// WithMaxRetries overrides how many times a request is retried after a
// retryable (429 or 5xx) response. 0 disables retries.
func WithMaxRetries(n int) Option {
	return func(c *Client) {
		c.maxRetries = n
	}
}

// WithRetryWait overrides the base wait between retries (used in tests to
// avoid slow test runs; production callers should rely on the default).
func WithRetryWait(d time.Duration) Option {
	return func(c *Client) {
		c.retryWait = d
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
		maxRetries: defaultMaxRetries,
		retryWait:  defaultRetryWait,
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

// IsNotFound reports whether err represents a 404 response from the API,
// including when err wraps an *APIError via %w.
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// requestOptions carries per-request overrides.
type requestOptions struct {
	query map[string]string
}

type RequestOption func(*requestOptions)

// Query adds a query string parameter to a single request.
func Query(key, value string) RequestOption {
	return func(o *requestOptions) {
		if o.query == nil {
			o.query = map[string]string{}
		}
		o.query[key] = value
	}
}

// isRetryableStatus reports whether a response is worth retrying: SendGrid
// rate limiting (429) or a transient server-side error (5xx). Other 4xx
// responses represent a request the client got wrong, and retrying won't
// help.
func isRetryableStatus(code int) bool {
	return code == http.StatusTooManyRequests || (code >= 500 && code <= 599)
}

// retryDelay returns how long to wait before the next attempt. A
// Retry-After response header (seconds, per SendGrid's rate-limit
// behavior) takes precedence over the exponential backoff schedule.
func retryDelay(retryAfterHeader string, attempt int, base time.Duration) time.Duration {
	if retryAfterHeader != "" {
		if secs, err := strconv.Atoi(retryAfterHeader); err == nil && secs >= 0 {
			return time.Duration(secs) * time.Second
		}
	}
	d := base * time.Duration(math.Pow(2, float64(attempt)))
	if d > maxRetryWait {
		return maxRetryWait
	}
	return d
}

// do performs an HTTP request against the SendGrid API, retrying on 429/5xx
// responses up to c.maxRetries times. body and out may be nil. A non-2xx
// response (after retries are exhausted) is returned as *APIError.
func (c *Client) do(ctx context.Context, method, path string, body, out any, opts ...RequestOption) error {
	var reqOpts requestOptions
	for _, opt := range opts {
		opt(&reqOpts)
	}

	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("sendgrid: encoding request body: %w", err)
		}
		bodyBytes = b
	}

	fullURL := c.host + path

	for attempt := 0; ; attempt++ {
		var reqBody io.Reader
		if bodyBytes != nil {
			reqBody = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
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
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if c.onBehalfOf != "" {
			req.Header.Set("On-Behalf-Of", c.onBehalfOf)
		}
		if c.userAgent != "" {
			req.Header.Set("User-Agent", c.userAgent)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("sendgrid: performing request: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		retryAfter := resp.Header.Get("Retry-After")
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("sendgrid: reading response body: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			apiErr := &APIError{StatusCode: resp.StatusCode}
			// Best-effort decode; some error responses aren't JSON.
			_ = json.Unmarshal(respBody, apiErr)

			if !isRetryableStatus(resp.StatusCode) || attempt >= c.maxRetries {
				return apiErr
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay(retryAfter, attempt, c.retryWait)):
			}
			continue
		}

		if out != nil && len(respBody) > 0 {
			if err := json.Unmarshal(respBody, out); err != nil {
				return fmt.Errorf("sendgrid: decoding response body: %w", err)
			}
		}
		return nil
	}
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

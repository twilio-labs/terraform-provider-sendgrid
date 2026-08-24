package sendgrid

import "context"

// EventWebhook represents a single Event Webhook configuration.
type EventWebhook struct {
	ID               string `json:"id,omitempty"`
	Enabled          bool   `json:"enabled"`
	URL              string `json:"url"`
	FriendlyName     string `json:"friendly_name,omitempty"`
	GroupResubscribe bool   `json:"group_resubscribe"`
	Delivered        bool   `json:"delivered"`
	GroupUnsubscribe bool   `json:"group_unsubscribe"`
	SpamReport       bool   `json:"spam_report"`
	Bounce           bool   `json:"bounce"`
	Deferred         bool   `json:"deferred"`
	Unsubscribe      bool   `json:"unsubscribe"`
	Processed        bool   `json:"processed"`
	Open             bool   `json:"open"`
	Click            bool   `json:"click"`
	Dropped          bool   `json:"dropped"`
	OAuthClientID    string `json:"oauth_client_id,omitempty"`
	OAuthTokenURL    string `json:"oauth_token_url,omitempty"`
	CreatedAt        string `json:"created_date,omitempty"`
	UpdatedAt        string `json:"updated_date,omitempty"`
}

// EventWebhookParams are the mutable fields of an Event Webhook, used for
// both create and update.
//
// FriendlyName, OAuthClientID, OAuthClientSecret, and OAuthTokenURL
// deliberately lack `omitempty`: the SendGrid API's documented way to clear
// a previously-set value on update is to send an explicit empty string, and
// `omitempty` would drop the field from the request entirely instead,
// leaving the old value in place server-side.
type EventWebhookParams struct {
	Enabled           bool   `json:"enabled"`
	URL               string `json:"url"`
	FriendlyName      string `json:"friendly_name"`
	GroupResubscribe  bool   `json:"group_resubscribe"`
	Delivered         bool   `json:"delivered"`
	GroupUnsubscribe  bool   `json:"group_unsubscribe"`
	SpamReport        bool   `json:"spam_report"`
	Bounce            bool   `json:"bounce"`
	Deferred          bool   `json:"deferred"`
	Unsubscribe       bool   `json:"unsubscribe"`
	Processed         bool   `json:"processed"`
	Open              bool   `json:"open"`
	Click             bool   `json:"click"`
	Dropped           bool   `json:"dropped"`
	OAuthClientID     string `json:"oauth_client_id"`
	OAuthClientSecret string `json:"oauth_client_secret"`
	OAuthTokenURL     string `json:"oauth_token_url"`
}

// CreateEventWebhook creates a new Event Webhook configuration.
func (c *Client) CreateEventWebhook(ctx context.Context, p EventWebhookParams, opts ...RequestOption) (*EventWebhook, error) {
	var out EventWebhook
	if err := c.post(ctx, "/v3/user/webhooks/event/settings", p, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetEventWebhook retrieves a single Event Webhook by ID.
func (c *Client) GetEventWebhook(ctx context.Context, id string, opts ...RequestOption) (*EventWebhook, error) {
	var out EventWebhook
	if err := c.get(ctx, "/v3/user/webhooks/event/settings/"+esc(id), &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateEventWebhook updates an existing Event Webhook by ID.
func (c *Client) UpdateEventWebhook(ctx context.Context, id string, p EventWebhookParams, opts ...RequestOption) (*EventWebhook, error) {
	var out EventWebhook
	if err := c.patch(ctx, "/v3/user/webhooks/event/settings/"+esc(id), p, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteEventWebhook deletes an Event Webhook by ID.
func (c *Client) DeleteEventWebhook(ctx context.Context, id string, opts ...RequestOption) error {
	return c.delete(ctx, "/v3/user/webhooks/event/settings/"+esc(id), opts...)
}

// SetEventWebhookSigned enables or disables signature verification for an
// Event Webhook and returns its public key (empty if disabled).
func (c *Client) SetEventWebhookSigned(ctx context.Context, id string, enabled bool, opts ...RequestOption) (string, error) {
	body := map[string]any{"enabled": enabled}
	var out struct {
		PublicKey string `json:"public_key"`
	}
	if err := c.patch(ctx, "/v3/user/webhooks/event/settings/signed/"+esc(id), body, &out, opts...); err != nil {
		return "", err
	}
	return out.PublicKey, nil
}

// GetEventWebhookPublicKey retrieves the public key for a signed Event
// Webhook.
func (c *Client) GetEventWebhookPublicKey(ctx context.Context, id string, opts ...RequestOption) (string, error) {
	var out struct {
		PublicKey string `json:"public_key"`
	}
	if err := c.get(ctx, "/v3/user/webhooks/event/settings/signed/"+esc(id), &out, opts...); err != nil {
		return "", err
	}
	return out.PublicKey, nil
}

package sendgrid

import "context"

// APIKey represents a SendGrid API key.
type APIKey struct {
	APIKeyID string   `json:"api_key_id,omitempty"`
	APIKey   string   `json:"api_key,omitempty"`
	Name     string   `json:"name"`
	Scopes   []string `json:"scopes,omitempty"`
}

// CreateAPIKey creates a new API key. The returned APIKey.APIKey (the secret)
// is only ever populated on this call; SendGrid never returns it again.
func (c *Client) CreateAPIKey(ctx context.Context, name string, scopes []string, opts ...RequestOption) (*APIKey, error) {
	body := map[string]any{"name": name}
	if len(scopes) > 0 {
		body["scopes"] = scopes
	}
	var out APIKey
	if err := c.post(ctx, "/v3/api_keys", body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAPIKey retrieves an API key's name and scopes by ID. The secret value is
// never returned.
func (c *Client) GetAPIKey(ctx context.Context, id string, opts ...RequestOption) (*APIKey, error) {
	var out APIKey
	if err := c.get(ctx, "/v3/api_keys/"+id, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAPIKey updates an API key's name and scopes.
func (c *Client) UpdateAPIKey(ctx context.Context, id, name string, scopes []string, opts ...RequestOption) (*APIKey, error) {
	body := map[string]any{"name": name, "scopes": scopes}
	var out APIKey
	if err := c.put(ctx, "/v3/api_keys/"+id, body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteAPIKey revokes an API key.
func (c *Client) DeleteAPIKey(ctx context.Context, id string, opts ...RequestOption) error {
	return c.delete(ctx, "/v3/api_keys/"+id, opts...)
}

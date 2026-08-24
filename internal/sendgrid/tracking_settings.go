package sendgrid

import "context"

// ClickTrackingSettings represents the account's click tracking settings.
type ClickTrackingSettings struct {
	Enabled    bool `json:"enabled"`
	EnableText bool `json:"enable_text"`
}

// GetClickTrackingSettings retrieves the current click tracking settings.
func (c *Client) GetClickTrackingSettings(ctx context.Context, opts ...RequestOption) (*ClickTrackingSettings, error) {
	var out ClickTrackingSettings
	if err := c.get(ctx, "/v3/tracking_settings/click", &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateClickTrackingSettings enables or disables click tracking.
func (c *Client) UpdateClickTrackingSettings(ctx context.Context, enabled bool, opts ...RequestOption) (*ClickTrackingSettings, error) {
	body := map[string]any{"enabled": enabled}
	var out ClickTrackingSettings
	if err := c.patch(ctx, "/v3/tracking_settings/click", body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// OpenTrackingSettings represents the account's open tracking settings.
type OpenTrackingSettings struct {
	Enabled bool `json:"enabled"`
}

// GetOpenTrackingSettings retrieves the current open tracking settings.
func (c *Client) GetOpenTrackingSettings(ctx context.Context, opts ...RequestOption) (*OpenTrackingSettings, error) {
	var out OpenTrackingSettings
	if err := c.get(ctx, "/v3/tracking_settings/open", &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateOpenTrackingSettings enables or disables open tracking.
func (c *Client) UpdateOpenTrackingSettings(ctx context.Context, enabled bool, opts ...RequestOption) (*OpenTrackingSettings, error) {
	body := map[string]any{"enabled": enabled}
	var out OpenTrackingSettings
	if err := c.patch(ctx, "/v3/tracking_settings/open", body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// SubscriptionTrackingSettings represents the account's subscription
// tracking settings.
type SubscriptionTrackingSettings struct {
	Enabled      bool   `json:"enabled"`
	HTMLContent  string `json:"html_content,omitempty"`
	Landing      string `json:"landing,omitempty"`
	PlainContent string `json:"plain_content,omitempty"`
	Replace      string `json:"replace,omitempty"`
	URL          string `json:"url,omitempty"`
}

// GetSubscriptionTrackingSettings retrieves the current subscription
// tracking settings.
func (c *Client) GetSubscriptionTrackingSettings(ctx context.Context, opts ...RequestOption) (*SubscriptionTrackingSettings, error) {
	var out SubscriptionTrackingSettings
	if err := c.get(ctx, "/v3/tracking_settings/subscription", &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateSubscriptionTrackingSettings updates the account's subscription
// tracking settings.
func (c *Client) UpdateSubscriptionTrackingSettings(ctx context.Context, s SubscriptionTrackingSettings, opts ...RequestOption) (*SubscriptionTrackingSettings, error) {
	var out SubscriptionTrackingSettings
	if err := c.patch(ctx, "/v3/tracking_settings/subscription", s, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// GoogleAnalyticsSettings represents the account's Google Analytics
// tracking settings.
type GoogleAnalyticsSettings struct {
	Enabled     bool   `json:"enabled"`
	UTMCampaign string `json:"utm_campaign,omitempty"`
	UTMContent  string `json:"utm_content,omitempty"`
	UTMMedium   string `json:"utm_medium,omitempty"`
	UTMSource   string `json:"utm_source,omitempty"`
	UTMTerm     string `json:"utm_term,omitempty"`
}

// GetGoogleAnalyticsSettings retrieves the current Google Analytics
// tracking settings.
func (c *Client) GetGoogleAnalyticsSettings(ctx context.Context, opts ...RequestOption) (*GoogleAnalyticsSettings, error) {
	var out GoogleAnalyticsSettings
	if err := c.get(ctx, "/v3/tracking_settings/google_analytics", &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateGoogleAnalyticsSettings updates the account's Google Analytics
// tracking settings.
func (c *Client) UpdateGoogleAnalyticsSettings(ctx context.Context, s GoogleAnalyticsSettings, opts ...RequestOption) (*GoogleAnalyticsSettings, error) {
	var out GoogleAnalyticsSettings
	if err := c.patch(ctx, "/v3/tracking_settings/google_analytics", s, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

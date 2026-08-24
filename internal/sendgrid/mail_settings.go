package sendgrid

import "context"

// AddressWhitelistMailSettings represents the account's address whitelist
// mail setting.
type AddressWhitelistMailSettings struct {
	Enabled bool     `json:"enabled"`
	List    []string `json:"list"`
}

// GetAddressWhitelistMailSettings retrieves the current address whitelist
// mail setting.
func (c *Client) GetAddressWhitelistMailSettings(ctx context.Context, opts ...RequestOption) (*AddressWhitelistMailSettings, error) {
	var out AddressWhitelistMailSettings
	if err := c.get(ctx, "/v3/mail_settings/address_whitelist", &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateAddressWhitelistMailSettings updates the account's address
// whitelist mail setting.
func (c *Client) UpdateAddressWhitelistMailSettings(ctx context.Context, s AddressWhitelistMailSettings, opts ...RequestOption) (*AddressWhitelistMailSettings, error) {
	var out AddressWhitelistMailSettings
	if err := c.patch(ctx, "/v3/mail_settings/address_whitelist", s, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// BouncePurgeMailSettings represents the account's bounce purge mail
// setting.
type BouncePurgeMailSettings struct {
	Enabled     bool `json:"enabled"`
	SoftBounces *int `json:"soft_bounces"`
	HardBounces *int `json:"hard_bounces"`
}

// GetBouncePurgeMailSettings retrieves the current bounce purge mail
// setting.
func (c *Client) GetBouncePurgeMailSettings(ctx context.Context, opts ...RequestOption) (*BouncePurgeMailSettings, error) {
	var out BouncePurgeMailSettings
	if err := c.get(ctx, "/v3/mail_settings/bounce_purge", &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateBouncePurgeMailSettings updates the account's bounce purge mail
// setting.
func (c *Client) UpdateBouncePurgeMailSettings(ctx context.Context, s BouncePurgeMailSettings, opts ...RequestOption) (*BouncePurgeMailSettings, error) {
	var out BouncePurgeMailSettings
	if err := c.patch(ctx, "/v3/mail_settings/bounce_purge", s, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// FooterMailSettings represents the account's footer mail setting.
type FooterMailSettings struct {
	Enabled      bool   `json:"enabled"`
	HTMLContent  string `json:"html_content,omitempty"`
	PlainContent string `json:"plain_content,omitempty"`
}

// GetFooterMailSettings retrieves the current footer mail setting.
func (c *Client) GetFooterMailSettings(ctx context.Context, opts ...RequestOption) (*FooterMailSettings, error) {
	var out FooterMailSettings
	if err := c.get(ctx, "/v3/mail_settings/footer", &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateFooterMailSettings updates the account's footer mail setting.
func (c *Client) UpdateFooterMailSettings(ctx context.Context, s FooterMailSettings, opts ...RequestOption) (*FooterMailSettings, error) {
	var out FooterMailSettings
	if err := c.patch(ctx, "/v3/mail_settings/footer", s, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// ForwardBounceMailSettings represents the account's bounce forwarding mail
// setting.
type ForwardBounceMailSettings struct {
	Enabled bool    `json:"enabled"`
	Email   *string `json:"email"`
}

// GetForwardBounceMailSettings retrieves the current bounce forwarding mail
// setting.
func (c *Client) GetForwardBounceMailSettings(ctx context.Context, opts ...RequestOption) (*ForwardBounceMailSettings, error) {
	var out ForwardBounceMailSettings
	if err := c.get(ctx, "/v3/mail_settings/forward_bounce", &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateForwardBounceMailSettings updates the account's bounce forwarding
// mail setting.
func (c *Client) UpdateForwardBounceMailSettings(ctx context.Context, s ForwardBounceMailSettings, opts ...RequestOption) (*ForwardBounceMailSettings, error) {
	var out ForwardBounceMailSettings
	if err := c.patch(ctx, "/v3/mail_settings/forward_bounce", s, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// ForwardSpamMailSettings represents the account's spam forwarding mail
// setting.
type ForwardSpamMailSettings struct {
	Enabled bool   `json:"enabled"`
	Email   string `json:"email"`
}

// GetForwardSpamMailSettings retrieves the current spam forwarding mail
// setting.
func (c *Client) GetForwardSpamMailSettings(ctx context.Context, opts ...RequestOption) (*ForwardSpamMailSettings, error) {
	var out ForwardSpamMailSettings
	if err := c.get(ctx, "/v3/mail_settings/forward_spam", &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateForwardSpamMailSettings updates the account's spam forwarding mail
// setting.
func (c *Client) UpdateForwardSpamMailSettings(ctx context.Context, s ForwardSpamMailSettings, opts ...RequestOption) (*ForwardSpamMailSettings, error) {
	var out ForwardSpamMailSettings
	if err := c.patch(ctx, "/v3/mail_settings/forward_spam", s, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// TemplateMailSettings represents the account's legacy email template mail
// setting.
type TemplateMailSettings struct {
	Enabled     bool   `json:"enabled"`
	HTMLContent string `json:"html_content,omitempty"`
}

// GetTemplateMailSettings retrieves the current legacy email template mail
// setting.
func (c *Client) GetTemplateMailSettings(ctx context.Context, opts ...RequestOption) (*TemplateMailSettings, error) {
	var out TemplateMailSettings
	if err := c.get(ctx, "/v3/mail_settings/template", &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateTemplateMailSettings updates the account's legacy email template
// mail setting.
func (c *Client) UpdateTemplateMailSettings(ctx context.Context, s TemplateMailSettings, opts ...RequestOption) (*TemplateMailSettings, error) {
	var out TemplateMailSettings
	if err := c.patch(ctx, "/v3/mail_settings/template", s, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

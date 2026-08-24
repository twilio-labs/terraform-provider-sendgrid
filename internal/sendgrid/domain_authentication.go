package sendgrid

import (
	"context"
	"fmt"
)

// DNSRecord is a single DNS record required to validate a domain
// authentication.
type DNSRecord struct {
	Valid bool   `json:"valid"`
	Type  string `json:"type"`
	Host  string `json:"host"`
	Data  string `json:"data"`
}

// DomainAuthenticationDNS holds the DNS records SendGrid expects to find for
// a domain authentication to be considered valid.
type DomainAuthenticationDNS struct {
	MailCNAME DNSRecord `json:"mail_cname"`
	DKIM1     DNSRecord `json:"dkim1"`
	DKIM2     DNSRecord `json:"dkim2"`
}

// DomainAuthentication represents a SendGrid authenticated (whitelabeled)
// sending domain.
type DomainAuthentication struct {
	ID                 int64                   `json:"id,omitempty"`
	UserID             int64                   `json:"user_id,omitempty"`
	Domain             string                  `json:"domain"`
	Subdomain          string                  `json:"subdomain,omitempty"`
	Username           string                  `json:"username,omitempty"`
	IPs                []string                `json:"ips,omitempty"`
	CustomSPF          bool                    `json:"custom_spf"`
	Default            bool                    `json:"default"`
	Legacy             bool                    `json:"legacy"`
	AutomaticSecurity  bool                    `json:"automatic_security"`
	CustomDKIMSelector string                  `json:"custom_dkim_selector,omitempty"`
	Region             string                  `json:"region,omitempty"`
	Valid              bool                    `json:"valid"`
	DNS                DomainAuthenticationDNS `json:"dns"`
}

// CreateDomainAuthenticationParams are the fields accepted when creating a
// new domain authentication. All fields are immutable after creation except
// where noted on DomainAuthentication.
type CreateDomainAuthenticationParams struct {
	Domain             string   `json:"domain"`
	Subdomain          string   `json:"subdomain,omitempty"`
	Username           string   `json:"username,omitempty"`
	IPs                []string `json:"ips,omitempty"`
	CustomSPF          bool     `json:"custom_spf"`
	Default            bool     `json:"default"`
	AutomaticSecurity  bool     `json:"automatic_security"`
	CustomDKIMSelector string   `json:"custom_dkim_selector,omitempty"`
	Region             string   `json:"region,omitempty"`
}

// CreateDomainAuthentication authenticates a new sending domain.
func (c *Client) CreateDomainAuthentication(ctx context.Context, req CreateDomainAuthenticationParams, opts ...RequestOption) (*DomainAuthentication, error) {
	var out DomainAuthentication
	if err := c.post(ctx, "/v3/whitelabel/domains", req, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetDomainAuthentication retrieves a single authenticated domain by ID.
func (c *Client) GetDomainAuthentication(ctx context.Context, id string, opts ...RequestOption) (*DomainAuthentication, error) {
	var out DomainAuthentication
	if err := c.get(ctx, "/v3/whitelabel/domains/"+id, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateDomainAuthentication updates the default and custom_spf flags of an
// authenticated domain; these are the only mutable fields.
func (c *Client) UpdateDomainAuthentication(ctx context.Context, id string, defaultDomain, customSPF bool, opts ...RequestOption) (*DomainAuthentication, error) {
	body := map[string]any{"default": defaultDomain, "custom_spf": customSPF}
	var out DomainAuthentication
	if err := c.patch(ctx, "/v3/whitelabel/domains/"+id, body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteDomainAuthentication removes an authenticated domain.
func (c *Client) DeleteDomainAuthentication(ctx context.Context, id string, opts ...RequestOption) error {
	return c.delete(ctx, fmt.Sprintf("/v3/whitelabel/domains/%s", id), opts...)
}

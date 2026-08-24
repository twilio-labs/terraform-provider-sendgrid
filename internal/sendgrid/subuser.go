package sendgrid

import (
	"context"
	"fmt"
)

// Subuser represents a SendGrid subuser account. ID is the subuser's numeric
// identifier regardless of which endpoint populated it — the create endpoint
// returns it as "user_id" and the list endpoint returns it as "id"; callers
// in this package normalize both into this one field so nothing downstream
// has to know which endpoint a given Subuser value came from.
type Subuser struct {
	ID       int64  `json:"-"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
	Disabled bool   `json:"disabled"`
	Region   string `json:"region,omitempty"`
}

// subuserListEntry matches the shape returned by GET /v3/subusers, where the
// numeric identifier is named "id".
type subuserListEntry struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Disabled bool   `json:"disabled"`
	Region   string `json:"region,omitempty"`
}

type createSubuserRequest struct {
	Username      string   `json:"username"`
	Email         string   `json:"email"`
	Password      string   `json:"password"`
	IPs           []string `json:"ips"`
	Region        string   `json:"region,omitempty"`
	IncludeRegion bool     `json:"include_region"`
}

type createSubuserResponse struct {
	Username string `json:"username"`
	UserID   int64  `json:"user_id"`
	Email    string `json:"email"`
	Region   string `json:"region,omitempty"`
}

// CreateSubuser creates a new subuser. SendGrid requires at least one IP to
// be assigned at creation time.
func (c *Client) CreateSubuser(ctx context.Context, username, email, password string, ips []string, region string, opts ...RequestOption) (*Subuser, error) {
	req := createSubuserRequest{
		Username:      username,
		Email:         email,
		Password:      password,
		IPs:           ips,
		Region:        region,
		IncludeRegion: true,
	}
	var out createSubuserResponse
	if err := c.post(ctx, "/v3/subusers", req, &out, opts...); err != nil {
		return nil, err
	}
	return &Subuser{
		ID:       out.UserID,
		Username: out.Username,
		Email:    out.Email,
		Region:   out.Region,
	}, nil
}

// GetSubuserByUsername retrieves a subuser by username. There is no
// single-resource GET endpoint for subusers, so this filters the list.
// Returns nil, nil if no matching subuser exists.
func (c *Client) GetSubuserByUsername(ctx context.Context, username string, opts ...RequestOption) (*Subuser, error) {
	opts = append(opts, Query("username", username), Query("include_region", "true"), Query("region", "all"))
	var out []subuserListEntry
	if err := c.get(ctx, "/v3/subusers", &out, opts...); err != nil {
		return nil, err
	}
	for _, s := range out {
		if s.Username == username {
			return &Subuser{
				ID:       s.ID,
				Username: s.Username,
				Email:    s.Email,
				Disabled: s.Disabled,
				Region:   s.Region,
			}, nil
		}
	}
	return nil, nil
}

// UpdateSubuserIPs replaces the IP addresses assigned to a subuser.
func (c *Client) UpdateSubuserIPs(ctx context.Context, username string, ips []string, opts ...RequestOption) ([]string, error) {
	var out struct {
		IPs []string `json:"ips"`
	}
	if err := c.put(ctx, "/v3/subusers/"+username+"/ips", ips, &out, opts...); err != nil {
		return nil, err
	}
	return out.IPs, nil
}

// UpdateSubuserDisabled enables or disables a subuser.
func (c *Client) UpdateSubuserDisabled(ctx context.Context, username string, disabled bool, opts ...RequestOption) error {
	body := map[string]any{"disabled": disabled}
	return c.patch(ctx, "/v3/subusers/"+username, body, nil, opts...)
}

// DeleteSubuser permanently deletes a subuser.
func (c *Client) DeleteSubuser(ctx context.Context, username string, opts ...RequestOption) error {
	return c.delete(ctx, fmt.Sprintf("/v3/subusers/%s", username), opts...)
}

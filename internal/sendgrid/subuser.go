package sendgrid

import (
	"context"
	"fmt"
)

// Subuser represents a SendGrid subuser account.
type Subuser struct {
	ID       int64  `json:"id,omitempty"`
	UserID   int64  `json:"user_id,omitempty"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
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
		UserID:   out.UserID,
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
	var out []Subuser
	if err := c.get(ctx, "/v3/subusers", &out, opts...); err != nil {
		return nil, err
	}
	for _, s := range out {
		if s.Username == username {
			return &s, nil
		}
	}
	return nil, nil
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

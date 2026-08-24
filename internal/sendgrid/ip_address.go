package sendgrid

import "context"

// IPAddressPool identifies an IP pool an IP address belongs to.
type IPAddressPool struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// IPAddress represents the response from adding an IP address to the
// account via the IP Address Management API.
type IPAddress struct {
	IP               string   `json:"ip"`
	IsAutoWarmup     bool     `json:"is_auto_warmup"`
	IsParentAssigned bool     `json:"is_parent_assigned"`
	Subusers         []string `json:"subusers"`
}

// IPAddressDetail represents the full details of a single IP address, as
// returned by GetIPAddress. Subusers is populated separately via
// ListIPAddressSubusers, since the underlying GET endpoint doesn't return it.
type IPAddressDetail struct {
	IP               string          `json:"ip"`
	IsParentAssigned bool            `json:"is_parent_assigned"`
	IsAutoWarmup     bool            `json:"is_auto_warmup"`
	Pools            []IPAddressPool `json:"pools"`
	AddedAt          int64           `json:"added_at"`
	UpdatedAt        *int64          `json:"updated_at"`
	IsEnabled        *bool           `json:"is_enabled"`
	IsLeased         bool            `json:"is_leased"`
	Subusers         []string        `json:"-"`
}

// AddIPAddress adds a new dedicated IP address to the account. subusers, if
// non-empty, are the numeric Subuser IDs to assign the IP to immediately.
func (c *Client) AddIPAddress(ctx context.Context, isAutoWarmup, isParentAssigned bool, subusers []string) (*IPAddress, error) {
	body := map[string]any{
		"is_auto_warmup":     isAutoWarmup,
		"is_parent_assigned": isParentAssigned,
	}
	if len(subusers) > 0 {
		body["subusers"] = subusers
	}
	var out IPAddress
	if err := c.post(ctx, "/v3/send_ips/ips", body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIPAddress retrieves the details of a single dedicated IP address.
func (c *Client) GetIPAddress(ctx context.Context, ip string) (*IPAddressDetail, error) {
	var out IPAddressDetail
	if err := c.get(ctx, "/v3/send_ips/ips/"+esc(ip), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateIPAddress updates mutable properties of a dedicated IP address. Only
// non-nil fields are included in the request.
func (c *Client) UpdateIPAddress(ctx context.Context, ip string, isAutoWarmup, isParentAssigned, isEnabled *bool) (*IPAddressDetail, error) {
	body := map[string]any{}
	if isAutoWarmup != nil {
		body["is_auto_warmup"] = *isAutoWarmup
	}
	if isParentAssigned != nil {
		body["is_parent_assigned"] = *isParentAssigned
	}
	if isEnabled != nil {
		body["is_enabled"] = *isEnabled
	}
	var out IPAddressDetail
	if err := c.patch(ctx, "/v3/send_ips/ips/"+esc(ip), body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListIPAddressSubusers retrieves the Subuser IDs currently assigned to a
// dedicated IP address.
func (c *Client) ListIPAddressSubusers(ctx context.Context, ip string) ([]string, error) {
	var out struct {
		Subusers []string `json:"subusers"`
	}
	if err := c.get(ctx, "/v3/send_ips/ips/"+esc(ip)+"/subusers", &out); err != nil {
		return nil, err
	}
	return out.Subusers, nil
}

// AddSubusersToIPAddress assigns a batch of Subuser IDs to a dedicated IP
// address. All assignments must succeed or the whole request fails.
func (c *Client) AddSubusersToIPAddress(ctx context.Context, ip string, subuserIDs []string) error {
	body := map[string]any{"subusers": subuserIDs}
	return c.post(ctx, "/v3/send_ips/ips/"+esc(ip)+"/subusers:batchAdd", body, nil)
}

// RemoveSubusersFromIPAddress unassigns a batch of Subuser IDs from a
// dedicated IP address.
func (c *Client) RemoveSubusersFromIPAddress(ctx context.Context, ip string, subuserIDs []string) error {
	body := map[string]any{"subusers": subuserIDs}
	return c.post(ctx, "/v3/send_ips/ips/"+esc(ip)+"/subusers:batchDelete", body, nil)
}

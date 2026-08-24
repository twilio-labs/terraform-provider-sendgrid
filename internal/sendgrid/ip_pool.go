package sendgrid

import "context"

// IPPool represents a SendGrid IP pool: a named group of dedicated IPs used
// to segment sending reputation.
type IPPool struct {
	Name string   `json:"name"`
	IPs  []string `json:"ips,omitempty"`
}

// CreateIPPool creates a new, empty IP pool.
func (c *Client) CreateIPPool(ctx context.Context, name string, opts ...RequestOption) (*IPPool, error) {
	body := map[string]any{"name": name}
	var out IPPool
	if err := c.post(ctx, "/v3/ips/pools", body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetIPPool retrieves an IP pool by name, including the IPs currently
// assigned to it.
func (c *Client) GetIPPool(ctx context.Context, name string, opts ...RequestOption) (*IPPool, error) {
	var out struct {
		PoolName string   `json:"pool_name"`
		IPs      []string `json:"ips"`
	}
	if err := c.get(ctx, "/v3/ips/pools/"+esc(name), &out, opts...); err != nil {
		return nil, err
	}
	return &IPPool{Name: out.PoolName, IPs: out.IPs}, nil
}

// UpdateIPPool renames an IP pool.
func (c *Client) UpdateIPPool(ctx context.Context, name, newName string, opts ...RequestOption) (*IPPool, error) {
	body := map[string]any{"name": newName}
	var out IPPool
	if err := c.put(ctx, "/v3/ips/pools/"+esc(name), body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteIPPool deletes an IP pool. The pool must be empty of IPs first.
func (c *Client) DeleteIPPool(ctx context.Context, name string, opts ...RequestOption) error {
	return c.delete(ctx, "/v3/ips/pools/"+esc(name), opts...)
}

// AddIPToPool assigns an IP address to a pool. The IP must already be
// activated for sending on the account.
func (c *Client) AddIPToPool(ctx context.Context, poolName, ip string, opts ...RequestOption) error {
	body := map[string]any{"ip": ip}
	return c.post(ctx, "/v3/ips/pools/"+esc(poolName)+"/ips", body, nil, opts...)
}

// RemoveIPFromPool removes an IP address from a pool.
func (c *Client) RemoveIPFromPool(ctx context.Context, poolName, ip string, opts ...RequestOption) error {
	return c.delete(ctx, "/v3/ips/pools/"+esc(poolName)+"/ips/"+esc(ip), opts...)
}

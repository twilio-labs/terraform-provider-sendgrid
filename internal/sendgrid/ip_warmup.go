package sendgrid

import "context"

// IPWarmup represents the warmup status of a single IP address.
type IPWarmup struct {
	IP        string `json:"ip"`
	StartDate int64  `json:"start_date"`
}

// StartIPWarmup begins warming up an IP address that has already been added
// to the account.
func (c *Client) StartIPWarmup(ctx context.Context, ip string) (*IPWarmup, error) {
	body := map[string]any{"ip": ip}
	var out []IPWarmup
	if err := c.post(ctx, "/v3/ips/warmup", body, &out); err != nil {
		return nil, err
	}
	return firstMatchingIPWarmup(out, ip), nil
}

// GetIPWarmup retrieves the warmup status for a specific IP address. It
// returns (nil, nil) if the IP is not currently in warmup, rather than an
// error, since the API represents "not found" as an empty array with a 200
// status rather than a 404.
func (c *Client) GetIPWarmup(ctx context.Context, ip string) (*IPWarmup, error) {
	var out []IPWarmup
	if err := c.get(ctx, "/v3/ips/warmup/"+esc(ip), &out); err != nil {
		return nil, err
	}
	return firstMatchingIPWarmup(out, ip), nil
}

// StopIPWarmup stops warming up an IP address.
func (c *Client) StopIPWarmup(ctx context.Context, ip string) error {
	return c.delete(ctx, "/v3/ips/warmup/"+esc(ip))
}

// firstMatchingIPWarmup finds the entry for ip in a warmup list response,
// falling back to the sole entry if there's exactly one and it wasn't an
// exact match (defensive: the API's single-IP endpoints are documented to
// return the same array shape as the multi-IP one).
func firstMatchingIPWarmup(list []IPWarmup, ip string) *IPWarmup {
	for i := range list {
		if list[i].IP == ip {
			return &list[i]
		}
	}
	if len(list) == 1 {
		return &list[0]
	}
	return nil
}

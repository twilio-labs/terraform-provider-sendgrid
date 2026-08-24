package sendgrid

import "context"

// Segment represents a Marketing Campaigns contact segment (v2, SQL-based).
type Segment struct {
	ID            string   `json:"id,omitempty"`
	Name          string   `json:"name"`
	QueryDSL      string   `json:"query_dsl"`
	ParentListIDs []string `json:"parent_list_ids,omitempty"`
	ContactsCount int64    `json:"contacts_count,omitempty"`
	CreatedAt     string   `json:"created_at,omitempty"`
	UpdatedAt     string   `json:"updated_at,omitempty"`
}

// CreateSegment creates a new contact segment.
func (c *Client) CreateSegment(ctx context.Context, name, queryDSL string, parentListIDs []string, opts ...RequestOption) (*Segment, error) {
	body := map[string]any{"name": name, "query_dsl": queryDSL}
	if len(parentListIDs) > 0 {
		body["parent_list_ids"] = parentListIDs
	}
	var out Segment
	if err := c.post(ctx, "/v3/marketing/segments/2.0", body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSegment retrieves a single segment by ID.
func (c *Client) GetSegment(ctx context.Context, id string, opts ...RequestOption) (*Segment, error) {
	opts = append(opts, Query("contacts_sample", "false"))
	var out Segment
	if err := c.get(ctx, "/v3/marketing/segments/2.0/"+id, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateSegment updates a segment's name and/or query. parent_list_ids
// cannot be changed after creation.
func (c *Client) UpdateSegment(ctx context.Context, id, name, queryDSL string, opts ...RequestOption) (*Segment, error) {
	body := map[string]any{"name": name, "query_dsl": queryDSL}
	var out Segment
	if err := c.patch(ctx, "/v3/marketing/segments/2.0/"+id, body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteSegment deletes a segment.
func (c *Client) DeleteSegment(ctx context.Context, id string, opts ...RequestOption) error {
	return c.delete(ctx, "/v3/marketing/segments/2.0/"+id, opts...)
}

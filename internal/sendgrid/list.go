package sendgrid

import "context"

// List represents a SendGrid Marketing Campaigns contact list.
type List struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name"`
	ContactCount int64  `json:"contact_count,omitempty"`
}

// CreateList creates a new contact list.
func (c *Client) CreateList(ctx context.Context, name string, opts ...RequestOption) (*List, error) {
	body := map[string]any{"name": name}
	var out List
	if err := c.post(ctx, "/v3/marketing/lists", body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetList retrieves a single contact list by ID.
func (c *Client) GetList(ctx context.Context, id string, opts ...RequestOption) (*List, error) {
	var out List
	if err := c.get(ctx, "/v3/marketing/lists/"+id, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateList renames a contact list.
func (c *Client) UpdateList(ctx context.Context, id, name string, opts ...RequestOption) (*List, error) {
	body := map[string]any{"name": name}
	var out List
	if err := c.patch(ctx, "/v3/marketing/lists/"+id, body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteList deletes a contact list. If deleteContacts is true, contacts on
// the list are also deleted asynchronously; otherwise only the list itself
// is removed and its contacts are left in place.
func (c *Client) DeleteList(ctx context.Context, id string, deleteContacts bool, opts ...RequestOption) error {
	value := "false"
	if deleteContacts {
		value = "true"
	}
	opts = append(opts, Query("delete_contacts", value))
	return c.delete(ctx, "/v3/marketing/lists/"+id, opts...)
}

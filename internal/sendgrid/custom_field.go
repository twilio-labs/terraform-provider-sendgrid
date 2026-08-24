package sendgrid

import "context"

// CustomField represents a Marketing Campaigns custom contact field
// definition.
type CustomField struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name"`
	FieldType string `json:"field_type,omitempty"`
}

type listFieldDefinitionsResponse struct {
	CustomFields []CustomField `json:"custom_fields"`
}

// CreateCustomField creates a new custom field definition. fieldType must be
// "Text", "Number", or "Date".
func (c *Client) CreateCustomField(ctx context.Context, name, fieldType string, opts ...RequestOption) (*CustomField, error) {
	body := map[string]any{"name": name, "field_type": fieldType}
	var out CustomField
	if err := c.post(ctx, "/v3/marketing/field_definitions", body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetCustomField retrieves a single custom field definition by ID. There is
// no single-resource GET endpoint, so this filters the list of all custom
// field definitions. Returns nil, nil if no matching field exists (e.g. it
// was deleted, or the ID belongs to a read-only reserved field).
func (c *Client) GetCustomField(ctx context.Context, id string, opts ...RequestOption) (*CustomField, error) {
	var out listFieldDefinitionsResponse
	if err := c.get(ctx, "/v3/marketing/field_definitions", &out, opts...); err != nil {
		return nil, err
	}
	for _, f := range out.CustomFields {
		if f.ID == id {
			return &f, nil
		}
	}
	return nil, nil
}

// UpdateCustomField renames a custom field definition. The field type
// cannot be changed after creation.
func (c *Client) UpdateCustomField(ctx context.Context, id, name string, opts ...RequestOption) (*CustomField, error) {
	body := map[string]any{"name": name}
	var out CustomField
	if err := c.patch(ctx, "/v3/marketing/field_definitions/"+esc(id), body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteCustomField deletes a custom field definition.
func (c *Client) DeleteCustomField(ctx context.Context, id string, opts ...RequestOption) error {
	return c.delete(ctx, "/v3/marketing/field_definitions/"+esc(id), opts...)
}

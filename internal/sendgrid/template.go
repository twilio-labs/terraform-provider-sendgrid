package sendgrid

import "context"

// Template represents a SendGrid transactional template.
type Template struct {
	ID         string            `json:"id,omitempty"`
	Name       string            `json:"name"`
	Generation string            `json:"generation,omitempty"`
	UpdatedAt  string            `json:"updated_at,omitempty"`
	Versions   []TemplateVersion `json:"versions,omitempty"`
}

// TemplateVersion represents a single version of a transactional template.
type TemplateVersion struct {
	ID                   string `json:"id,omitempty"`
	TemplateID           string `json:"template_id,omitempty"`
	Active               *int   `json:"active,omitempty"`
	Name                 string `json:"name"`
	HTMLContent          string `json:"html_content,omitempty"`
	PlainContent         string `json:"plain_content,omitempty"`
	GeneratePlainContent *bool  `json:"generate_plain_content,omitempty"`
	Subject              string `json:"subject"`
	Editor               string `json:"editor,omitempty"`
	TestData             string `json:"test_data,omitempty"`
	UpdatedAt            string `json:"updated_at,omitempty"`
	ThumbnailURL         string `json:"thumbnail_url,omitempty"`
}

// IsActive reports whether this version is SendGrid's active version for
// its template, centralizing the nil-guarded 0/1-to-bool conversion so
// callers don't each re-derive it (and risk skipping the nil check).
func (v TemplateVersion) IsActive() bool {
	return v.Active != nil && *v.Active == 1
}

// CreateTemplate creates a new transactional template. generation must be
// "legacy" or "dynamic".
func (c *Client) CreateTemplate(ctx context.Context, name, generation string, opts ...RequestOption) (*Template, error) {
	body := map[string]any{"name": name}
	if generation != "" {
		body["generation"] = generation
	}
	var out Template
	if err := c.post(ctx, "/v3/templates", body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTemplate retrieves a single transactional template by ID, including its
// versions.
func (c *Client) GetTemplate(ctx context.Context, id string, opts ...RequestOption) (*Template, error) {
	var out Template
	if err := c.get(ctx, "/v3/templates/"+id, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateTemplate renames a transactional template. The template body itself
// is edited by creating a new TemplateVersion.
func (c *Client) UpdateTemplate(ctx context.Context, id, name string, opts ...RequestOption) (*Template, error) {
	body := map[string]any{"name": name}
	var out Template
	if err := c.patch(ctx, "/v3/templates/"+id, body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteTemplate deletes a transactional template and all of its versions.
func (c *Client) DeleteTemplate(ctx context.Context, id string, opts ...RequestOption) error {
	return c.delete(ctx, "/v3/templates/"+id, opts...)
}

// CreateTemplateVersion creates a new version of a transactional template.
func (c *Client) CreateTemplateVersion(ctx context.Context, templateID string, v TemplateVersion, opts ...RequestOption) (*TemplateVersion, error) {
	var out TemplateVersion
	if err := c.post(ctx, "/v3/templates/"+templateID+"/versions", v, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetTemplateVersion retrieves a single template version.
func (c *Client) GetTemplateVersion(ctx context.Context, templateID, versionID string, opts ...RequestOption) (*TemplateVersion, error) {
	var out TemplateVersion
	if err := c.get(ctx, "/v3/templates/"+templateID+"/versions/"+versionID, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateTemplateVersion edits the content of a template version.
func (c *Client) UpdateTemplateVersion(ctx context.Context, templateID, versionID string, v TemplateVersion, opts ...RequestOption) (*TemplateVersion, error) {
	var out TemplateVersion
	if err := c.patch(ctx, "/v3/templates/"+templateID+"/versions/"+versionID, v, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteTemplateVersion deletes a template version.
func (c *Client) DeleteTemplateVersion(ctx context.Context, templateID, versionID string, opts ...RequestOption) error {
	return c.delete(ctx, "/v3/templates/"+templateID+"/versions/"+versionID, opts...)
}

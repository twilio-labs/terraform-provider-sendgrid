package sendgrid

import "context"

// SingleSendTo specifies the recipients of a Single Send.
type SingleSendTo struct {
	ListIDs    []string `json:"list_ids,omitempty"`
	SegmentIDs []string `json:"segment_ids,omitempty"`
	All        bool     `json:"all,omitempty"`
}

// SingleSendEmailConfig specifies the content of a Single Send.
type SingleSendEmailConfig struct {
	Subject              string `json:"subject,omitempty"`
	HTMLContent          string `json:"html_content,omitempty"`
	PlainContent         string `json:"plain_content,omitempty"`
	GeneratePlainContent bool   `json:"generate_plain_content"`
	DesignID             string `json:"design_id,omitempty"`
	Editor               string `json:"editor,omitempty"`
	SuppressionGroupID   *int64 `json:"suppression_group_id,omitempty"`
	CustomUnsubscribeURL string `json:"custom_unsubscribe_url,omitempty"`
	SenderID             *int64 `json:"sender_id,omitempty"`
	IPPool               string `json:"ip_pool,omitempty"`
}

// SingleSend represents a Marketing Campaigns Single Send.
type SingleSend struct {
	ID          string                `json:"id,omitempty"`
	Name        string                `json:"name"`
	Categories  []string              `json:"categories,omitempty"`
	SendAt      string                `json:"send_at,omitempty"`
	Status      string                `json:"status,omitempty"`
	CreatedAt   string                `json:"created_at,omitempty"`
	UpdatedAt   string                `json:"updated_at,omitempty"`
	SendTo      SingleSendTo          `json:"send_to,omitempty"`
	EmailConfig SingleSendEmailConfig `json:"email_config,omitempty"`
}

// SingleSendParams are the mutable fields of a Single Send, used for both
// create and update.
type SingleSendParams struct {
	Name        string                `json:"name"`
	Categories  []string              `json:"categories,omitempty"`
	SendAt      string                `json:"send_at,omitempty"`
	SendTo      SingleSendTo          `json:"send_to"`
	EmailConfig SingleSendEmailConfig `json:"email_config"`
}

// CreateSingleSend creates a new draft Single Send. It is not scheduled or
// sent until ScheduleSingleSend is also called.
func (c *Client) CreateSingleSend(ctx context.Context, p SingleSendParams, opts ...RequestOption) (*SingleSend, error) {
	var out SingleSend
	if err := c.post(ctx, "/v3/marketing/singlesends", p, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSingleSend retrieves a single Single Send by ID.
func (c *Client) GetSingleSend(ctx context.Context, id string, opts ...RequestOption) (*SingleSend, error) {
	var out SingleSend
	if err := c.get(ctx, "/v3/marketing/singlesends/"+id, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateSingleSend updates a draft Single Send. Single Sends that have
// already been scheduled or sent cannot be updated.
func (c *Client) UpdateSingleSend(ctx context.Context, id string, p SingleSendParams, opts ...RequestOption) (*SingleSend, error) {
	var out SingleSend
	if err := c.patch(ctx, "/v3/marketing/singlesends/"+id, p, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteSingleSend deletes a Single Send.
func (c *Client) DeleteSingleSend(ctx context.Context, id string, opts ...RequestOption) error {
	return c.delete(ctx, "/v3/marketing/singlesends/"+id, opts...)
}

// ScheduleSingleSend schedules a Single Send to be sent at sendAt, or
// immediately if sendAt is "now". Returns the resulting status.
func (c *Client) ScheduleSingleSend(ctx context.Context, id, sendAt string, opts ...RequestOption) (status string, err error) {
	body := map[string]any{"send_at": sendAt}
	var out struct {
		Status string `json:"status"`
	}
	if err := c.put(ctx, "/v3/marketing/singlesends/"+id+"/schedule", body, &out, opts...); err != nil {
		return "", err
	}
	return out.Status, nil
}

// UnscheduleSingleSend reverts a scheduled Single Send back to a draft.
func (c *Client) UnscheduleSingleSend(ctx context.Context, id string, opts ...RequestOption) error {
	return c.delete(ctx, "/v3/marketing/singlesends/"+id+"/schedule", opts...)
}

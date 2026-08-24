package sendgrid

import "context"

// TeammateInvite represents a pending teammate invitation. SendGrid has no
// SCIM API; teammate provisioning is invite-based via the REST Teammates
// API. A person exists only in the pending-invites list, keyed by Token,
// until they accept the invite and become a Teammate with a Username.
type TeammateInvite struct {
	Token          string   `json:"token"`
	Email          string   `json:"email"`
	Scopes         []string `json:"scopes"`
	IsAdmin        bool     `json:"is_admin"`
	ExpirationDate int64    `json:"expiration_date,omitempty"`
}

// Teammate represents an accepted teammate.
type Teammate struct {
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	FirstName string   `json:"first_name,omitempty"`
	LastName  string   `json:"last_name,omitempty"`
	UserType  string   `json:"user_type,omitempty"`
	IsAdmin   bool     `json:"is_admin"`
	Scopes    []string `json:"scopes,omitempty"`
}

// InviteTeammate sends a teammate invitation. The response is the invite
// itself (with a token identifying it), not a created teammate: the
// invitee does not become a queryable teammate until they accept.
func (c *Client) InviteTeammate(ctx context.Context, email string, scopes []string, isAdmin bool, opts ...RequestOption) (*TeammateInvite, error) {
	body := map[string]any{
		"email":    email,
		"scopes":   scopes,
		"is_admin": isAdmin,
	}
	var out TeammateInvite
	if err := c.post(ctx, "/v3/teammates", body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListPendingTeammateInvites retrieves all invitations that have not yet
// been accepted.
func (c *Client) ListPendingTeammateInvites(ctx context.Context, opts ...RequestOption) ([]TeammateInvite, error) {
	var out struct {
		Result []TeammateInvite `json:"result"`
	}
	if err := c.get(ctx, "/v3/teammates/pending", &out, opts...); err != nil {
		return nil, err
	}
	return out.Result, nil
}

// CancelTeammateInvite revokes a pending invitation by its token.
func (c *Client) CancelTeammateInvite(ctx context.Context, token string, opts ...RequestOption) error {
	return c.delete(ctx, "/v3/teammates/pending/"+esc(token), opts...)
}

// ListTeammates retrieves all teammates who have accepted their invite.
func (c *Client) ListTeammates(ctx context.Context, opts ...RequestOption) ([]Teammate, error) {
	var out struct {
		Result []Teammate `json:"result"`
	}
	if err := c.get(ctx, "/v3/teammates", &out, opts...); err != nil {
		return nil, err
	}
	return out.Result, nil
}

// GetTeammate retrieves a single accepted teammate by username.
func (c *Client) GetTeammate(ctx context.Context, username string, opts ...RequestOption) (*Teammate, error) {
	var out Teammate
	if err := c.get(ctx, "/v3/teammates/"+esc(username), &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateTeammate changes the scopes and admin status of an accepted
// teammate. SendGrid requires both fields be sent together.
func (c *Client) UpdateTeammate(ctx context.Context, username string, scopes []string, isAdmin bool, opts ...RequestOption) (*Teammate, error) {
	body := map[string]any{
		"scopes":   scopes,
		"is_admin": isAdmin,
	}
	var out Teammate
	if err := c.patch(ctx, "/v3/teammates/"+esc(username), body, &out, opts...); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteTeammate permanently removes an accepted teammate.
func (c *Client) DeleteTeammate(ctx context.Context, username string, opts ...RequestOption) error {
	return c.delete(ctx, "/v3/teammates/"+esc(username), opts...)
}

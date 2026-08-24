package sendgrid

import (
	"context"
	"net/url"
)

// PartnerAccountProfile holds the account-holder profile supplied when
// provisioning a partner account. SendGrid's API never echoes this data
// back, so it cannot be refreshed by a Read after creation.
type PartnerAccountProfile struct {
	FirstName      string `json:"first_name,omitempty"`
	LastName       string `json:"last_name,omitempty"`
	CompanyName    string `json:"company_name,omitempty"`
	CompanyWebsite string `json:"company_website,omitempty"`
	Email          string `json:"email,omitempty"`
	Phone          string `json:"phone,omitempty"`
	Timezone       string `json:"timezone,omitempty"`
}

// PartnerOffering is a package or add-on offering assignable to a partner
// account (e.g. email infrastructure, Marketing Campaigns, dedicated IPs).
type PartnerOffering struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Quantity int64  `json:"quantity,omitempty"`
}

// PartnerAccount is a provisioned sub-account, as returned by the accounts
// list endpoint.
type PartnerAccount struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
}

type createPartnerAccountRequest struct {
	Profile   *PartnerAccountProfile `json:"profile,omitempty"`
	Offerings []PartnerOffering      `json:"offerings"`
}

type createPartnerAccountResponse struct {
	AccountID string `json:"account_id"`
}

// CreatePartnerAccount provisions a new partner sub-account with the given
// profile and offerings, returning the new account's ID.
func (c *Client) CreatePartnerAccount(ctx context.Context, profile *PartnerAccountProfile, offerings []PartnerOffering, opts ...RequestOption) (string, error) {
	body := createPartnerAccountRequest{Profile: profile, Offerings: offerings}
	var out createPartnerAccountResponse
	if err := c.post(ctx, "/v3/partners/accounts", body, &out, opts...); err != nil {
		return "", err
	}
	return out.AccountID, nil
}

type listPartnerAccountsResponse struct {
	Accounts []PartnerAccount `json:"accounts"`
	Pages    struct {
		Last string `json:"last"`
	} `json:"pages"`
}

// ListPartnerAccounts returns every provisioned partner account, paging
// through the API's cursor-based pagination (offset is the last account ID
// seen, per SendGrid's "AccountProvisioningLastParam").
func (c *Client) ListPartnerAccounts(ctx context.Context, opts ...RequestOption) ([]PartnerAccount, error) {
	const pageSize = 100
	var all []PartnerAccount
	offset := ""
	for {
		path := "/v3/partners/accounts?limit=100"
		if offset != "" {
			path = "/v3/partners/accounts?limit=100&offset=" + url.QueryEscape(offset)
		}
		var out listPartnerAccountsResponse
		if err := c.get(ctx, path, &out, opts...); err != nil {
			return nil, err
		}
		all = append(all, out.Accounts...)
		if len(out.Accounts) < pageSize || out.Pages.Last == "" {
			break
		}
		offset = out.Pages.Last
	}
	return all, nil
}

// DeletePartnerAccount deletes a provisioned partner account.
func (c *Client) DeletePartnerAccount(ctx context.Context, accountID string, opts ...RequestOption) error {
	return c.delete(ctx, "/v3/partners/accounts/"+esc(accountID), opts...)
}

type partnerOfferingsResponse struct {
	Offerings []PartnerOffering `json:"offerings"`
}

// GetPartnerAccountOfferings retrieves the offerings currently assigned to
// a partner account.
func (c *Client) GetPartnerAccountOfferings(ctx context.Context, accountID string, opts ...RequestOption) ([]PartnerOffering, error) {
	var out partnerOfferingsResponse
	if err := c.get(ctx, "/v3/partners/accounts/"+esc(accountID)+"/offerings", &out, opts...); err != nil {
		return nil, err
	}
	return out.Offerings, nil
}

// SetPartnerAccountOfferings replaces a partner account's offerings.
func (c *Client) SetPartnerAccountOfferings(ctx context.Context, accountID string, offerings []PartnerOffering, opts ...RequestOption) error {
	body := partnerOfferingsResponse{Offerings: offerings}
	return c.put(ctx, "/v3/partners/accounts/"+esc(accountID)+"/offerings", body, nil, opts...)
}

type partnerAccountStateResponse struct {
	State string `json:"state"`
}

// GetPartnerAccountState retrieves a partner account's state (e.g.
// "activated", "deactivated", "suspended", "banned", "indeterminate").
func (c *Client) GetPartnerAccountState(ctx context.Context, accountID string, opts ...RequestOption) (string, error) {
	var out partnerAccountStateResponse
	if err := c.get(ctx, "/v3/partners/accounts/"+esc(accountID)+"/state", &out, opts...); err != nil {
		return "", err
	}
	return out.State, nil
}

// SetPartnerAccountState transitions a partner account's state. Only
// "activated" and "deactivated" are settable; the other states SendGrid
// reports via GetPartnerAccountState are not user-settable.
func (c *Client) SetPartnerAccountState(ctx context.Context, accountID, state string, opts ...RequestOption) error {
	body := partnerAccountStateResponse{State: state}
	return c.put(ctx, "/v3/partners/accounts/"+esc(accountID)+"/state", body, nil, opts...)
}

// ListPartnerOfferings returns the catalog of offerings available to
// assign to partner accounts.
func (c *Client) ListPartnerOfferings(ctx context.Context, opts ...RequestOption) ([]PartnerOffering, error) {
	var out partnerOfferingsResponse
	if err := c.get(ctx, "/v3/partners/offerings", &out, opts...); err != nil {
		return nil, err
	}
	return out.Offerings, nil
}

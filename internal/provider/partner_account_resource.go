package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-labs/terraform-provider-sendgrid/internal/sendgrid"
)

var _ resource.Resource = &partnerAccountResource{}

func NewPartnerAccountResource() resource.Resource {
	return &partnerAccountResource{}
}

type partnerAccountResource struct {
	client *sg.Client
}

type partnerAccountProfileModel struct {
	FirstName      types.String `tfsdk:"first_name"`
	LastName       types.String `tfsdk:"last_name"`
	CompanyName    types.String `tfsdk:"company_name"`
	CompanyWebsite types.String `tfsdk:"company_website"`
	Email          types.String `tfsdk:"email"`
	Phone          types.String `tfsdk:"phone"`
	Timezone       types.String `tfsdk:"timezone"`
}

type partnerAccountOfferingModel struct {
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Quantity types.Int64  `tfsdk:"quantity"`
}

type partnerAccountResourceModel struct {
	ID        types.String                  `tfsdk:"id"`
	AccountID types.String                  `tfsdk:"account_id"`
	Profile   *partnerAccountProfileModel   `tfsdk:"profile"`
	Offerings []partnerAccountOfferingModel `tfsdk:"offerings"`
	State     types.String                  `tfsdk:"state"`
	CreatedAt types.String                  `tfsdk:"created_at"`
}

func (r *partnerAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_account"
}

func (r *partnerAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provisions a SendGrid partner sub-account via the Partner Accounts API. This is an " +
			"enterprise feature: the configured provider API key must belong to a SendGrid reseller/partner " +
			"account with partner-provisioning privileges. Calling this against an ordinary account's API key " +
			"will fail with an authorization error.\n\n" +
			"~> `profile` cannot be read back from the API after creation (SendGrid does not expose an " +
			"endpoint to retrieve it), so changing it forces replacement rather than an in-place update.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The provisioned account's ID (identical to `account_id`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_id": schema.StringAttribute{
				Computed:    true,
				Description: "The provisioned account's SendGrid account ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "RFC3339 timestamp of when the account was created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"profile": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "The account holder's profile, supplied only at creation time.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{
					"first_name":      schema.StringAttribute{Optional: true, Description: "The account holder's first name."},
					"last_name":       schema.StringAttribute{Optional: true, Description: "The account holder's last name."},
					"company_name":    schema.StringAttribute{Optional: true, Description: "The account holder's company name."},
					"company_website": schema.StringAttribute{Optional: true, Description: "The account holder's company website."},
					"email":           schema.StringAttribute{Optional: true, Description: "The account holder's email address."},
					"phone":           schema.StringAttribute{Optional: true, Description: "The account holder's phone number, in E.164 format."},
					"timezone":        schema.StringAttribute{Optional: true, Description: "The account holder's IANA time zone (e.g. `Asia/Tokyo`)."},
				},
			},
			"offerings": schema.ListNestedAttribute{
				Required:    true,
				Description: "The package and add-on offerings assigned to this account. An account may have exactly one `package` offering.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The offering's name (e.g. `org.ei.free.v1`).",
						},
						"type": schema.StringAttribute{
							Required:    true,
							Description: "The offering type: `package` or `addon`.",
							Validators: []validator.String{
								stringvalidator.OneOf("package", "addon"),
							},
						},
						"quantity": schema.Int64Attribute{
							Optional:    true,
							Computed:    true,
							Description: "The quantity of this offering. Must be `1` for `package` offerings.",
						},
					},
				},
			},
			"state": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "The account's state. Settable values are `activated` and `deactivated`; SendGrid " +
					"may independently transition an account to `suspended`, `banned`, or `indeterminate`, which " +
					"this attribute will reflect on the next read but which cannot be set here.",
				Validators: []validator.String{
					stringvalidator.OneOf("activated", "deactivated"),
				},
			},
		},
	}
}

func (r *partnerAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func offeringsFromModel(m []partnerAccountOfferingModel) []sg.PartnerOffering {
	out := make([]sg.PartnerOffering, 0, len(m))
	for _, o := range m {
		out = append(out, sg.PartnerOffering{
			Name:     o.Name.ValueString(),
			Type:     o.Type.ValueString(),
			Quantity: o.Quantity.ValueInt64(),
		})
	}
	return out
}

func flattenOfferings(offerings []sg.PartnerOffering) []partnerAccountOfferingModel {
	out := make([]partnerAccountOfferingModel, 0, len(offerings))
	for _, o := range offerings {
		out = append(out, partnerAccountOfferingModel{
			Name:     types.StringValue(o.Name),
			Type:     types.StringValue(o.Type),
			Quantity: types.Int64Value(o.Quantity),
		})
	}
	return out
}

func (r *partnerAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan partnerAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var profile *sg.PartnerAccountProfile
	if plan.Profile != nil {
		profile = &sg.PartnerAccountProfile{
			FirstName:      plan.Profile.FirstName.ValueString(),
			LastName:       plan.Profile.LastName.ValueString(),
			CompanyName:    plan.Profile.CompanyName.ValueString(),
			CompanyWebsite: plan.Profile.CompanyWebsite.ValueString(),
			Email:          plan.Profile.Email.ValueString(),
			Phone:          plan.Profile.Phone.ValueString(),
			Timezone:       plan.Profile.Timezone.ValueString(),
		}
	}

	accountID, err := r.client.CreatePartnerAccount(ctx, profile, offeringsFromModel(plan.Offerings))
	if err != nil {
		resp.Diagnostics.AddError("Error creating partner account", err.Error())
		return
	}

	// The account now exists remotely with its offerings already assigned
	// by the create call. Persist it before attempting the optional state
	// transition below, so a failure there doesn't leave an orphaned,
	// untracked account.
	plan.ID = types.StringValue(accountID)
	plan.AccountID = types.StringValue(accountID)
	plan.CreatedAt = types.StringValue("")
	plan.State = types.StringUnknown()
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configuredState := plan.State.ValueString()
	if !plan.State.IsNull() && !plan.State.IsUnknown() && configuredState != "" {
		if err := r.client.SetPartnerAccountState(ctx, accountID, configuredState); err != nil {
			resp.Diagnostics.AddError("Error setting partner account state", err.Error())
			return
		}
	}

	accounts, err := r.client.ListPartnerAccounts(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading partner account", err.Error())
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}
	for i := range accounts {
		if accounts[i].ID == accountID {
			plan.CreatedAt = types.StringValue(accounts[i].CreatedAt)
			break
		}
	}

	partnerState, err := r.client.GetPartnerAccountState(ctx, accountID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading partner account state", err.Error())
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}
	plan.State = types.StringValue(partnerState)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *partnerAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state partnerAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID := state.ID.ValueString()

	accounts, err := r.client.ListPartnerAccounts(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error listing partner accounts", err.Error())
		return
	}
	var found *sg.PartnerAccount
	for i := range accounts {
		if accounts[i].ID == accountID {
			found = &accounts[i]
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.CreatedAt = types.StringValue(found.CreatedAt)

	offerings, err := r.client.GetPartnerAccountOfferings(ctx, accountID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading partner account offerings", err.Error())
		return
	}
	state.Offerings = flattenOfferings(offerings)

	partnerState, err := r.client.GetPartnerAccountState(ctx, accountID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading partner account state", err.Error())
		return
	}
	state.State = types.StringValue(partnerState)

	// profile is intentionally left untouched: the API cannot return it.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *partnerAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state partnerAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID := state.ID.ValueString()
	plan.ID = state.ID
	plan.AccountID = state.AccountID
	plan.CreatedAt = state.CreatedAt

	if !offeringsEqual(plan.Offerings, state.Offerings) {
		if err := r.client.SetPartnerAccountOfferings(ctx, accountID, offeringsFromModel(plan.Offerings)); err != nil {
			resp.Diagnostics.AddError("Error updating partner account offerings", err.Error())
			plan.Offerings = state.Offerings
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.State.ValueString() != state.State.ValueString() && !plan.State.IsUnknown() {
		if err := r.client.SetPartnerAccountState(ctx, accountID, plan.State.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error updating partner account state", err.Error())
			plan.State = state.State
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func offeringsEqual(a, b []partnerAccountOfferingModel) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Name.ValueString() != b[i].Name.ValueString() ||
			a[i].Type.ValueString() != b[i].Type.ValueString() ||
			a[i].Quantity.ValueInt64() != b[i].Quantity.ValueInt64() {
			return false
		}
	}
	return true
}

func (r *partnerAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state partnerAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeletePartnerAccount(ctx, state.ID.ValueString()); err != nil && !sg.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting partner account", err.Error())
	}
}

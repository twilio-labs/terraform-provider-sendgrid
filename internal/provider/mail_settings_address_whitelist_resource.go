package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

var _ resource.Resource = &addressWhitelistMailSettingsResource{}

func NewAddressWhitelistMailSettingsResource() resource.Resource {
	return &addressWhitelistMailSettingsResource{}
}

type addressWhitelistMailSettingsResource struct {
	client *sg.Client
}

type addressWhitelistMailSettingsResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
	List    types.Set    `tfsdk:"list"`
}

func (r *addressWhitelistMailSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_address_whitelist_mail_settings"
}

func (r *addressWhitelistMailSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the account-wide address whitelist mail setting: mail sent to addresses or " +
			"domains on this list bypasses spam filtering. This is a singleton: SendGrid has exactly one " +
			"address whitelist configuration per account (or per subuser, via the provider's `on_behalf_of`); " +
			"creating this resource updates that configuration in place, and destroying it only removes it " +
			"from Terraform state without resetting SendGrid's setting.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always `address_whitelist_mail_settings`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the address whitelist mail setting is enabled.",
			},
			"list": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Email addresses and/or domains to whitelist. All email addresses belonging to " +
					"a whitelisted domain bypass spam filtering.",
			},
		},
	}
}

func (r *addressWhitelistMailSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *addressWhitelistMailSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan addressWhitelistMailSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, diags := stringSetToSlice(ctx, plan.List)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateAddressWhitelistMailSettings(ctx, sg.AddressWhitelistMailSettings{
		Enabled: plan.Enabled.ValueBool(),
		List:    list,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating address whitelist mail settings", err.Error())
		return
	}

	plan.ID = types.StringValue("address_whitelist_mail_settings")
	plan.Enabled = types.BoolValue(s.Enabled)
	resp.Diagnostics.Append(setStringSet(ctx, &plan.List, s.List)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *addressWhitelistMailSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state addressWhitelistMailSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetAddressWhitelistMailSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading address whitelist mail settings", err.Error())
		return
	}

	state.ID = types.StringValue("address_whitelist_mail_settings")
	state.Enabled = types.BoolValue(s.Enabled)
	resp.Diagnostics.Append(setStringSet(ctx, &state.List, s.List)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *addressWhitelistMailSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan addressWhitelistMailSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, diags := stringSetToSlice(ctx, plan.List)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateAddressWhitelistMailSettings(ctx, sg.AddressWhitelistMailSettings{
		Enabled: plan.Enabled.ValueBool(),
		List:    list,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating address whitelist mail settings", err.Error())
		return
	}

	plan.ID = types.StringValue("address_whitelist_mail_settings")
	plan.Enabled = types.BoolValue(s.Enabled)
	resp.Diagnostics.Append(setStringSet(ctx, &plan.List, s.List)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *addressWhitelistMailSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No API action: SendGrid has no concept of "unset" for this setting.
	// Deleting this resource only removes it from Terraform state.
}

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-labs/terraform-provider-sendgrid/internal/sendgrid"
)

var _ resource.Resource = &forwardSpamMailSettingsResource{}

func NewForwardSpamMailSettingsResource() resource.Resource {
	return &forwardSpamMailSettingsResource{}
}

type forwardSpamMailSettingsResource struct {
	client *sg.Client
}

type forwardSpamMailSettingsResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Email   types.String `tfsdk:"email"`
}

func (r *forwardSpamMailSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_forward_spam_mail_settings"
}

func (r *forwardSpamMailSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the account-wide spam forwarding mail setting, which forwards spam reports to " +
			"a specified email address. This is a singleton: SendGrid has exactly one spam forwarding " +
			"configuration per account (or per subuser, via the provider's `on_behalf_of`); creating this " +
			"resource updates that configuration in place, and destroying it only removes it from Terraform " +
			"state without resetting SendGrid's setting.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always `forward_spam_mail_settings`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the spam forwarding mail setting is enabled.",
			},
			"email": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The email address that spam reports are forwarded to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *forwardSpamMailSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func forwardSpamMailSettingsFromModel(m forwardSpamMailSettingsResourceModel) sg.ForwardSpamMailSettings {
	return sg.ForwardSpamMailSettings{
		Enabled: m.Enabled.ValueBool(),
		Email:   m.Email.ValueString(),
	}
}

func flattenForwardSpamMailSettings(s *sg.ForwardSpamMailSettings, m *forwardSpamMailSettingsResourceModel) {
	m.ID = types.StringValue("forward_spam_mail_settings")
	m.Enabled = types.BoolValue(s.Enabled)
	m.Email = types.StringValue(s.Email)
}

func (r *forwardSpamMailSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan forwardSpamMailSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateForwardSpamMailSettings(ctx, forwardSpamMailSettingsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating spam forwarding mail settings", err.Error())
		return
	}

	flattenForwardSpamMailSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *forwardSpamMailSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state forwardSpamMailSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetForwardSpamMailSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading spam forwarding mail settings", err.Error())
		return
	}

	flattenForwardSpamMailSettings(s, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *forwardSpamMailSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan forwardSpamMailSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateForwardSpamMailSettings(ctx, forwardSpamMailSettingsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating spam forwarding mail settings", err.Error())
		return
	}

	flattenForwardSpamMailSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *forwardSpamMailSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No API action: SendGrid has no concept of "unset" for this setting.
	// Deleting this resource only removes it from Terraform state.
}

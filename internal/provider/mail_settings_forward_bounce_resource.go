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

var _ resource.Resource = &forwardBounceMailSettingsResource{}

func NewForwardBounceMailSettingsResource() resource.Resource {
	return &forwardBounceMailSettingsResource{}
}

type forwardBounceMailSettingsResource struct {
	client *sg.Client
}

type forwardBounceMailSettingsResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
	Email   types.String `tfsdk:"email"`
}

func (r *forwardBounceMailSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_forward_bounce_mail_settings"
}

func (r *forwardBounceMailSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the account-wide bounce forwarding mail setting, which forwards bounce " +
			"reports to a specified email address. This is a singleton: SendGrid has exactly one bounce " +
			"forwarding configuration per account (or per subuser, via the provider's `on_behalf_of`); " +
			"creating this resource updates that configuration in place, and destroying it only removes it " +
			"from Terraform state without resetting SendGrid's setting.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always `forward_bounce_mail_settings`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the bounce forwarding mail setting is enabled.",
			},
			"email": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The email address that bounce reports are forwarded to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *forwardBounceMailSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func forwardBounceMailSettingsFromModel(m forwardBounceMailSettingsResourceModel) sg.ForwardBounceMailSettings {
	s := sg.ForwardBounceMailSettings{Enabled: m.Enabled.ValueBool()}
	if !m.Email.IsNull() && !m.Email.IsUnknown() {
		v := m.Email.ValueString()
		s.Email = &v
	}
	return s
}

func flattenForwardBounceMailSettings(s *sg.ForwardBounceMailSettings, m *forwardBounceMailSettingsResourceModel) {
	m.ID = types.StringValue("forward_bounce_mail_settings")
	m.Enabled = types.BoolValue(s.Enabled)
	if s.Email != nil {
		m.Email = types.StringValue(*s.Email)
	} else {
		m.Email = types.StringNull()
	}
}

func (r *forwardBounceMailSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan forwardBounceMailSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateForwardBounceMailSettings(ctx, forwardBounceMailSettingsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating bounce forwarding mail settings", err.Error())
		return
	}

	flattenForwardBounceMailSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *forwardBounceMailSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state forwardBounceMailSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetForwardBounceMailSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading bounce forwarding mail settings", err.Error())
		return
	}

	flattenForwardBounceMailSettings(s, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *forwardBounceMailSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan forwardBounceMailSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateForwardBounceMailSettings(ctx, forwardBounceMailSettingsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating bounce forwarding mail settings", err.Error())
		return
	}

	flattenForwardBounceMailSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *forwardBounceMailSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No API action: SendGrid has no concept of "unset" for this setting.
	// Deleting this resource only removes it from Terraform state.
}

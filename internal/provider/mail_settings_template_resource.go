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

var _ resource.Resource = &templateMailSettingsResource{}

func NewTemplateMailSettingsResource() resource.Resource {
	return &templateMailSettingsResource{}
}

type templateMailSettingsResource struct {
	client *sg.Client
}

type templateMailSettingsResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	HTMLContent types.String `tfsdk:"html_content"`
}

func (r *templateMailSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template_mail_settings"
}

func (r *templateMailSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the account-wide legacy email template mail setting, which wraps every " +
			"outgoing email in custom HTML. This is a singleton: SendGrid has exactly one legacy template " +
			"configuration per account (or per subuser, via the provider's `on_behalf_of`); creating this " +
			"resource updates that configuration in place, and destroying it only removes it from Terraform " +
			"state without resetting SendGrid's setting.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always `template_mail_settings`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the legacy email template mail setting is enabled.",
			},
			"html_content": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "The custom HTML content for the legacy email template. Must contain the " +
					"`<% body %>` tag, or the original email body will not be included.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *templateMailSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func templateMailSettingsFromModel(m templateMailSettingsResourceModel) sg.TemplateMailSettings {
	return sg.TemplateMailSettings{
		Enabled:     m.Enabled.ValueBool(),
		HTMLContent: m.HTMLContent.ValueString(),
	}
}

func flattenTemplateMailSettings(s *sg.TemplateMailSettings, m *templateMailSettingsResourceModel) {
	m.ID = types.StringValue("template_mail_settings")
	m.Enabled = types.BoolValue(s.Enabled)
	m.HTMLContent = types.StringValue(s.HTMLContent)
}

func (r *templateMailSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan templateMailSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateTemplateMailSettings(ctx, templateMailSettingsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating legacy template mail settings", err.Error())
		return
	}

	flattenTemplateMailSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateMailSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state templateMailSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetTemplateMailSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading legacy template mail settings", err.Error())
		return
	}

	flattenTemplateMailSettings(s, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *templateMailSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan templateMailSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateTemplateMailSettings(ctx, templateMailSettingsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating legacy template mail settings", err.Error())
		return
	}

	flattenTemplateMailSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateMailSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No API action: SendGrid has no concept of "unset" for this setting.
	// Deleting this resource only removes it from Terraform state.
}

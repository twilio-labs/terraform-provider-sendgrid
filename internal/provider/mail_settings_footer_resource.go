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

var _ resource.Resource = &footerMailSettingsResource{}

func NewFooterMailSettingsResource() resource.Resource {
	return &footerMailSettingsResource{}
}

type footerMailSettingsResource struct {
	client *sg.Client
}

type footerMailSettingsResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	HTMLContent  types.String `tfsdk:"html_content"`
	PlainContent types.String `tfsdk:"plain_content"`
}

func (r *footerMailSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_footer_mail_settings"
}

func (r *footerMailSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the account-wide footer mail setting, which appends custom content to the " +
			"bottom of every email. This is a singleton: SendGrid has exactly one footer configuration per " +
			"account (or per subuser, via the provider's `on_behalf_of`); creating this resource updates that " +
			"configuration in place, and destroying it only removes it from Terraform state without resetting " +
			"SendGrid's setting.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always `footer_mail_settings`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the footer mail setting is enabled.",
			},
			"html_content": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The custom HTML content appended to the bottom of HTML emails.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"plain_content": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The plain text content appended to the bottom of plain-text emails.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *footerMailSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func footerMailSettingsFromModel(m footerMailSettingsResourceModel) sg.FooterMailSettings {
	return sg.FooterMailSettings{
		Enabled:      m.Enabled.ValueBool(),
		HTMLContent:  m.HTMLContent.ValueString(),
		PlainContent: m.PlainContent.ValueString(),
	}
}

func flattenFooterMailSettings(s *sg.FooterMailSettings, m *footerMailSettingsResourceModel) {
	m.ID = types.StringValue("footer_mail_settings")
	m.Enabled = types.BoolValue(s.Enabled)
	m.HTMLContent = types.StringValue(s.HTMLContent)
	m.PlainContent = types.StringValue(s.PlainContent)
}

func (r *footerMailSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan footerMailSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateFooterMailSettings(ctx, footerMailSettingsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating footer mail settings", err.Error())
		return
	}

	flattenFooterMailSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *footerMailSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state footerMailSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetFooterMailSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading footer mail settings", err.Error())
		return
	}

	flattenFooterMailSettings(s, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *footerMailSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan footerMailSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateFooterMailSettings(ctx, footerMailSettingsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating footer mail settings", err.Error())
		return
	}

	flattenFooterMailSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *footerMailSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No API action: SendGrid has no concept of "unset" for this setting.
	// Deleting this resource only removes it from Terraform state.
}

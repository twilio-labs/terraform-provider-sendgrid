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

var _ resource.Resource = &subscriptionTrackingSettingsResource{}

func NewSubscriptionTrackingSettingsResource() resource.Resource {
	return &subscriptionTrackingSettingsResource{}
}

type subscriptionTrackingSettingsResource struct {
	client *sg.Client
}

type subscriptionTrackingSettingsResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Enabled      types.Bool   `tfsdk:"enabled"`
	HTMLContent  types.String `tfsdk:"html_content"`
	Landing      types.String `tfsdk:"landing"`
	PlainContent types.String `tfsdk:"plain_content"`
	Replace      types.String `tfsdk:"replace"`
	URL          types.String `tfsdk:"url"`
}

func (r *subscriptionTrackingSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription_tracking_settings"
}

func (r *subscriptionTrackingSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the account-wide subscription tracking setting, which adds a " +
			"subscribe/unsubscribe link to the bottom of emails. This is a singleton: SendGrid has exactly " +
			"one subscription tracking configuration per account (or per subuser, via the provider's " +
			"`on_behalf_of`); creating this resource updates that configuration in place, and destroying it " +
			"only removes it from Terraform state without resetting SendGrid's setting.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always `subscription_tracking_settings`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether subscription tracking is enabled.",
			},
			"html_content": schema.StringAttribute{
				Optional:    true,
				Description: "The HTML for the unsubscribe link, appended to the bottom of HTML emails.",
			},
			"landing": schema.StringAttribute{
				Optional:    true,
				Description: "The HTML displayed on SendGrid's hosted landing page after a recipient clicks unsubscribe.",
			},
			"plain_content": schema.StringAttribute{
				Optional: true,
				Description: "The plain-text unsubscribe content, appended to the bottom of plain-text emails. " +
					"Must contain the `<% %>` tag, or recipients will have no URL to unsubscribe from.",
			},
			"replace": schema.StringAttribute{
				Optional:    true,
				Description: "A custom replacement tag to place unsubscribe content anywhere in an email template.",
			},
			"url": schema.StringAttribute{
				Optional:    true,
				Description: "The URL where unsubscribing recipients are sent, if not using SendGrid's hosted landing page.",
			},
		},
	}
}

func (r *subscriptionTrackingSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func subscriptionTrackingSettingsFromModel(m subscriptionTrackingSettingsResourceModel) sg.SubscriptionTrackingSettings {
	return sg.SubscriptionTrackingSettings{
		Enabled:      m.Enabled.ValueBool(),
		HTMLContent:  m.HTMLContent.ValueString(),
		Landing:      m.Landing.ValueString(),
		PlainContent: m.PlainContent.ValueString(),
		Replace:      m.Replace.ValueString(),
		URL:          m.URL.ValueString(),
	}
}

func flattenSubscriptionTrackingSettings(s *sg.SubscriptionTrackingSettings, m *subscriptionTrackingSettingsResourceModel) {
	m.ID = types.StringValue("subscription_tracking_settings")
	m.Enabled = types.BoolValue(s.Enabled)
	m.HTMLContent = types.StringValue(s.HTMLContent)
	m.Landing = types.StringValue(s.Landing)
	m.PlainContent = types.StringValue(s.PlainContent)
	m.Replace = types.StringValue(s.Replace)
	m.URL = types.StringValue(s.URL)
}

func (r *subscriptionTrackingSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subscriptionTrackingSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateSubscriptionTrackingSettings(ctx, subscriptionTrackingSettingsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating subscription tracking settings", err.Error())
		return
	}

	flattenSubscriptionTrackingSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subscriptionTrackingSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subscriptionTrackingSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetSubscriptionTrackingSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading subscription tracking settings", err.Error())
		return
	}

	flattenSubscriptionTrackingSettings(s, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *subscriptionTrackingSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan subscriptionTrackingSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateSubscriptionTrackingSettings(ctx, subscriptionTrackingSettingsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating subscription tracking settings", err.Error())
		return
	}

	flattenSubscriptionTrackingSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subscriptionTrackingSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No API action: SendGrid has no concept of "unset" for this setting.
}

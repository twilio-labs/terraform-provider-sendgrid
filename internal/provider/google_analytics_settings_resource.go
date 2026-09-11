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

var _ resource.Resource = &googleAnalyticsSettingsResource{}

func NewGoogleAnalyticsSettingsResource() resource.Resource {
	return &googleAnalyticsSettingsResource{}
}

type googleAnalyticsSettingsResource struct {
	client *sg.Client
}

type googleAnalyticsSettingsResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	UTMCampaign types.String `tfsdk:"utm_campaign"`
	UTMContent  types.String `tfsdk:"utm_content"`
	UTMMedium   types.String `tfsdk:"utm_medium"`
	UTMSource   types.String `tfsdk:"utm_source"`
	UTMTerm     types.String `tfsdk:"utm_term"`
}

func (r *googleAnalyticsSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_google_analytics_settings"
}

func (r *googleAnalyticsSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the account-wide Google Analytics tracking setting, which appends UTM " +
			"parameters to links in emails. This is a singleton: SendGrid has exactly one Google Analytics " +
			"configuration per account (or per subuser, via the provider's `on_behalf_of`); creating this " +
			"resource updates that configuration in place, and destroying it only removes it from Terraform " +
			"state without resetting SendGrid's setting.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always `google_analytics_settings`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether Google Analytics tracking is enabled.",
			},
			"utm_campaign": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the campaign.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"utm_content": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Used to differentiate ads.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"utm_medium": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The marketing medium, e.g. \"email\".",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"utm_source": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The referrer source.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"utm_term": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Paid keywords.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *googleAnalyticsSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func googleAnalyticsSettingsFromModel(m googleAnalyticsSettingsResourceModel) sg.GoogleAnalyticsSettings {
	return sg.GoogleAnalyticsSettings{
		Enabled:     m.Enabled.ValueBool(),
		UTMCampaign: m.UTMCampaign.ValueString(),
		UTMContent:  m.UTMContent.ValueString(),
		UTMMedium:   m.UTMMedium.ValueString(),
		UTMSource:   m.UTMSource.ValueString(),
		UTMTerm:     m.UTMTerm.ValueString(),
	}
}

func flattenGoogleAnalyticsSettings(s *sg.GoogleAnalyticsSettings, m *googleAnalyticsSettingsResourceModel) {
	m.ID = types.StringValue("google_analytics_settings")
	m.Enabled = types.BoolValue(s.Enabled)
	m.UTMCampaign = types.StringValue(s.UTMCampaign)
	m.UTMContent = types.StringValue(s.UTMContent)
	m.UTMMedium = types.StringValue(s.UTMMedium)
	m.UTMSource = types.StringValue(s.UTMSource)
	m.UTMTerm = types.StringValue(s.UTMTerm)
}

func (r *googleAnalyticsSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan googleAnalyticsSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateGoogleAnalyticsSettings(ctx, googleAnalyticsSettingsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating Google Analytics settings", err.Error())
		return
	}

	flattenGoogleAnalyticsSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *googleAnalyticsSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state googleAnalyticsSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetGoogleAnalyticsSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading Google Analytics settings", err.Error())
		return
	}

	flattenGoogleAnalyticsSettings(s, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *googleAnalyticsSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan googleAnalyticsSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateGoogleAnalyticsSettings(ctx, googleAnalyticsSettingsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating Google Analytics settings", err.Error())
		return
	}

	flattenGoogleAnalyticsSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *googleAnalyticsSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No API action: SendGrid has no concept of "unset" for this setting.
}

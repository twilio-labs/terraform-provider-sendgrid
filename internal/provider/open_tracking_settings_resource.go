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

var _ resource.Resource = &openTrackingSettingsResource{}

func NewOpenTrackingSettingsResource() resource.Resource {
	return &openTrackingSettingsResource{}
}

type openTrackingSettingsResource struct {
	client *sg.Client
}

type openTrackingSettingsResourceModel struct {
	ID      types.String `tfsdk:"id"`
	Enabled types.Bool   `tfsdk:"enabled"`
}

func (r *openTrackingSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_open_tracking_settings"
}

func (r *openTrackingSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the account-wide open tracking setting. This is a singleton: SendGrid has " +
			"exactly one open tracking configuration per account (or per subuser, via the provider's " +
			"`on_behalf_of`); creating this resource updates that configuration in place, and destroying it " +
			"only removes it from Terraform state without resetting SendGrid's setting.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always `open_tracking_settings`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether open tracking is enabled.",
			},
		},
	}
}

func (r *openTrackingSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func flattenOpenTrackingSettings(s *sg.OpenTrackingSettings, m *openTrackingSettingsResourceModel) {
	m.ID = types.StringValue("open_tracking_settings")
	m.Enabled = types.BoolValue(s.Enabled)
}

func (r *openTrackingSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan openTrackingSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateOpenTrackingSettings(ctx, plan.Enabled.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Error updating open tracking settings", err.Error())
		return
	}

	flattenOpenTrackingSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *openTrackingSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state openTrackingSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetOpenTrackingSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading open tracking settings", err.Error())
		return
	}

	flattenOpenTrackingSettings(s, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *openTrackingSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan openTrackingSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateOpenTrackingSettings(ctx, plan.Enabled.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Error updating open tracking settings", err.Error())
		return
	}

	flattenOpenTrackingSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *openTrackingSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No API action: SendGrid has no concept of "unset" for this setting.
}

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

var _ resource.Resource = &clickTrackingSettingsResource{}

func NewClickTrackingSettingsResource() resource.Resource {
	return &clickTrackingSettingsResource{}
}

type clickTrackingSettingsResource struct {
	client *sg.Client
}

type clickTrackingSettingsResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	EnableText types.Bool   `tfsdk:"enable_text"`
}

func (r *clickTrackingSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_click_tracking_settings"
}

func (r *clickTrackingSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the account-wide click tracking setting. This is a singleton: SendGrid has " +
			"exactly one click tracking configuration per account (or per subuser, via the provider's " +
			"`on_behalf_of`); creating this resource updates that configuration in place, and destroying it " +
			"only removes it from Terraform state without resetting SendGrid's setting.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always `click_tracking_settings`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether click tracking is enabled.",
			},
			"enable_text": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether click tracking is enabled for plain text emails. Read-only: SendGrid derives this from `enabled`.",
			},
		},
	}
}

func (r *clickTrackingSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func flattenClickTrackingSettings(s *sg.ClickTrackingSettings, m *clickTrackingSettingsResourceModel) {
	m.ID = types.StringValue("click_tracking_settings")
	m.Enabled = types.BoolValue(s.Enabled)
	m.EnableText = types.BoolValue(s.EnableText)
}

func (r *clickTrackingSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clickTrackingSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateClickTrackingSettings(ctx, plan.Enabled.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Error updating click tracking settings", err.Error())
		return
	}

	flattenClickTrackingSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clickTrackingSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clickTrackingSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetClickTrackingSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading click tracking settings", err.Error())
		return
	}

	flattenClickTrackingSettings(s, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *clickTrackingSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan clickTrackingSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateClickTrackingSettings(ctx, plan.Enabled.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Error updating click tracking settings", err.Error())
		return
	}

	flattenClickTrackingSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clickTrackingSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No API action: SendGrid has no concept of "unset" for this setting.
	// Deleting this resource only removes it from Terraform state.
}

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-labs/terraform-provider-sendgrid/internal/sendgrid"
)

var _ resource.Resource = &bouncePurgeMailSettingsResource{}

func NewBouncePurgeMailSettingsResource() resource.Resource {
	return &bouncePurgeMailSettingsResource{}
}

type bouncePurgeMailSettingsResource struct {
	client *sg.Client
}

type bouncePurgeMailSettingsResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	SoftBounces types.Int64  `tfsdk:"soft_bounces"`
	HardBounces types.Int64  `tfsdk:"hard_bounces"`
}

func (r *bouncePurgeMailSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bounce_purge_mail_settings"
}

func (r *bouncePurgeMailSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the account-wide bounce purge mail setting, which controls automatic purging " +
			"of contacts from bounce suppression lists after a number of days. This is a singleton: SendGrid " +
			"has exactly one bounce purge configuration per account (or per subuser, via the provider's " +
			"`on_behalf_of`); creating this resource updates that configuration in place, and destroying it " +
			"only removes it from Terraform state without resetting SendGrid's setting.\n\n" +
			"~> SendGrid rejects `enabled = false` unless `soft_bounces` and `hard_bounces` are both left " +
			"unset — confirmed against a live account (400: \"cannot be disabled unless both hard_bounces " +
			"and soft_bounces are empty\"). Set both or neither.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always `bounce_purge_mail_settings`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the bounce purge mail setting is enabled.",
			},
			"soft_bounces": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Description: "The number of days after which SendGrid purges contacts from the soft bounces " +
					"suppression list.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"hard_bounces": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Description: "The number of days after which SendGrid purges contacts from the hard bounces " +
					"suppression list.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *bouncePurgeMailSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func bouncePurgeMailSettingsFromModel(m bouncePurgeMailSettingsResourceModel) sg.BouncePurgeMailSettings {
	s := sg.BouncePurgeMailSettings{Enabled: m.Enabled.ValueBool()}
	if !m.SoftBounces.IsNull() && !m.SoftBounces.IsUnknown() {
		v := int(m.SoftBounces.ValueInt64())
		s.SoftBounces = &v
	}
	if !m.HardBounces.IsNull() && !m.HardBounces.IsUnknown() {
		v := int(m.HardBounces.ValueInt64())
		s.HardBounces = &v
	}
	return s
}

func flattenBouncePurgeMailSettings(s *sg.BouncePurgeMailSettings, m *bouncePurgeMailSettingsResourceModel) {
	m.ID = types.StringValue("bounce_purge_mail_settings")
	m.Enabled = types.BoolValue(s.Enabled)
	if s.SoftBounces != nil {
		m.SoftBounces = types.Int64Value(int64(*s.SoftBounces))
	} else {
		m.SoftBounces = types.Int64Null()
	}
	if s.HardBounces != nil {
		m.HardBounces = types.Int64Value(int64(*s.HardBounces))
	} else {
		m.HardBounces = types.Int64Null()
	}
}

func (r *bouncePurgeMailSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bouncePurgeMailSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateBouncePurgeMailSettings(ctx, bouncePurgeMailSettingsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating bounce purge mail settings", err.Error())
		return
	}

	flattenBouncePurgeMailSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bouncePurgeMailSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bouncePurgeMailSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetBouncePurgeMailSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading bounce purge mail settings", err.Error())
		return
	}

	flattenBouncePurgeMailSettings(s, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bouncePurgeMailSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bouncePurgeMailSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateBouncePurgeMailSettings(ctx, bouncePurgeMailSettingsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating bounce purge mail settings", err.Error())
		return
	}

	flattenBouncePurgeMailSettings(s, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *bouncePurgeMailSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No API action: SendGrid has no concept of "unset" for this setting.
	// Deleting this resource only removes it from Terraform state.
}

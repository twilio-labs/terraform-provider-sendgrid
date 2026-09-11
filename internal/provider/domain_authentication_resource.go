package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-labs/terraform-provider-sendgrid/internal/sendgrid"
)

var (
	_ resource.Resource                = &domainAuthenticationResource{}
	_ resource.ResourceWithImportState = &domainAuthenticationResource{}
)

func NewDomainAuthenticationResource() resource.Resource {
	return &domainAuthenticationResource{}
}

type domainAuthenticationResource struct {
	client *sg.Client
}

type domainAuthenticationResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Domain             types.String `tfsdk:"domain"`
	Subdomain          types.String `tfsdk:"subdomain"`
	Username           types.String `tfsdk:"username"`
	IPs                types.List   `tfsdk:"ips"`
	CustomSPF          types.Bool   `tfsdk:"custom_spf"`
	Default            types.Bool   `tfsdk:"default"`
	AutomaticSecurity  types.Bool   `tfsdk:"automatic_security"`
	CustomDKIMSelector types.String `tfsdk:"custom_dkim_selector"`
	Region             types.String `tfsdk:"region"`
	Valid              types.Bool   `tfsdk:"valid"`
	DNSRecords         types.Map    `tfsdk:"dns_records"`
}

func (r *domainAuthenticationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_authentication"
}

func (r *domainAuthenticationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a SendGrid authenticated sending domain (Domain Authentication, formerly " +
			"\"domain whitelabeling\"). After creating this resource, create the DNS records it reports in " +
			"`dns_records` with your DNS provider; SendGrid validates them asynchronously.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The authenticated domain's numeric ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Required:    true,
				Description: "The domain being authenticated.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subdomain": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The subdomain to use for this authenticated domain's DNS records.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Associates this authenticated domain with a subuser's username, allowing that subuser to see and modify it.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ips": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "IP addresses to include in this domain's custom SPF record. Only used when " +
					"`custom_spf` is `true` and `automatic_security` is `false`.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"custom_spf": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				Description: "Whether to use a custom SPF record instead of the one SendGrid manages. Only " +
					"available for domains set up for manual security (`automatic_security = false`).",
			},
			"default": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether to use this as the fallback authenticated domain when no other authenticated domain matches a sender's domain.",
			},
			"automatic_security": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Description: "Whether SendGrid automatically manages this domain's SPF record and DKIM key " +
					"rotation. If omitted, SendGrid's own default is used and read back into state after creation.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"custom_dkim_selector": schema.StringAttribute{
				Optional:    true,
				Description: "A custom three-character DKIM selector.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The region of the domain: `global` or `eu`. Defaults to `global`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"valid": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether SendGrid has validated the required DNS records for this domain.",
			},
			"dns_records": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The DNS records to create with your DNS provider, flattened to " +
					"`\"<record>_type\"`, `\"<record>_host\"`, and `\"<record>_data\"` keys for `mail_cname`, " +
					"`dkim1`, and `dkim2`.",
			},
		},
	}
}

func (r *domainAuthenticationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *domainAuthenticationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan domainAuthenticationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ips, diags := stringListToSlice(ctx, plan.IPs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := sg.CreateDomainAuthenticationParams{
		Domain:             plan.Domain.ValueString(),
		Subdomain:          plan.Subdomain.ValueString(),
		Username:           plan.Username.ValueString(),
		IPs:                ips,
		CustomSPF:          plan.CustomSPF.ValueBool(),
		Default:            plan.Default.ValueBool(),
		CustomDKIMSelector: plan.CustomDKIMSelector.ValueString(),
		Region:             plan.Region.ValueString(),
	}
	// automatic_security has no schema Default because SendGrid's own
	// default isn't documented; only send it when the caller actually set
	// it, so an omitted value doesn't get silently coerced to false.
	if !plan.AutomaticSecurity.IsUnknown() {
		v := plan.AutomaticSecurity.ValueBool()
		params.AutomaticSecurity = &v
	}

	domain, err := r.client.CreateDomainAuthentication(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating domain authentication", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenDomainAuthentication(ctx, domain, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *domainAuthenticationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state domainAuthenticationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.GetDomainAuthentication(ctx, state.ID.ValueString())
	if err != nil {
		if sg.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading domain authentication", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenDomainAuthentication(ctx, domain, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *domainAuthenticationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state domainAuthenticationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.UpdateDomainAuthentication(ctx, state.ID.ValueString(), plan.Default.ValueBool(), plan.CustomSPF.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Error updating domain authentication", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenDomainAuthentication(ctx, domain, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *domainAuthenticationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state domainAuthenticationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteDomainAuthentication(ctx, state.ID.ValueString()); err != nil && !sg.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting domain authentication", err.Error())
	}
}

func (r *domainAuthenticationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func flattenDomainAuthentication(ctx context.Context, d *sg.DomainAuthentication, m *domainAuthenticationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	m.ID = types.StringValue(idFromInt64(d.ID))
	m.Domain = types.StringValue(d.Domain)
	m.Subdomain = types.StringValue(d.Subdomain)
	m.Username = types.StringValue(d.Username)
	m.CustomSPF = types.BoolValue(d.CustomSPF)
	m.Default = types.BoolValue(d.Default)
	m.AutomaticSecurity = types.BoolValue(d.AutomaticSecurity)
	m.Valid = types.BoolValue(d.Valid)
	if d.Region != "" {
		m.Region = types.StringValue(d.Region)
	} else if m.Region.IsUnknown() || m.Region.IsNull() {
		m.Region = types.StringValue("global")
	}

	diags.Append(setStringList(ctx, &m.IPs, d.IPs)...)

	dnsMap, d2 := domainAuthenticationDNSMap(ctx, d.DNS)
	diags.Append(d2...)
	m.DNSRecords = dnsMap

	return diags
}

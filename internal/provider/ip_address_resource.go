package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

var (
	_ resource.Resource                = &ipAddressResource{}
	_ resource.ResourceWithImportState = &ipAddressResource{}
)

func NewIPAddressResource() resource.Resource {
	return &ipAddressResource{}
}

type ipAddressResource struct {
	client *sg.Client
}

type ipAddressResourceModel struct {
	ID               types.String `tfsdk:"id"`
	IP               types.String `tfsdk:"ip"`
	IsAutoWarmup     types.Bool   `tfsdk:"is_auto_warmup"`
	IsParentAssigned types.Bool   `tfsdk:"is_parent_assigned"`
	IsEnabled        types.Bool   `tfsdk:"is_enabled"`
	Subusers         types.Set    `tfsdk:"subusers"`
	Pools            types.List   `tfsdk:"pools"`
	AddedAt          types.Int64  `tfsdk:"added_at"`
	UpdatedAt        types.Int64  `tfsdk:"updated_at"`
	IsLeased         types.Bool   `tfsdk:"is_leased"`
}

func (r *ipAddressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address"
}

func (r *ipAddressResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Adds a dedicated IP address to the account via the IP Address Management API " +
			"(`/v3/send_ips/ips`), distinct from `sendgrid_ip_pool` (which manages pool membership only). " +
			"SendGrid assigns the actual IP address; it cannot be chosen by the caller.\n\n" +
			"~> Dedicated IPs are billed resources. There is no self-service API to release one — destroying " +
			"this resource only removes it from Terraform state. Releasing the IP itself requires contacting " +
			"SendGrid support or adjusting billing.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The IP address (identical to `ip`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ip": schema.StringAttribute{
				Computed:    true,
				Description: "The dedicated IP address assigned by SendGrid.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_auto_warmup": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the IP address is set to automatically warm up.",
			},
			"is_parent_assigned": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether a parent account user is able to send email from this IP address.",
			},
			"is_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Description: "Whether the IP address is billed and able to send email. Only applies to " +
					"non-SendGrid IPs added to the account; always null for SendGrid-leased IPs.",
			},
			"subusers": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Numeric Subuser IDs assigned to send from this IP address (the `id` attribute " +
					"of `sendgrid_subuser`, e.g. `sendgrid_subuser.example.id` — NOT the subuser's username).",
			},
			"pools": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Names of the IP pools this IP address currently belongs to. Read-only; manage " +
					"pool membership via `sendgrid_ip_pool`.",
			},
			"added_at": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp indicating when the IP address was added to the account.",
			},
			"updated_at": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp indicating when the IP address was last updated, if ever.",
			},
			"is_leased": schema.BoolAttribute{
				Computed: true,
				Description: "Whether the IP address is leased from Twilio SendGrid, as opposed to a " +
					"customer-owned IP added to the account.",
			},
		},
	}
}

func (r *ipAddressResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func flattenIPAddressDetail(ctx context.Context, d *sg.IPAddressDetail, m *ipAddressResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	m.ID = types.StringValue(d.IP)
	m.IP = types.StringValue(d.IP)
	m.IsAutoWarmup = types.BoolValue(d.IsAutoWarmup)
	m.IsParentAssigned = types.BoolValue(d.IsParentAssigned)
	if d.IsEnabled != nil {
		m.IsEnabled = types.BoolValue(*d.IsEnabled)
	} else {
		m.IsEnabled = types.BoolNull()
	}
	if d.UpdatedAt != nil {
		m.UpdatedAt = types.Int64Value(*d.UpdatedAt)
	} else {
		m.UpdatedAt = types.Int64Null()
	}
	m.AddedAt = types.Int64Value(d.AddedAt)
	m.IsLeased = types.BoolValue(d.IsLeased)

	poolNames := make([]string, 0, len(d.Pools))
	for _, p := range d.Pools {
		poolNames = append(poolNames, p.Name)
	}
	pools, poolDiags := types.ListValueFrom(ctx, types.StringType, poolNames)
	diags.Append(poolDiags...)
	m.Pools = pools

	diags.Append(setStringSet(ctx, &m.Subusers, d.Subusers)...)
	return diags
}

func (r *ipAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipAddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wantSubusers, diags := stringSetToSlice(ctx, plan.Subusers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	added, err := r.client.AddIPAddress(ctx, plan.IsAutoWarmup.ValueBool(), plan.IsParentAssigned.ValueBool(), wantSubusers)
	if err != nil {
		resp.Diagnostics.AddError("Error adding IP address", err.Error())
		return
	}

	// The IP now exists remotely and is billed. Persist it immediately, so a
	// failure below doesn't leave an untracked, billed IP outside of state.
	plan.ID = types.StringValue(added.IP)
	plan.IP = types.StringValue(added.IP)
	plan.IsAutoWarmup = types.BoolValue(added.IsAutoWarmup)
	plan.IsParentAssigned = types.BoolValue(added.IsParentAssigned)
	plan.IsEnabled = types.BoolNull()
	plan.UpdatedAt = types.Int64Null()
	plan.AddedAt = types.Int64Value(0)
	plan.IsLeased = types.BoolValue(true)
	emptyList, listDiags := types.ListValueFrom(ctx, types.StringType, []string{})
	resp.Diagnostics.Append(listDiags...)
	plan.Pools = emptyList
	resp.Diagnostics.Append(setStringSet(ctx, &plan.Subusers, added.Subusers)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Re-read to pick up server-computed fields (added_at, pools) now that
	// the IP is queryable.
	detail, err := r.client.GetIPAddress(ctx, added.IP)
	if err != nil {
		resp.Diagnostics.AddError("Error reading newly added IP address", err.Error())
		return
	}
	detail.Subusers, err = r.client.ListIPAddressSubusers(ctx, added.IP)
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP address subusers", err.Error())
		return
	}
	resp.Diagnostics.Append(flattenIPAddressDetail(ctx, detail, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipAddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ipAddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	detail, err := r.client.GetIPAddress(ctx, state.ID.ValueString())
	if err != nil {
		if sg.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading IP address", err.Error())
		return
	}
	detail.Subusers, err = r.client.ListIPAddressSubusers(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP address subusers", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenIPAddressDetail(ctx, detail, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ipAddressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ipAddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ip := state.ID.ValueString()

	currentSubusers, diags := stringSetToSlice(ctx, state.Subusers)
	resp.Diagnostics.Append(diags...)
	wantSubusers, diags := stringSetToSlice(ctx, plan.Subusers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = state.ID
	plan.IP = state.IP

	isAutoWarmup := plan.IsAutoWarmup.ValueBool()
	isParentAssigned := plan.IsParentAssigned.ValueBool()
	isEnabled := plan.IsEnabled.ValueBool()
	if _, err := r.client.UpdateIPAddress(ctx, ip, &isAutoWarmup, &isParentAssigned, &isEnabled); err != nil {
		resp.Diagnostics.AddError("Error updating IP address", err.Error())
		return
	}
	resp.Diagnostics.Append(setStringSet(ctx, &plan.Subusers, currentSubusers)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	toAdd, toRemove := diffIPPool(currentSubusers, wantSubusers)

	applied := make(map[string]bool, len(currentSubusers))
	for _, s := range currentSubusers {
		applied[s] = true
	}
	saveProgress := func() {
		appliedSlice := make([]string, 0, len(applied))
		for s := range applied {
			appliedSlice = append(appliedSlice, s)
		}
		resp.Diagnostics.Append(setStringSet(ctx, &plan.Subusers, appliedSlice)...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}

	if len(toRemove) > 0 {
		if err := r.client.RemoveSubusersFromIPAddress(ctx, ip, toRemove); err != nil {
			resp.Diagnostics.AddError("Error removing subusers from IP address", err.Error())
			saveProgress()
			return
		}
		for _, s := range toRemove {
			delete(applied, s)
		}
	}
	if len(toAdd) > 0 {
		if err := r.client.AddSubusersToIPAddress(ctx, ip, toAdd); err != nil {
			resp.Diagnostics.AddError("Error adding subusers to IP address", err.Error())
			saveProgress()
			return
		}
		for _, s := range toAdd {
			applied[s] = true
		}
	}

	saveProgress()
	if resp.Diagnostics.HasError() {
		return
	}

	detail, err := r.client.GetIPAddress(ctx, ip)
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP address", err.Error())
		return
	}
	detail.Subusers, err = r.client.ListIPAddressSubusers(ctx, ip)
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP address subusers", err.Error())
		return
	}
	resp.Diagnostics.Append(flattenIPAddressDetail(ctx, detail, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipAddressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ipAddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.AddWarning(
		"IP address not released",
		"SendGrid has no self-service API to release a dedicated IP address. The IP "+state.IP.ValueString()+
			" remains active (and billed) on the SendGrid account; it has only been removed from Terraform "+
			"state. Contact SendGrid support or adjust billing to actually release it.",
	)
}

func (r *ipAddressResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip"), req.ID)...)
}

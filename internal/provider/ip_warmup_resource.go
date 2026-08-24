package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

var (
	_ resource.Resource                = &ipWarmupResource{}
	_ resource.ResourceWithImportState = &ipWarmupResource{}
)

func NewIPWarmupResource() resource.Resource {
	return &ipWarmupResource{}
}

type ipWarmupResource struct {
	client *sg.Client
}

type ipWarmupResourceModel struct {
	ID        types.String `tfsdk:"id"`
	IP        types.String `tfsdk:"ip"`
	StartDate types.Int64  `tfsdk:"start_date"`
}

func (r *ipWarmupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_warmup"
}

func (r *ipWarmupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the warmup lifecycle of a dedicated IP address already on the account (via " +
			"`/v3/ips/warmup`). Creating this resource starts warmup; destroying it stops warmup. The IP " +
			"address itself must already exist (e.g. via `sendgrid_ip_address`) — this resource only manages " +
			"its warmup state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The IP address (identical to `ip`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ip": schema.StringAttribute{
				Required:    true,
				Description: "The IP address to warm up. Changing this forces replacement.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"start_date": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp indicating when the IP address entered warmup.",
			},
		},
	}
}

func (r *ipWarmupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func flattenIPWarmup(w *sg.IPWarmup, m *ipWarmupResourceModel) {
	m.ID = types.StringValue(w.IP)
	m.IP = types.StringValue(w.IP)
	m.StartDate = types.Int64Value(w.StartDate)
}

func (r *ipWarmupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipWarmupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, err := r.client.StartIPWarmup(ctx, plan.IP.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error starting IP warmup", err.Error())
		return
	}
	if w == nil {
		resp.Diagnostics.AddError("Error starting IP warmup", "SendGrid did not confirm the IP entered warmup.")
		return
	}

	flattenIPWarmup(w, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipWarmupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ipWarmupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, err := r.client.GetIPWarmup(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP warmup status", err.Error())
		return
	}
	if w == nil {
		// No longer in warmup (stopped externally, or warmup completed and
		// SendGrid dropped it from the warmup list).
		resp.State.RemoveResource(ctx)
		return
	}

	flattenIPWarmup(w, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is unreachable in practice: ip is RequiresReplace and start_date is
// fully computed, so no in-place change is ever planned. Implemented as a
// passthrough to satisfy the resource.Resource interface.
func (r *ipWarmupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ipWarmupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipWarmupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ipWarmupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.StopIPWarmup(ctx, state.ID.ValueString()); err != nil && !sg.IsNotFound(err) {
		resp.Diagnostics.AddError("Error stopping IP warmup", err.Error())
	}
}

func (r *ipWarmupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ip"), req.ID)...)
}

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

var (
	_ resource.Resource                = &ipPoolResource{}
	_ resource.ResourceWithImportState = &ipPoolResource{}
)

func NewIPPoolResource() resource.Resource {
	return &ipPoolResource{}
}

type ipPoolResource struct {
	client *sg.Client
}

type ipPoolResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	IPs  types.List   `tfsdk:"ips"`
}

func (r *ipPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_pool"
}

func (r *ipPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a SendGrid IP pool: a named group of dedicated IPs used to segment sending " +
			"reputation. IPs must already be activated for sending on the account (via the SendGrid app) " +
			"before they can be added to a pool.\n\n" +
			"~> The pool's `id` is its name; renaming the pool (changing `name`) updates the ID accordingly " +
			"rather than replacing the resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The pool's name (identical to `name`).",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The pool's name.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 64),
				},
			},
			"ips": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "The IP addresses assigned to this pool.",
			},
		},
	}
}

func (r *ipPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *ipPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wantIPs, diags := stringListToSlice(ctx, plan.IPs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := r.client.CreateIPPool(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating IP pool", err.Error())
		return
	}

	// The pool now exists remotely. Persist it before adding any IPs below,
	// so a failure there doesn't leave Terraform unable to track it.
	plan.ID = types.StringValue(pool.Name)
	resp.Diagnostics.Append(setStringList(ctx, &plan.IPs, nil)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var added []string
	for _, ip := range wantIPs {
		if err := r.client.AddIPToPool(ctx, pool.Name, ip); err != nil {
			resp.Diagnostics.AddError("Error adding IP to pool", err.Error())
			resp.Diagnostics.Append(setStringList(ctx, &plan.IPs, added)...)
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}
		added = append(added, ip)
	}

	resp.Diagnostics.Append(setStringList(ctx, &plan.IPs, added)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ipPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := r.client.GetIPPool(ctx, state.ID.ValueString())
	if err != nil {
		if sg.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading IP pool", err.Error())
		return
	}

	state.ID = types.StringValue(pool.Name)
	state.Name = types.StringValue(pool.Name)
	resp.Diagnostics.Append(setStringList(ctx, &state.IPs, pool.IPs)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ipPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ipPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	poolName := state.Name.ValueString()
	if plan.Name.ValueString() != state.Name.ValueString() {
		pool, err := r.client.UpdateIPPool(ctx, state.Name.ValueString(), plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error renaming IP pool", err.Error())
			return
		}
		poolName = pool.Name
		plan.ID = types.StringValue(pool.Name)
		plan.IPs = state.IPs
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		plan.ID = state.ID
	}

	currentIPs, diags := stringListToSlice(ctx, state.IPs)
	resp.Diagnostics.Append(diags...)
	wantIPs, diags := stringListToSlice(ctx, plan.IPs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	current := make(map[string]bool, len(currentIPs))
	for _, ip := range currentIPs {
		current[ip] = true
	}
	want := make(map[string]bool, len(wantIPs))
	for _, ip := range wantIPs {
		want[ip] = true
	}

	applied := append([]string{}, currentIPs...)
	saveProgress := func() {
		resp.Diagnostics.Append(setStringList(ctx, &plan.IPs, applied)...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}

	for _, ip := range currentIPs {
		if want[ip] {
			continue
		}
		if err := r.client.RemoveIPFromPool(ctx, poolName, ip); err != nil {
			resp.Diagnostics.AddError("Error removing IP from pool", err.Error())
			saveProgress()
			return
		}
		for i, a := range applied {
			if a == ip {
				applied = append(applied[:i], applied[i+1:]...)
				break
			}
		}
	}
	for _, ip := range wantIPs {
		if current[ip] {
			continue
		}
		if err := r.client.AddIPToPool(ctx, poolName, ip); err != nil {
			resp.Diagnostics.AddError("Error adding IP to pool", err.Error())
			saveProgress()
			return
		}
		applied = append(applied, ip)
	}

	saveProgress()
}

func (r *ipPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ipPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteIPPool(ctx, state.ID.ValueString()); err != nil && !sg.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting IP pool", err.Error())
	}
}

func (r *ipPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

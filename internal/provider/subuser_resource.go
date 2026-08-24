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
	_ resource.Resource                = &subuserResource{}
	_ resource.ResourceWithImportState = &subuserResource{}
)

func NewSubuserResource() resource.Resource {
	return &subuserResource{}
}

type subuserResource struct {
	client *sg.Client
}

type subuserResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Username types.String `tfsdk:"username"`
	Email    types.String `tfsdk:"email"`
	Password types.String `tfsdk:"password"`
	IPs      types.List   `tfsdk:"ips"`
	Region   types.String `tfsdk:"region"`
	Disabled types.Bool   `tfsdk:"disabled"`
}

func (r *subuserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subuser"
}

func (r *subuserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a SendGrid subuser: an isolated sub-account with its own sending reputation, " +
			"templates, and API keys, typically used to separate sending on behalf of different products or " +
			"customers under one parent account.\n\n" +
			"~> SendGrid's API does not support changing a subuser's `username`, `email`, `password`, or " +
			"`region` after creation; changing any of these forces recreation. `ips` and `disabled` can be " +
			"updated in place.\n\n" +
			"~> SendGrid has no endpoint that returns a subuser's currently assigned IPs, so `ips` cannot be " +
			"refreshed by `terraform plan` and is not populated by `terraform import`; on import, the first " +
			"plan will show a change to `ips` and apply it via the update-IPs endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The subuser's numeric user ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "The subuser's username. Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "The subuser's email address. Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The password the subuser will use to log into SendGrid. Cannot be changed via this resource after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ips": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "The IP addresses assigned to this subuser. At least one is required at creation " +
					"time; can be updated in place afterward.",
			},
			"region": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The region this subuser is assigned to: `global` or `eu`. Defaults to `global`. Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"disabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether this subuser is disabled. Defaults to `false`.",
			},
		},
	}
}

func (r *subuserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *subuserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan subuserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ips, diags := stringListToSlice(ctx, plan.IPs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub, err := r.client.CreateSubuser(ctx, plan.Username.ValueString(), plan.Email.ValueString(),
		plan.Password.ValueString(), ips, plan.Region.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating subuser", err.Error())
		return
	}

	if plan.Disabled.ValueBool() {
		if err := r.client.UpdateSubuserDisabled(ctx, sub.Username, true); err != nil {
			resp.Diagnostics.AddError("Error disabling subuser after creation", err.Error())
			return
		}
	} else {
		plan.Disabled = types.BoolValue(false)
	}

	plan.ID = types.StringValue(idFromInt64(sub.UserID))
	if sub.Region != "" {
		plan.Region = types.StringValue(sub.Region)
	} else {
		plan.Region = types.StringValue("global")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subuserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state subuserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub, err := r.client.GetSubuserByUsername(ctx, state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading subuser", err.Error())
		return
	}
	if sub == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = types.StringValue(idFromInt64(sub.ID))
	state.Email = types.StringValue(sub.Email)
	state.Disabled = types.BoolValue(sub.Disabled)
	if sub.Region != "" {
		state.Region = types.StringValue(sub.Region)
	}
	// state.IPs is intentionally left untouched: SendGrid has no endpoint
	// that returns a subuser's currently assigned IPs, so the value already
	// in state is authoritative.

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *subuserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state subuserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Disabled.ValueBool() != state.Disabled.ValueBool() {
		if err := r.client.UpdateSubuserDisabled(ctx, state.Username.ValueString(), plan.Disabled.ValueBool()); err != nil {
			resp.Diagnostics.AddError("Error updating subuser", err.Error())
			return
		}
	}

	if !plan.IPs.Equal(state.IPs) {
		ips, diags := stringListToSlice(ctx, plan.IPs)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		updated, err := r.client.UpdateSubuserIPs(ctx, state.Username.ValueString(), ips)
		if err != nil {
			resp.Diagnostics.AddError("Error updating subuser IPs", err.Error())
			return
		}
		resp.Diagnostics.Append(setStringList(ctx, &plan.IPs, updated)...)
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *subuserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state subuserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSubuser(ctx, state.Username.ValueString()); err != nil && !sg.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting subuser", err.Error())
	}
}

func (r *subuserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("username"), req, resp)
}

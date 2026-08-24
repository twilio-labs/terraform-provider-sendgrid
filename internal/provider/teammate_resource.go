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
	_ resource.Resource                = &teammateResource{}
	_ resource.ResourceWithImportState = &teammateResource{}
)

func NewTeammateResource() resource.Resource {
	return &teammateResource{}
}

type teammateResource struct {
	client *sg.Client
}

type teammateResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Email    types.String `tfsdk:"email"`
	Scopes   types.Set    `tfsdk:"scopes"`
	IsAdmin  types.Bool   `tfsdk:"is_admin"`
	Token    types.String `tfsdk:"token"`
	Username types.String `tfsdk:"username"`
	Status   types.String `tfsdk:"status"`
}

func (r *teammateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teammate"
}

func (r *teammateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Invites a teammate to the account via SendGrid's REST Teammates API.\n\n" +
			"~> SendGrid has no SCIM API; this resource uses the standard Teammates API " +
			"(`/v3/teammates`), which is invite-based. Creating this resource sends an email " +
			"invitation — it does not immediately create a queryable teammate. Until the invite is " +
			"accepted, the teammate exists only as a pending invite (tracked here via `token`); once " +
			"accepted, it becomes an account teammate (tracked via `username`). `status` reflects which " +
			"phase the invite is currently in. Changing `scopes` or `is_admin` while still pending " +
			"cancels the old invite and sends a new one (the invite `token` changes as a result).\n\n" +
			"~> Import is by `email` (there is no endpoint to look up an invite or teammate by email " +
			"directly, so the first refresh after import searches both the pending-invite and " +
			"accepted-teammate lists for a match).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Same as `email`, which uniquely and stably identifies this teammate across the pending/accepted lifecycle.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "The teammate's email address. Cannot be changed after creation; a new invite must be sent instead.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scopes": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "The permission scopes granted to the teammate. Must be empty if `is_admin` is `true`.",
			},
			"is_admin": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the teammate has admin privileges. Defaults to `false`.",
			},
			"token": schema.StringAttribute{
				Computed:    true,
				Description: "The invite token, while the invite is pending. Empty once accepted.",
			},
			"username": schema.StringAttribute{
				Computed:    true,
				Description: "The teammate's username, once they have accepted the invite. Empty while pending.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Either `pending` (invite sent, not yet accepted) or `accepted`.",
			},
		},
	}
}

func (r *teammateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

// findTeammateByEmail searches the pending-invite list, then the accepted-
// teammate list, for an entry matching email. Exactly one of the two return
// values is non-nil if found; both are nil if email matches neither list.
// Shared by the resource and data source so the two-phase lookup logic
// lives in exactly one place.
func findTeammateByEmail(ctx context.Context, client *sg.Client, email string) (*sg.TeammateInvite, *sg.Teammate, error) {
	invites, err := client.ListPendingTeammateInvites(ctx)
	if err != nil {
		return nil, nil, err
	}
	for _, inv := range invites {
		if inv.Email == email {
			return &inv, nil, nil
		}
	}

	teammates, err := client.ListTeammates(ctx)
	if err != nil {
		return nil, nil, err
	}
	for _, tm := range teammates {
		if tm.Email == email {
			return nil, &tm, nil
		}
	}

	return nil, nil, nil
}

func (r *teammateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan teammateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopes, diags := stringSetToSlice(ctx, plan.Scopes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	invite, err := r.client.InviteTeammate(ctx, plan.Email.ValueString(), scopes, plan.IsAdmin.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError("Error inviting teammate", err.Error())
		return
	}

	plan.ID = types.StringValue(invite.Email)
	plan.Token = types.StringValue(invite.Token)
	plan.Username = types.StringValue("")
	plan.Status = types.StringValue("pending")
	plan.IsAdmin = types.BoolValue(invite.IsAdmin)
	resp.Diagnostics.Append(setStringSet(ctx, &plan.Scopes, invite.Scopes)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *teammateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state teammateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	invite, teammate, err := findTeammateByEmail(ctx, r.client, state.Email.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading teammate", err.Error())
		return
	}

	if invite == nil && teammate == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if invite != nil {
		state.Token = types.StringValue(invite.Token)
		state.Username = types.StringValue("")
		state.Status = types.StringValue("pending")
		state.IsAdmin = types.BoolValue(invite.IsAdmin)
		resp.Diagnostics.Append(setStringSet(ctx, &state.Scopes, invite.Scopes)...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	// Accepted: refresh scopes/is_admin from the single-teammate endpoint,
	// since the list endpoint that found it doesn't return scopes.
	full, err := r.client.GetTeammate(ctx, teammate.Username)
	if err != nil {
		resp.Diagnostics.AddError("Error reading teammate", err.Error())
		return
	}

	state.Token = types.StringValue("")
	state.Username = types.StringValue(full.Username)
	state.Status = types.StringValue("accepted")
	state.IsAdmin = types.BoolValue(full.IsAdmin)
	resp.Diagnostics.Append(setStringSet(ctx, &state.Scopes, full.Scopes)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *teammateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state teammateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopes, diags := stringSetToSlice(ctx, plan.Scopes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	isAdmin := plan.IsAdmin.ValueBool()

	// result tracks the last-known-good state, checkpointed after each
	// side-effecting call so a later failure doesn't orphan an already-
	// applied earlier one.
	result := state

	if state.Status.ValueString() == "accepted" {
		tm, err := r.client.UpdateTeammate(ctx, state.Username.ValueString(), scopes, isAdmin)
		if err != nil {
			resp.Diagnostics.AddError("Error updating teammate", err.Error())
			resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
			return
		}
		result.IsAdmin = types.BoolValue(tm.IsAdmin)
		resp.Diagnostics.Append(setStringSet(ctx, &result.Scopes, tm.Scopes)...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		return
	}

	// Still pending: there is no PATCH for a pending invite. Cancel the old
	// one and send a fresh one with the new scopes/is_admin.
	if err := r.client.CancelTeammateInvite(ctx, state.Token.ValueString()); err != nil && !sg.IsNotFound(err) {
		resp.Diagnostics.AddError("Error updating teammate", err.Error())
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		return
	}

	invite, err := r.client.InviteTeammate(ctx, state.Email.ValueString(), scopes, isAdmin)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error re-inviting teammate after updating scopes",
			"The previous invite was cancelled, but sending the replacement invite failed: "+err.Error()+
				". The teammate is no longer invited; the next apply will need to send a fresh invite.",
		)
		result.Token = types.StringValue("")
		result.Status = types.StringValue("")
		resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
		return
	}

	result.Token = types.StringValue(invite.Token)
	result.Status = types.StringValue("pending")
	result.IsAdmin = types.BoolValue(invite.IsAdmin)
	resp.Diagnostics.Append(setStringSet(ctx, &result.Scopes, invite.Scopes)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &result)...)
}

func (r *teammateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state teammateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.Status.ValueString() == "accepted" {
		if err := r.client.DeleteTeammate(ctx, state.Username.ValueString()); err != nil && !sg.IsNotFound(err) {
			resp.Diagnostics.AddError("Error deleting teammate", err.Error())
		}
		return
	}

	if err := r.client.CancelTeammateInvite(ctx, state.Token.ValueString()); err != nil && !sg.IsNotFound(err) {
		resp.Diagnostics.AddError("Error cancelling teammate invite", err.Error())
	}
}

func (r *teammateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("email"), req.ID)...)
}

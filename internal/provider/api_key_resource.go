package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

// autoInjectedScopes are scopes SendGrid silently adds to an API key
// regardless of what was requested (e.g. two-factor auth requirements).
// Terraform must not treat their presence as drift unless the user
// explicitly asked for them.
var autoInjectedScopes = map[string]bool{
	"2fa_required":                 true,
	"sender_verification_eligible": true,
	"sender_verification_legacy":   true,
}

var (
	_ resource.Resource                = &apiKeyResource{}
	_ resource.ResourceWithImportState = &apiKeyResource{}
)

func NewAPIKeyResource() resource.Resource {
	return &apiKeyResource{}
}

type apiKeyResource struct {
	client *sg.Client
}

type apiKeyResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Scopes types.Set    `tfsdk:"scopes"`
	APIKey types.String `tfsdk:"api_key"`
}

func (r *apiKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (r *apiKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a SendGrid API key. See the [API Key Permissions List]" +
			"(https://docs.sendgrid.com/api-reference/how-to-use-the-sendgrid-v3-api/authorization) for available scopes.\n\n" +
			"~> SendGrid never returns an API key's secret value after creation. Destroying and recreating this " +
			"resource is the only way to obtain a new secret; there is no way to recover a lost one.\n\n" +
			"~> SendGrid automatically adds `2fa_required`, `sender_verification_eligible`, and/or " +
			"`sender_verification_legacy` to some keys' scopes regardless of what is requested. This provider " +
			"ignores those scopes when they are not explicitly present in `scopes` to avoid permanent diffs.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The API key ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "A descriptive name for this API key. Does not need to be unique.",
			},
			"scopes": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "The permissions granted to this API key. Omitting this grants Full Access.",
			},
			"api_key": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The API key secret. Only ever populated at creation time.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *apiKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *apiKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan apiKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	scopes, diags := stringSetToSlice(ctx, plan.Scopes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, err := r.client.CreateAPIKey(ctx, plan.Name.ValueString(), scopes)
	if err != nil {
		resp.Diagnostics.AddError("Error creating API key", err.Error())
		return
	}

	plan.ID = types.StringValue(key.APIKeyID)
	plan.APIKey = types.StringValue(key.APIKey)
	resp.Diagnostics.Append(setStringSet(ctx, &plan.Scopes, filterAutoInjectedScopes(key.Scopes, scopes))...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state apiKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	priorScopes, diags := stringSetToSlice(ctx, state.Scopes)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, err := r.client.GetAPIKey(ctx, state.ID.ValueString())
	if err != nil {
		if sg.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading API key", err.Error())
		return
	}

	state.Name = types.StringValue(key.Name)
	resp.Diagnostics.Append(setStringSet(ctx, &state.Scopes, filterAutoInjectedScopes(key.Scopes, priorScopes))...)
	// api_key is intentionally left untouched: SendGrid never returns it
	// after creation, so the value already in state is authoritative.

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *apiKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state apiKeyResourceModel
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

	key, err := r.client.UpdateAPIKey(ctx, state.ID.ValueString(), plan.Name.ValueString(), scopes)
	if err != nil {
		resp.Diagnostics.AddError("Error updating API key", err.Error())
		return
	}

	plan.ID = state.ID
	plan.APIKey = state.APIKey
	resp.Diagnostics.Append(setStringSet(ctx, &plan.Scopes, filterAutoInjectedScopes(key.Scopes, scopes))...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *apiKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state apiKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAPIKey(ctx, state.ID.ValueString()); err != nil && !sg.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting API key", err.Error())
	}
}

func (r *apiKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// filterAutoInjectedScopes drops any scope in observed that SendGrid may have
// silently added, unless it was already part of the caller's desired scopes.
func filterAutoInjectedScopes(observed, desired []string) []string {
	wanted := make(map[string]bool, len(desired))
	for _, s := range desired {
		wanted[s] = true
	}
	filtered := make([]string, 0, len(observed))
	for _, s := range observed {
		if autoInjectedScopes[s] && !wanted[s] {
			continue
		}
		filtered = append(filtered, s)
	}
	return filtered
}

func clientFromProviderData(providerData any) (*sg.Client, diag.Diagnostics) {
	var diags diag.Diagnostics
	if providerData == nil {
		return nil, diags
	}
	client, ok := providerData.(*sg.Client)
	if !ok {
		diags.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sendgrid.Client, got: %T. This is a bug in the provider.", providerData),
		)
		return nil, diags
	}
	return client, diags
}

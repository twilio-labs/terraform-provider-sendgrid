package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-labs/terraform-provider-sendgrid/internal/sendgrid"
)

var (
	_ resource.Resource                = &templateResource{}
	_ resource.ResourceWithImportState = &templateResource{}
)

func NewTemplateResource() resource.Resource {
	return &templateResource{}
}

type templateResource struct {
	client *sg.Client
}

type templateResourceModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Generation types.String `tfsdk:"generation"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

func (r *templateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template"
}

func (r *templateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a SendGrid transactional template. A template's content is managed separately " +
			"via one or more `sendgrid_template_version` resources.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The template's UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name for this transactional template. Does not need to be unique.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(100),
				},
			},
			"generation": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("dynamic"),
				Description: "Whether this template supports dynamic replacements (Handlebars): `legacy` or `dynamic`.",
				Validators: []validator.String{
					stringvalidator.OneOf("legacy", "dynamic"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The time this template was last updated.",
			},
		},
	}
}

func (r *templateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func (r *templateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan templateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := r.client.CreateTemplate(ctx, plan.Name.ValueString(), plan.Generation.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating template", err.Error())
		return
	}

	plan.ID = types.StringValue(tmpl.ID)
	plan.Generation = types.StringValue(tmpl.Generation)
	plan.UpdatedAt = types.StringValue(tmpl.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state templateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := r.client.GetTemplate(ctx, state.ID.ValueString())
	if err != nil {
		if sg.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading template", err.Error())
		return
	}

	state.Name = types.StringValue(tmpl.Name)
	state.Generation = types.StringValue(tmpl.Generation)
	state.UpdatedAt = types.StringValue(tmpl.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *templateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state templateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := r.client.UpdateTemplate(ctx, state.ID.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating template", err.Error())
		return
	}

	plan.ID = state.ID
	plan.Generation = state.Generation
	plan.UpdatedAt = types.StringValue(tmpl.UpdatedAt)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state templateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTemplate(ctx, state.ID.ValueString()); err != nil && !sg.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting template", err.Error())
	}
}

func (r *templateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

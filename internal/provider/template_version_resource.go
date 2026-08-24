package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

var (
	_ resource.Resource                = &templateVersionResource{}
	_ resource.ResourceWithImportState = &templateVersionResource{}
)

func NewTemplateVersionResource() resource.Resource {
	return &templateVersionResource{}
}

type templateVersionResource struct {
	client *sg.Client
}

type templateVersionResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	TemplateID           types.String `tfsdk:"template_id"`
	Active               types.Bool   `tfsdk:"active"`
	Name                 types.String `tfsdk:"name"`
	HTMLContent          types.String `tfsdk:"html_content"`
	PlainContent         types.String `tfsdk:"plain_content"`
	GeneratePlainContent types.Bool   `tfsdk:"generate_plain_content"`
	Subject              types.String `tfsdk:"subject"`
	Editor               types.String `tfsdk:"editor"`
	TestData             types.String `tfsdk:"test_data"`
	UpdatedAt            types.String `tfsdk:"updated_at"`
}

func (r *templateVersionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template_version"
}

func (r *templateVersionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a single version of a SendGrid transactional template. Only one version of a " +
			"template can be `active` (served when sending mail with that template) at a time; setting " +
			"`active = true` on one version does not automatically deactivate others in Terraform state, " +
			"matching the API's own behavior.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The template version's UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"template_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the template this version belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether this is the active version served for the template.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of this template version.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(100),
				},
			},
			"html_content": schema.StringAttribute{
				Optional:    true,
				Description: "The HTML content of this version. Maximum of 1,048,576 bytes.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(1048576),
				},
			},
			"plain_content": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Description: "The plain-text content of this version. Generated from `html_content` unless " +
					"`generate_plain_content` is `false`.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(1048576),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"generate_plain_content": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "If true, `plain_content` is always regenerated from `html_content`.",
			},
			"subject": schema.StringAttribute{
				Required:    true,
				Description: "The subject of this template version.",
				Validators: []validator.String{
					stringvalidator.LengthAtMost(255),
				},
			},
			"editor": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("code"),
				Description: "The editor used in the SendGrid UI for this version: `code` or `design`.",
				Validators: []validator.String{
					stringvalidator.OneOf("code", "design"),
				},
			},
			"test_data": schema.StringAttribute{
				Optional:    true,
				Description: "For dynamic templates only: mock JSON data used for template preview and test sends.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The time this template version was last updated.",
			},
		},
	}
}

func (r *templateVersionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func templateVersionFromModel(m templateVersionResourceModel) sg.TemplateVersion {
	active := 0
	if m.Active.ValueBool() {
		active = 1
	}
	generatePlain := m.GeneratePlainContent.ValueBool()
	return sg.TemplateVersion{
		Active:               &active,
		Name:                 m.Name.ValueString(),
		HTMLContent:          m.HTMLContent.ValueString(),
		PlainContent:         m.PlainContent.ValueString(),
		GeneratePlainContent: &generatePlain,
		Subject:              m.Subject.ValueString(),
		Editor:               m.Editor.ValueString(),
		TestData:             m.TestData.ValueString(),
	}
}

func (r *templateVersionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan templateVersionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	v, err := r.client.CreateTemplateVersion(ctx, plan.TemplateID.ValueString(), templateVersionFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating template version", err.Error())
		return
	}

	flattenTemplateVersion(v, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateVersionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state templateVersionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	v, err := r.client.GetTemplateVersion(ctx, state.TemplateID.ValueString(), state.ID.ValueString())
	if err != nil {
		if sg.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading template version", err.Error())
		return
	}

	flattenTemplateVersion(v, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *templateVersionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state templateVersionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	v, err := r.client.UpdateTemplateVersion(ctx, state.TemplateID.ValueString(), state.ID.ValueString(), templateVersionFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating template version", err.Error())
		return
	}

	plan.ID = state.ID
	plan.TemplateID = state.TemplateID
	flattenTemplateVersion(v, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *templateVersionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state templateVersionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteTemplateVersion(ctx, state.TemplateID.ValueString(), state.ID.ValueString()); err != nil && !sg.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting template version", err.Error())
	}
}

func (r *templateVersionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	templateID, versionID, err := parseTemplateVersionImportID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Import Identifier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("template_id"), templateID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), versionID)...)
}

// parseTemplateVersionImportID splits a "template_id/version_id" import
// identifier. version_id is a UUID and so cannot itself contain "/", so a
// simple two-part split (rather than requiring exactly one separator) is
// sufficient and matches how SplitN(id, "/", 2) behaves on well-formed input.
func parseTemplateVersionImportID(id string) (templateID, versionID string, err error) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("expected import identifier in the form template_id/version_id, got: %s", id)
	}
	return parts[0], parts[1], nil
}

func flattenTemplateVersion(v *sg.TemplateVersion, m *templateVersionResourceModel) {
	m.ID = types.StringValue(v.ID)
	if v.TemplateID != "" {
		m.TemplateID = types.StringValue(v.TemplateID)
	}
	m.Active = types.BoolValue(v.IsActive())
	m.Name = types.StringValue(v.Name)
	m.HTMLContent = types.StringValue(v.HTMLContent)
	m.PlainContent = types.StringValue(v.PlainContent)
	if v.GeneratePlainContent != nil {
		m.GeneratePlainContent = types.BoolValue(*v.GeneratePlainContent)
	}
	m.Subject = types.StringValue(v.Subject)
	if v.Editor != "" {
		m.Editor = types.StringValue(v.Editor)
	}
	m.UpdatedAt = types.StringValue(v.UpdatedAt)
}

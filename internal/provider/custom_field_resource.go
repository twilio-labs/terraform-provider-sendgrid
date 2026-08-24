package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

var (
	_ resource.Resource                = &customFieldResource{}
	_ resource.ResourceWithImportState = &customFieldResource{}
)

func NewCustomFieldResource() resource.Resource {
	return &customFieldResource{}
}

type customFieldResource struct {
	client *sg.Client
}

type customFieldResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	FieldType types.String `tfsdk:"field_type"`
}

func (r *customFieldResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_field"
}

func (r *customFieldResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a SendGrid Marketing Campaigns custom contact field definition. Accounts are " +
			"limited to 500 custom fields. Field names must be case-insensitively unique and cannot collide " +
			"with SendGrid's reserved field names.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The custom field's ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The field name. Must be alphanumeric/underscore and start with a letter or underscore.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
			},
			"field_type": schema.StringAttribute{
				Required:    true,
				Description: "The field's data type: `Text`, `Number`, or `Date`. Cannot be changed after creation.",
				Validators: []validator.String{
					stringvalidator.OneOf("Text", "Number", "Date"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *customFieldResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func flattenCustomField(f *sg.CustomField, m *customFieldResourceModel) {
	m.ID = types.StringValue(f.ID)
	m.Name = types.StringValue(f.Name)
	m.FieldType = types.StringValue(f.FieldType)
}

func (r *customFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customFieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	f, err := r.client.CreateCustomField(ctx, plan.Name.ValueString(), plan.FieldType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating custom field", err.Error())
		return
	}

	flattenCustomField(f, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customFieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	f, err := r.client.GetCustomField(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading custom field", err.Error())
		return
	}
	if f == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	flattenCustomField(f, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *customFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state customFieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	f, err := r.client.UpdateCustomField(ctx, state.ID.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating custom field", err.Error())
		return
	}

	flattenCustomField(f, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customFieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteCustomField(ctx, state.ID.ValueString()); err != nil && !sg.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting custom field", err.Error())
	}
}

func (r *customFieldResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

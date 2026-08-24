package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

var (
	_ resource.Resource                = &segmentResource{}
	_ resource.ResourceWithImportState = &segmentResource{}
)

func NewSegmentResource() resource.Resource {
	return &segmentResource{}
}

type segmentResource struct {
	client *sg.Client
}

type segmentResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	QueryDSL      types.String `tfsdk:"query_dsl"`
	ParentListIDs types.List   `tfsdk:"parent_list_ids"`
	ContactsCount types.Int64  `tfsdk:"contacts_count"`
}

func (r *segmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_segment"
}

func (r *segmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a SendGrid Marketing Campaigns contact segment: a dynamic group of contacts " +
			"matching a SQL-like query.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The segment's UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The segment's name. Must be unique across all segments in the account.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
			},
			"query_dsl": schema.StringAttribute{
				Required:    true,
				Description: "The SQL-like query used to filter contacts into this segment.",
			},
			"parent_list_ids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Restricts this segment to contacts on these list IDs. Cannot be changed after creation.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"contacts_count": schema.Int64Attribute{
				Computed:    true,
				Description: "The number of contacts currently matching this segment.",
			},
		},
	}
}

func (r *segmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func flattenSegment(ctx context.Context, s *sg.Segment, m *segmentResourceModel) diag.Diagnostics {
	m.ID = types.StringValue(s.ID)
	m.Name = types.StringValue(s.Name)
	m.QueryDSL = types.StringValue(s.QueryDSL)
	m.ContactsCount = types.Int64Value(s.ContactsCount)
	return setStringList(ctx, &m.ParentListIDs, s.ParentListIDs)
}

func (r *segmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan segmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	parentListIDs, diags := stringListToSlice(ctx, plan.ParentListIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.CreateSegment(ctx, plan.Name.ValueString(), plan.QueryDSL.ValueString(), parentListIDs)
	if err != nil {
		resp.Diagnostics.AddError("Error creating segment", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenSegment(ctx, s, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *segmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state segmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetSegment(ctx, state.ID.ValueString())
	if err != nil {
		if sg.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading segment", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenSegment(ctx, s, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *segmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state segmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateSegment(ctx, state.ID.ValueString(), plan.Name.ValueString(), plan.QueryDSL.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating segment", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenSegment(ctx, s, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *segmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state segmentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSegment(ctx, state.ID.ValueString()); err != nil && !sg.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting segment", err.Error())
	}
}

func (r *segmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

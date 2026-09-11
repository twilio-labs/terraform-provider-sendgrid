package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-labs/terraform-provider-sendgrid/internal/sendgrid"
)

var _ datasource.DataSource = &segmentDataSource{}

func NewSegmentDataSource() datasource.DataSource {
	return &segmentDataSource{}
}

type segmentDataSource struct {
	client *sg.Client
}

type segmentDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	QueryDSL      types.String `tfsdk:"query_dsl"`
	ParentListIDs types.List   `tfsdk:"parent_list_ids"`
	ContactsCount types.Int64  `tfsdk:"contacts_count"`
}

func (d *segmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_segment"
}

func (d *segmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves an existing SendGrid Marketing Campaigns contact segment by ID.",
		Attributes: map[string]schema.Attribute{
			"id":              schema.StringAttribute{Required: true, Description: "The segment's UUID."},
			"name":            schema.StringAttribute{Computed: true, Description: "The segment's name."},
			"query_dsl":       schema.StringAttribute{Computed: true, Description: "The SQL-like query used to filter contacts into this segment."},
			"parent_list_ids": schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Restricts this segment to contacts on these list IDs."},
			"contacts_count":  schema.Int64Attribute{Computed: true, Description: "The number of contacts currently matching this segment."},
		},
	}
}

func (d *segmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *segmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config segmentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := d.client.GetSegment(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading segment", err.Error())
		return
	}

	config.ID = types.StringValue(s.ID)
	config.Name = types.StringValue(s.Name)
	config.QueryDSL = types.StringValue(s.QueryDSL)
	config.ContactsCount = types.Int64Value(s.ContactsCount)
	resp.Diagnostics.Append(setStringList(ctx, &config.ParentListIDs, s.ParentListIDs)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

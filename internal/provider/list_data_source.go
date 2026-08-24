package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

var _ datasource.DataSource = &listDataSource{}

func NewListDataSource() datasource.DataSource {
	return &listDataSource{}
}

type listDataSource struct {
	client *sg.Client
}

func (d *listDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_list"
}

func (d *listDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves an existing SendGrid Marketing Campaigns contact list by ID.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Required: true, Description: "The list's UUID."},
			"name":          schema.StringAttribute{Computed: true, Description: "The list's name."},
			"contact_count": schema.Int64Attribute{Computed: true, Description: "The number of contacts currently on this list."},
		},
	}
}

func (d *listDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *listDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config listResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	l, err := d.client.GetList(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading list", err.Error())
		return
	}

	flattenList(l, &config)
	config.ID = types.StringValue(l.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

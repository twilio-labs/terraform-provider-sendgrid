package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-labs/terraform-provider-sendgrid/internal/sendgrid"
)

var _ datasource.DataSource = &customFieldDataSource{}

func NewCustomFieldDataSource() datasource.DataSource {
	return &customFieldDataSource{}
}

type customFieldDataSource struct {
	client *sg.Client
}

type customFieldDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	FieldType types.String `tfsdk:"field_type"`
}

func (d *customFieldDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_field"
}

func (d *customFieldDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves an existing SendGrid Marketing Campaigns custom field definition by ID.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Required: true, Description: "The custom field's ID."},
			"name":       schema.StringAttribute{Computed: true, Description: "The field name."},
			"field_type": schema.StringAttribute{Computed: true, Description: "The field's data type: `Text`, `Number`, or `Date`."},
		},
	}
}

func (d *customFieldDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *customFieldDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config customFieldDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	f, err := d.client.GetCustomField(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading custom field", err.Error())
		return
	}
	if f == nil {
		resp.Diagnostics.AddError("Custom field not found", "No custom field with ID \""+config.ID.ValueString()+"\" was found.")
		return
	}

	config.ID = types.StringValue(f.ID)
	config.Name = types.StringValue(f.Name)
	config.FieldType = types.StringValue(f.FieldType)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

var _ datasource.DataSource = &apiKeyDataSource{}

func NewAPIKeyDataSource() datasource.DataSource {
	return &apiKeyDataSource{}
}

type apiKeyDataSource struct {
	client *sg.Client
}

type apiKeyDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Scopes types.Set    `tfsdk:"scopes"`
}

func (d *apiKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (d *apiKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves an existing SendGrid API key's name and scopes by ID. The secret value is " +
			"never returned by the API and is not available here.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The API key ID.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "A descriptive name for this API key.",
			},
			"scopes": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The permissions granted to this API key.",
			},
		},
	}
}

func (d *apiKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *apiKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config apiKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key, err := d.client.GetAPIKey(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading API key", err.Error())
		return
	}

	config.Name = types.StringValue(key.Name)
	resp.Diagnostics.Append(setStringSet(ctx, &config.Scopes, key.Scopes)...)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

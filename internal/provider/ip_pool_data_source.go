package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

var _ datasource.DataSource = &ipPoolDataSource{}

func NewIPPoolDataSource() datasource.DataSource {
	return &ipPoolDataSource{}
}

type ipPoolDataSource struct {
	client *sg.Client
}

type ipPoolDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	IPs  types.Set    `tfsdk:"ips"`
}

func (d *ipPoolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_pool"
}

func (d *ipPoolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves an existing SendGrid IP pool by name, including its assigned IPs.",
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{Computed: true, Description: "The pool's name (identical to `name`)."},
			"name": schema.StringAttribute{Required: true, Description: "The pool's name."},
			"ips":  schema.SetAttribute{Computed: true, ElementType: types.StringType, Description: "The IP addresses assigned to this pool."},
		},
	}
}

func (d *ipPoolDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *ipPoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ipPoolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := d.client.GetIPPool(ctx, config.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP pool", err.Error())
		return
	}

	config.ID = types.StringValue(pool.Name)
	config.Name = types.StringValue(pool.Name)
	resp.Diagnostics.Append(setStringSet(ctx, &config.IPs, pool.IPs)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

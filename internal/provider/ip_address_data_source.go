package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

var _ datasource.DataSource = &ipAddressDataSource{}

func NewIPAddressDataSource() datasource.DataSource {
	return &ipAddressDataSource{}
}

type ipAddressDataSource struct {
	client *sg.Client
}

func (d *ipAddressDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address"
}

func (d *ipAddressDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the current details of a dedicated IP address already on the account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The IP address (identical to `ip`).",
			},
			"ip": schema.StringAttribute{
				Required:    true,
				Description: "The dedicated IP address to look up.",
			},
			"is_auto_warmup": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the IP address is set to automatically warm up.",
			},
			"is_parent_assigned": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether a parent account user is able to send email from this IP address.",
			},
			"is_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the IP address is billed and able to send email.",
			},
			"subusers": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Numeric Subuser IDs assigned to send from this IP address.",
			},
			"pools": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Names of the IP pools this IP address currently belongs to.",
			},
			"added_at": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp indicating when the IP address was added to the account.",
			},
			"updated_at": schema.Int64Attribute{
				Computed:    true,
				Description: "Unix timestamp indicating when the IP address was last updated, if ever.",
			},
			"is_leased": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the IP address is leased from Twilio SendGrid.",
			},
		},
	}
}

func (d *ipAddressDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *ipAddressDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ipAddressResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ip := config.IP.ValueString()
	detail, err := d.client.GetIPAddress(ctx, ip)
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP address", err.Error())
		return
	}
	detail.Subusers, err = d.client.ListIPAddressSubusers(ctx, ip)
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP address subusers", err.Error())
		return
	}

	resp.Diagnostics.Append(flattenIPAddressDetail(ctx, detail, &config)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

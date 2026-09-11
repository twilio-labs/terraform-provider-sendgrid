package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-labs/terraform-provider-sendgrid/internal/sendgrid"
)

var _ datasource.DataSource = &partnerOfferingsDataSource{}

func NewPartnerOfferingsDataSource() datasource.DataSource {
	return &partnerOfferingsDataSource{}
}

type partnerOfferingsDataSource struct {
	client *sg.Client
}

type partnerOfferingModel struct {
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	Quantity types.Int64  `tfsdk:"quantity"`
}

type partnerOfferingsDataSourceModel struct {
	ID        types.String           `tfsdk:"id"`
	Offerings []partnerOfferingModel `tfsdk:"offerings"`
}

func (d *partnerOfferingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_partner_offerings"
}

func (d *partnerOfferingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the catalog of package and add-on offerings available to assign to partner " +
			"accounts via `sendgrid_partner_account`. Requires a partner-tier API key.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Always `partner_offerings`.",
			},
			"offerings": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The available offerings.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":     schema.StringAttribute{Computed: true, Description: "The offering's name."},
						"type":     schema.StringAttribute{Computed: true, Description: "The offering type: `package` or `addon`."},
						"quantity": schema.Int64Attribute{Computed: true, Description: "The offering's quantity, where meaningful."},
					},
				},
			},
		},
	}
}

func (d *partnerOfferingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *partnerOfferingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state partnerOfferingsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	offerings, err := d.client.ListPartnerOfferings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading partner offerings", err.Error())
		return
	}

	state.ID = types.StringValue("partner_offerings")
	state.Offerings = make([]partnerOfferingModel, 0, len(offerings))
	for _, o := range offerings {
		state.Offerings = append(state.Offerings, partnerOfferingModel{
			Name:     types.StringValue(o.Name),
			Type:     types.StringValue(o.Type),
			Quantity: types.Int64Value(o.Quantity),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

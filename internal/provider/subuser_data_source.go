package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

var _ datasource.DataSource = &subuserDataSource{}

func NewSubuserDataSource() datasource.DataSource {
	return &subuserDataSource{}
}

type subuserDataSource struct {
	client *sg.Client
}

type subuserDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Username types.String `tfsdk:"username"`
	Email    types.String `tfsdk:"email"`
	Region   types.String `tfsdk:"region"`
	Disabled types.Bool   `tfsdk:"disabled"`
}

func (d *subuserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subuser"
}

func (d *subuserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves an existing SendGrid subuser by username.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The subuser's numeric user ID.",
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "The subuser's username.",
			},
			"email": schema.StringAttribute{
				Computed:    true,
				Description: "The subuser's email address.",
			},
			"region": schema.StringAttribute{
				Computed:    true,
				Description: "The region this subuser is assigned to.",
			},
			"disabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this subuser is disabled.",
			},
		},
	}
}

func (d *subuserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *subuserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config subuserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sub, err := d.client.GetSubuserByUsername(ctx, config.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading subuser", err.Error())
		return
	}
	if sub == nil {
		resp.Diagnostics.AddError("Subuser not found", "No subuser with username \""+config.Username.ValueString()+"\" was found.")
		return
	}

	config.ID = types.StringValue(idFromInt64(sub.ID))
	config.Email = types.StringValue(sub.Email)
	config.Region = types.StringValue(sub.Region)
	config.Disabled = types.BoolValue(sub.Disabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

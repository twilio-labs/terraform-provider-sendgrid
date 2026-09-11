package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-labs/terraform-provider-sendgrid/internal/sendgrid"
)

var _ datasource.DataSource = &teammateDataSource{}

func NewTeammateDataSource() datasource.DataSource {
	return &teammateDataSource{}
}

type teammateDataSource struct {
	client *sg.Client
}

type teammateDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Email    types.String `tfsdk:"email"`
	Scopes   types.Set    `tfsdk:"scopes"`
	IsAdmin  types.Bool   `tfsdk:"is_admin"`
	Username types.String `tfsdk:"username"`
	Status   types.String `tfsdk:"status"`
}

func (d *teammateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teammate"
}

func (d *teammateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves an existing SendGrid teammate, or pending teammate invite, by email address.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Same as `email`.",
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "The teammate's email address.",
			},
			"scopes": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The permission scopes granted to the teammate.",
			},
			"is_admin": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the teammate has admin privileges.",
			},
			"username": schema.StringAttribute{
				Computed:    true,
				Description: "The teammate's username, once they have accepted the invite. Empty while pending.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Either `pending` (invite sent, not yet accepted) or `accepted`.",
			},
		},
	}
}

func (d *teammateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *teammateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config teammateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	email := config.Email.ValueString()
	invite, teammate, err := findTeammateByEmail(ctx, d.client, email)
	if err != nil {
		resp.Diagnostics.AddError("Error reading teammate", err.Error())
		return
	}

	if invite == nil && teammate == nil {
		resp.Diagnostics.AddError("Teammate not found", "No teammate or pending invite with email \""+email+"\" was found.")
		return
	}

	config.ID = types.StringValue(email)

	if invite != nil {
		config.Username = types.StringValue("")
		config.Status = types.StringValue("pending")
		config.IsAdmin = types.BoolValue(invite.IsAdmin)
		resp.Diagnostics.Append(setStringSet(ctx, &config.Scopes, invite.Scopes)...)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	full, err := d.client.GetTeammate(ctx, teammate.Username)
	if err != nil {
		resp.Diagnostics.AddError("Error reading teammate", err.Error())
		return
	}

	config.Username = types.StringValue(full.Username)
	config.Status = types.StringValue("accepted")
	config.IsAdmin = types.BoolValue(full.IsAdmin)
	resp.Diagnostics.Append(setStringSet(ctx, &config.Scopes, full.Scopes)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-labs/terraform-provider-sendgrid/internal/sendgrid"
)

var _ datasource.DataSource = &templateDataSource{}

func NewTemplateDataSource() datasource.DataSource {
	return &templateDataSource{}
}

type templateDataSource struct {
	client *sg.Client
}

type templateVersionSummaryModel struct {
	ID        types.String `tfsdk:"id"`
	Active    types.Bool   `tfsdk:"active"`
	Name      types.String `tfsdk:"name"`
	Subject   types.String `tfsdk:"subject"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

type templateDataSourceModel struct {
	ID         types.String                  `tfsdk:"id"`
	Name       types.String                  `tfsdk:"name"`
	Generation types.String                  `tfsdk:"generation"`
	UpdatedAt  types.String                  `tfsdk:"updated_at"`
	Versions   []templateVersionSummaryModel `tfsdk:"versions"`
}

func (d *templateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_template"
}

func (d *templateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves an existing SendGrid transactional template by ID, including a summary of its versions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The template's UUID.",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of this transactional template.",
			},
			"generation": schema.StringAttribute{
				Computed:    true,
				Description: "Whether this template supports dynamic replacements: `legacy` or `dynamic`.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The time this template was last updated.",
			},
			"versions": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Summary of each version of this template.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The version's UUID.",
						},
						"active": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether this is the active version.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of this version.",
						},
						"subject": schema.StringAttribute{
							Computed:    true,
							Description: "The subject of this version.",
						},
						"updated_at": schema.StringAttribute{
							Computed:    true,
							Description: "The time this version was last updated.",
						},
					},
				},
			},
		},
	}
}

func (d *templateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *templateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config templateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmpl, err := d.client.GetTemplate(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading template", err.Error())
		return
	}

	config.Name = types.StringValue(tmpl.Name)
	config.Generation = types.StringValue(tmpl.Generation)
	config.UpdatedAt = types.StringValue(tmpl.UpdatedAt)
	config.Versions = make([]templateVersionSummaryModel, 0, len(tmpl.Versions))
	for _, v := range tmpl.Versions {
		config.Versions = append(config.Versions, templateVersionSummaryModel{
			ID:        types.StringValue(v.ID),
			Active:    types.BoolValue(v.IsActive()),
			Name:      types.StringValue(v.Name),
			Subject:   types.StringValue(v.Subject),
			UpdatedAt: types.StringValue(v.UpdatedAt),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

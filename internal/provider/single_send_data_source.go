package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-labs/terraform-provider-sendgrid/internal/sendgrid"
)

var _ datasource.DataSource = &singleSendDataSource{}

func NewSingleSendDataSource() datasource.DataSource {
	return &singleSendDataSource{}
}

type singleSendDataSource struct {
	client *sg.Client
}

type singleSendDataSourceModel struct {
	ID          types.String                `tfsdk:"id"`
	Name        types.String                `tfsdk:"name"`
	Categories  types.List                  `tfsdk:"categories"`
	SendAt      types.String                `tfsdk:"send_at"`
	Status      types.String                `tfsdk:"status"`
	SendTo      *singleSendToModel          `tfsdk:"send_to"`
	EmailConfig *singleSendEmailConfigModel `tfsdk:"email_config"`
}

func (d *singleSendDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_single_send"
}

func (d *singleSendDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves an existing SendGrid Marketing Campaigns Single Send by ID.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Required: true, Description: "The Single Send's UUID."},
			"name":       schema.StringAttribute{Computed: true, Description: "The Single Send's name."},
			"categories": schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Categories associated with this Single Send."},
			"send_at":    schema.StringAttribute{Computed: true, Description: "Not populated: SendGrid does not return this value in a directly comparable form. Always empty."},
			"status":     schema.StringAttribute{Computed: true, Description: "The Single Send's status: `draft`, `scheduled`, or `triggered`."},
			"send_to": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The recipients of this Single Send.",
				Attributes: map[string]schema.Attribute{
					"list_ids":    schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Recipient list IDs."},
					"segment_ids": schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Recipient segment IDs."},
					"all":         schema.BoolAttribute{Computed: true, Description: "Whether this Single Send goes to all contacts."},
				},
			},
			"email_config": schema.SingleNestedAttribute{
				Computed:    true,
				Description: "The content of this Single Send.",
				Attributes: map[string]schema.Attribute{
					"subject":                schema.StringAttribute{Computed: true, Description: "The subject line."},
					"html_content":           schema.StringAttribute{Computed: true, Description: "The HTML content."},
					"plain_content":          schema.StringAttribute{Computed: true, Description: "The plain-text content."},
					"generate_plain_content": schema.BoolAttribute{Computed: true, Description: "Whether plain_content is generated from html_content."},
					"design_id":              schema.StringAttribute{Computed: true, Description: "The saved design ID, if any."},
					"editor":                 schema.StringAttribute{Computed: true, Description: "The editor used: `code` or `design`."},
					"suppression_group_id":   schema.Int64Attribute{Computed: true, Description: "The unsubscribe group ID."},
					"custom_unsubscribe_url": schema.StringAttribute{Computed: true, Description: "The custom unsubscribe URL."},
					"sender_id":              schema.Int64Attribute{Computed: true, Description: "The verified sender ID."},
					"ip_pool":                schema.StringAttribute{Computed: true, Description: "The IP pool this Single Send sends from."},
				},
			},
		},
	}
}

func (d *singleSendDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *singleSendDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config singleSendDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := d.client.GetSingleSend(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading single send", err.Error())
		return
	}

	config.ID = types.StringValue(s.ID)
	config.Name = types.StringValue(s.Name)
	config.Status = types.StringValue(s.Status)
	// send_at is a one-time scheduling instruction, not returned by the API
	// in a directly comparable form.
	config.SendAt = types.StringNull()
	resp.Diagnostics.Append(setStringList(ctx, &config.Categories, s.Categories)...)

	config.SendTo = &singleSendToModel{All: types.BoolValue(s.SendTo.All)}
	resp.Diagnostics.Append(setStringList(ctx, &config.SendTo.ListIDs, s.SendTo.ListIDs)...)
	resp.Diagnostics.Append(setStringList(ctx, &config.SendTo.SegmentIDs, s.SendTo.SegmentIDs)...)

	config.EmailConfig = &singleSendEmailConfigModel{
		Subject:              types.StringValue(s.EmailConfig.Subject),
		HTMLContent:          types.StringValue(s.EmailConfig.HTMLContent),
		PlainContent:         types.StringValue(s.EmailConfig.PlainContent),
		GeneratePlainContent: types.BoolValue(s.EmailConfig.GeneratePlainContent),
		DesignID:             types.StringValue(s.EmailConfig.DesignID),
		Editor:               types.StringValue(s.EmailConfig.Editor),
		CustomUnsubscribeURL: types.StringValue(s.EmailConfig.CustomUnsubscribeURL),
		IPPool:               types.StringValue(s.EmailConfig.IPPool),
	}
	if s.EmailConfig.SuppressionGroupID != nil {
		config.EmailConfig.SuppressionGroupID = types.Int64Value(*s.EmailConfig.SuppressionGroupID)
	} else {
		config.EmailConfig.SuppressionGroupID = types.Int64Null()
	}
	if s.EmailConfig.SenderID != nil {
		config.EmailConfig.SenderID = types.Int64Value(*s.EmailConfig.SenderID)
	} else {
		config.EmailConfig.SenderID = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

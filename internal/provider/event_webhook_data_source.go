package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

var _ datasource.DataSource = &eventWebhookDataSource{}

func NewEventWebhookDataSource() datasource.DataSource {
	return &eventWebhookDataSource{}
}

type eventWebhookDataSource struct {
	client *sg.Client
}

type eventWebhookDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	URL              types.String `tfsdk:"url"`
	FriendlyName     types.String `tfsdk:"friendly_name"`
	GroupResubscribe types.Bool   `tfsdk:"group_resubscribe"`
	Delivered        types.Bool   `tfsdk:"delivered"`
	GroupUnsubscribe types.Bool   `tfsdk:"group_unsubscribe"`
	SpamReport       types.Bool   `tfsdk:"spam_report"`
	Bounce           types.Bool   `tfsdk:"bounce"`
	Deferred         types.Bool   `tfsdk:"deferred"`
	Unsubscribe      types.Bool   `tfsdk:"unsubscribe"`
	Processed        types.Bool   `tfsdk:"processed"`
	Open             types.Bool   `tfsdk:"open"`
	Click            types.Bool   `tfsdk:"click"`
	Dropped          types.Bool   `tfsdk:"dropped"`
	OAuthClientID    types.String `tfsdk:"oauth_client_id"`
	OAuthTokenURL    types.String `tfsdk:"oauth_token_url"`
	Signed           types.Bool   `tfsdk:"signed"`
	PublicKey        types.String `tfsdk:"public_key"`
}

func (d *eventWebhookDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_webhook"
}

func eventFlagComputed(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{Computed: true, Description: desc}
}

func (d *eventWebhookDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves an existing SendGrid Event Webhook by ID.",
		Attributes: map[string]schema.Attribute{
			"id":                schema.StringAttribute{Required: true, Description: "The Event Webhook's ID."},
			"enabled":           schema.BoolAttribute{Computed: true, Description: "Whether this Event Webhook is enabled."},
			"url":               schema.StringAttribute{Computed: true, Description: "The URL SendGrid POSTs event data to."},
			"friendly_name":     schema.StringAttribute{Computed: true, Description: "The friendly name for this webhook."},
			"group_resubscribe": eventFlagComputed("Receive group resubscribe events."),
			"delivered":         eventFlagComputed("Receive delivered events."),
			"group_unsubscribe": eventFlagComputed("Receive group unsubscribe events."),
			"spam_report":       eventFlagComputed("Receive spam report events."),
			"bounce":            eventFlagComputed("Receive bounce events."),
			"deferred":          eventFlagComputed("Receive deferred events."),
			"unsubscribe":       eventFlagComputed("Receive unsubscribe events."),
			"processed":         eventFlagComputed("Receive processed events."),
			"open":              eventFlagComputed("Receive open events."),
			"click":             eventFlagComputed("Receive click events."),
			"dropped":           eventFlagComputed("Receive dropped events."),
			"oauth_client_id":   schema.StringAttribute{Computed: true, Description: "OAuth client ID configured for this webhook."},
			"oauth_token_url":   schema.StringAttribute{Computed: true, Description: "OAuth token URL configured for this webhook."},
			"signed":            schema.BoolAttribute{Computed: true, Description: "Whether signature verification is enabled."},
			"public_key":        schema.StringAttribute{Computed: true, Description: "The public key used to verify request signatures, if signed."},
		},
	}
}

func (d *eventWebhookDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *eventWebhookDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config eventWebhookDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, err := d.client.GetEventWebhook(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading event webhook", err.Error())
		return
	}

	config.ID = types.StringValue(w.ID)
	config.Enabled = types.BoolValue(w.Enabled)
	config.URL = types.StringValue(w.URL)
	config.FriendlyName = types.StringValue(w.FriendlyName)
	config.GroupResubscribe = types.BoolValue(w.GroupResubscribe)
	config.Delivered = types.BoolValue(w.Delivered)
	config.GroupUnsubscribe = types.BoolValue(w.GroupUnsubscribe)
	config.SpamReport = types.BoolValue(w.SpamReport)
	config.Bounce = types.BoolValue(w.Bounce)
	config.Deferred = types.BoolValue(w.Deferred)
	config.Unsubscribe = types.BoolValue(w.Unsubscribe)
	config.Processed = types.BoolValue(w.Processed)
	config.Open = types.BoolValue(w.Open)
	config.Click = types.BoolValue(w.Click)
	config.Dropped = types.BoolValue(w.Dropped)
	config.OAuthClientID = types.StringValue(w.OAuthClientID)
	config.OAuthTokenURL = types.StringValue(w.OAuthTokenURL)

	publicKey, err := d.client.GetEventWebhookPublicKey(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading event webhook signature status", err.Error())
		return
	}
	config.PublicKey = types.StringValue(publicKey)
	config.Signed = types.BoolValue(publicKey != "")

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

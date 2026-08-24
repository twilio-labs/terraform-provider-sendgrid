package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

var (
	_ resource.Resource                = &eventWebhookResource{}
	_ resource.ResourceWithImportState = &eventWebhookResource{}
)

func NewEventWebhookResource() resource.Resource {
	return &eventWebhookResource{}
}

type eventWebhookResource struct {
	client *sg.Client
}

type eventWebhookResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	URL               types.String `tfsdk:"url"`
	FriendlyName      types.String `tfsdk:"friendly_name"`
	GroupResubscribe  types.Bool   `tfsdk:"group_resubscribe"`
	Delivered         types.Bool   `tfsdk:"delivered"`
	GroupUnsubscribe  types.Bool   `tfsdk:"group_unsubscribe"`
	SpamReport        types.Bool   `tfsdk:"spam_report"`
	Bounce            types.Bool   `tfsdk:"bounce"`
	Deferred          types.Bool   `tfsdk:"deferred"`
	Unsubscribe       types.Bool   `tfsdk:"unsubscribe"`
	Processed         types.Bool   `tfsdk:"processed"`
	Open              types.Bool   `tfsdk:"open"`
	Click             types.Bool   `tfsdk:"click"`
	Dropped           types.Bool   `tfsdk:"dropped"`
	OAuthClientID     types.String `tfsdk:"oauth_client_id"`
	OAuthClientSecret types.String `tfsdk:"oauth_client_secret"`
	OAuthTokenURL     types.String `tfsdk:"oauth_token_url"`
	Signed            types.Bool   `tfsdk:"signed"`
	PublicKey         types.String `tfsdk:"public_key"`
}

func eventFlag(desc string) schema.BoolAttribute {
	return schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Default:     booldefault.StaticBool(false),
		Description: desc,
	}
}

func (r *eventWebhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_event_webhook"
}

func (r *eventWebhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a SendGrid Event Webhook configuration. Accounts on plans that support " +
			"multiple webhooks may have more than one `sendgrid_event_webhook` resource, each with a " +
			"distinct `url`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The Event Webhook's ID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether this Event Webhook is enabled.",
			},
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The URL SendGrid will POST event data to. No two webhooks may share a URL.",
			},
			"friendly_name": schema.StringAttribute{
				Optional:    true,
				Description: "An optional, non-unique friendly name for this webhook.",
			},
			"group_resubscribe": eventFlag("Receive group resubscribe events. Requires Subscription Tracking."),
			"delivered":         eventFlag("Receive delivered events."),
			"group_unsubscribe": eventFlag("Receive group unsubscribe events. Requires Subscription Tracking."),
			"spam_report":       eventFlag("Receive spam report events."),
			"bounce":            eventFlag("Receive bounce events."),
			"deferred":          eventFlag("Receive deferred events."),
			"unsubscribe":       eventFlag("Receive unsubscribe events. Requires Subscription Tracking."),
			"processed":         eventFlag("Receive processed events."),
			"open":              eventFlag("Receive open events. Requires Open Tracking."),
			"click":             eventFlag("Receive click events. Requires Click Tracking."),
			"dropped":           eventFlag("Receive dropped events."),
			"oauth_client_id": schema.StringAttribute{
				Optional:    true,
				Description: "OAuth client ID SendGrid uses to generate an access token for this webhook.",
			},
			"oauth_client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "OAuth client secret. Write-only: SendGrid never returns this value back.",
			},
			"oauth_token_url": schema.StringAttribute{
				Optional:    true,
				Description: "URL SendGrid sends the OAuth client ID and secret to in order to generate an access token.",
			},
			"signed": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether signature verification is enabled for this webhook.",
			},
			"public_key": schema.StringAttribute{
				Computed:    true,
				Description: "The public key used to verify this webhook's request signatures. Only set when `signed` is `true`.",
			},
		},
	}
}

func (r *eventWebhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

func eventWebhookParamsFromModel(m eventWebhookResourceModel) sg.EventWebhookParams {
	return sg.EventWebhookParams{
		Enabled:           m.Enabled.ValueBool(),
		URL:               m.URL.ValueString(),
		FriendlyName:      m.FriendlyName.ValueString(),
		GroupResubscribe:  m.GroupResubscribe.ValueBool(),
		Delivered:         m.Delivered.ValueBool(),
		GroupUnsubscribe:  m.GroupUnsubscribe.ValueBool(),
		SpamReport:        m.SpamReport.ValueBool(),
		Bounce:            m.Bounce.ValueBool(),
		Deferred:          m.Deferred.ValueBool(),
		Unsubscribe:       m.Unsubscribe.ValueBool(),
		Processed:         m.Processed.ValueBool(),
		Open:              m.Open.ValueBool(),
		Click:             m.Click.ValueBool(),
		Dropped:           m.Dropped.ValueBool(),
		OAuthClientID:     m.OAuthClientID.ValueString(),
		OAuthClientSecret: m.OAuthClientSecret.ValueString(),
		OAuthTokenURL:     m.OAuthTokenURL.ValueString(),
	}
}

func flattenEventWebhook(w *sg.EventWebhook, m *eventWebhookResourceModel) {
	m.ID = types.StringValue(w.ID)
	m.Enabled = types.BoolValue(w.Enabled)
	m.URL = types.StringValue(w.URL)
	if w.FriendlyName != "" {
		m.FriendlyName = types.StringValue(w.FriendlyName)
	}
	m.GroupResubscribe = types.BoolValue(w.GroupResubscribe)
	m.Delivered = types.BoolValue(w.Delivered)
	m.GroupUnsubscribe = types.BoolValue(w.GroupUnsubscribe)
	m.SpamReport = types.BoolValue(w.SpamReport)
	m.Bounce = types.BoolValue(w.Bounce)
	m.Deferred = types.BoolValue(w.Deferred)
	m.Unsubscribe = types.BoolValue(w.Unsubscribe)
	m.Processed = types.BoolValue(w.Processed)
	m.Open = types.BoolValue(w.Open)
	m.Click = types.BoolValue(w.Click)
	m.Dropped = types.BoolValue(w.Dropped)
	if w.OAuthClientID != "" {
		m.OAuthClientID = types.StringValue(w.OAuthClientID)
	}
	if w.OAuthTokenURL != "" {
		m.OAuthTokenURL = types.StringValue(w.OAuthTokenURL)
	}
}

func (r *eventWebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan eventWebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, err := r.client.CreateEventWebhook(ctx, eventWebhookParamsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error creating event webhook", err.Error())
		return
	}
	flattenEventWebhook(w, &plan)

	if plan.Signed.ValueBool() {
		publicKey, err := r.client.SetEventWebhookSigned(ctx, w.ID, true)
		if err != nil {
			resp.Diagnostics.AddError("Error enabling signature verification", err.Error())
			return
		}
		plan.PublicKey = types.StringValue(publicKey)
	} else {
		plan.PublicKey = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *eventWebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state eventWebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, err := r.client.GetEventWebhook(ctx, state.ID.ValueString())
	if err != nil {
		if sg.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading event webhook", err.Error())
		return
	}
	flattenEventWebhook(w, &state)

	publicKey, err := r.client.GetEventWebhookPublicKey(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading event webhook signature status", err.Error())
		return
	}
	state.PublicKey = types.StringValue(publicKey)
	state.Signed = types.BoolValue(publicKey != "")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *eventWebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state eventWebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	w, err := r.client.UpdateEventWebhook(ctx, state.ID.ValueString(), eventWebhookParamsFromModel(plan))
	if err != nil {
		resp.Diagnostics.AddError("Error updating event webhook", err.Error())
		return
	}
	plan.ID = state.ID
	flattenEventWebhook(w, &plan)

	if plan.Signed.ValueBool() != state.Signed.ValueBool() {
		publicKey, err := r.client.SetEventWebhookSigned(ctx, state.ID.ValueString(), plan.Signed.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Error updating signature verification", err.Error())
			return
		}
		plan.PublicKey = types.StringValue(publicKey)
	} else {
		plan.PublicKey = state.PublicKey
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *eventWebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state eventWebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteEventWebhook(ctx, state.ID.ValueString()); err != nil && !sg.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting event webhook", err.Error())
	}
}

func (r *eventWebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

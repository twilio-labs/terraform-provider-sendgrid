package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-labs/terraform-provider-sendgrid/internal/sendgrid"
)

var (
	_ resource.Resource                = &singleSendResource{}
	_ resource.ResourceWithImportState = &singleSendResource{}
)

func NewSingleSendResource() resource.Resource {
	return &singleSendResource{}
}

type singleSendResource struct {
	client *sg.Client
}

type singleSendToModel struct {
	ListIDs    types.List `tfsdk:"list_ids"`
	SegmentIDs types.List `tfsdk:"segment_ids"`
	All        types.Bool `tfsdk:"all"`
}

type singleSendEmailConfigModel struct {
	Subject              types.String `tfsdk:"subject"`
	HTMLContent          types.String `tfsdk:"html_content"`
	PlainContent         types.String `tfsdk:"plain_content"`
	GeneratePlainContent types.Bool   `tfsdk:"generate_plain_content"`
	DesignID             types.String `tfsdk:"design_id"`
	Editor               types.String `tfsdk:"editor"`
	SuppressionGroupID   types.Int64  `tfsdk:"suppression_group_id"`
	CustomUnsubscribeURL types.String `tfsdk:"custom_unsubscribe_url"`
	SenderID             types.Int64  `tfsdk:"sender_id"`
	IPPool               types.String `tfsdk:"ip_pool"`
}

type singleSendResourceModel struct {
	ID          types.String                `tfsdk:"id"`
	Name        types.String                `tfsdk:"name"`
	Categories  types.List                  `tfsdk:"categories"`
	SendAt      types.String                `tfsdk:"send_at"`
	Status      types.String                `tfsdk:"status"`
	SendTo      *singleSendToModel          `tfsdk:"send_to"`
	EmailConfig *singleSendEmailConfigModel `tfsdk:"email_config"`
}

func (r *singleSendResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_single_send"
}

func (r *singleSendResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a SendGrid Marketing Campaigns Single Send: a one-time campaign email.\n\n" +
			"~> Setting `send_at` to a real timestamp schedules the Single Send to actually go out at that " +
			"time — this is a real side effect, not just metadata. Setting it to `\"now\"` sends it " +
			"immediately. Once a Single Send has been sent, it can no longer be updated or scheduled again " +
			"through this resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The Single Send's UUID.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The Single Send's name.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 100),
				},
			},
			"categories": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Categories to associate with this Single Send, for filtering statistics. Up to 10.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"send_at": schema.StringAttribute{
				Optional: true,
				Description: "An ISO 8601 date-time to schedule this Single Send for, or the literal string " +
					"`\"now\"` to send it immediately. Leave unset to keep it as an unscheduled draft.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "The Single Send's status: `draft`, `scheduled`, or `triggered`.",
			},
			"send_to": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The recipients of this Single Send.",
				Attributes: map[string]schema.Attribute{
					"list_ids": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Description: "Recipient list IDs. Up to 50.",
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
					},
					"segment_ids": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						ElementType: types.StringType,
						Description: "Recipient segment IDs. Up to 10.",
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
					},
					"all": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Send to all contacts. If false, at least one of `list_ids` or `segment_ids` is required before scheduling.",
					},
				},
			},
			"email_config": schema.SingleNestedAttribute{
				Required:    true,
				Description: "The content of this Single Send.",
				Attributes: map[string]schema.Attribute{
					"subject": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The subject line. Omit when using `design_id`.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"html_content": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The HTML content. Omit when using `design_id`.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"plain_content": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The plain-text content. Generated from `html_content` unless `generate_plain_content` is `false`.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"generate_plain_content": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
						Description: "If true, `plain_content` is always regenerated from `html_content`.",
					},
					"design_id": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "A saved design ID, used in place of `html_content`/`plain_content`/`subject`.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"editor": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("code"),
						Description: "The editor used in the SendGrid UI: `code` or `design`.",
						Validators: []validator.String{
							stringvalidator.OneOf("code", "design"),
						},
					},
					"suppression_group_id": schema.Int64Attribute{
						Optional:    true,
						Description: "The unsubscribe group ID. Required unless `custom_unsubscribe_url` is set.",
					},
					"custom_unsubscribe_url": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "A custom unsubscribe URL. Required unless `suppression_group_id` is set.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"sender_id": schema.Int64Attribute{
						Optional:    true,
						Description: "The verified sender ID.",
					},
					"ip_pool": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "The name of the IP pool to send from.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
		},
	}
}

func (r *singleSendResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	r.client = client
}

// singleSendParamsFromModel builds API params from m. send_to and
// email_config are schema.Required, so the framework's own config
// validation prevents Create/Update from ever running with either pointer
// nil in normal operation — the nil checks here are cheap local insurance
// against that invariant living only in a separate file's schema.
func singleSendParamsFromModel(ctx context.Context, m singleSendResourceModel) (sg.SingleSendParams, diag.Diagnostics) {
	var diags diag.Diagnostics

	sendTo := m.SendTo
	if sendTo == nil {
		sendTo = &singleSendToModel{}
	}
	emailConfig := m.EmailConfig
	if emailConfig == nil {
		emailConfig = &singleSendEmailConfigModel{}
	}

	categories, d := stringListToSlice(ctx, m.Categories)
	diags.Append(d...)
	listIDs, d := stringListToSlice(ctx, sendTo.ListIDs)
	diags.Append(d...)
	segmentIDs, d := stringListToSlice(ctx, sendTo.SegmentIDs)
	diags.Append(d...)

	p := sg.SingleSendParams{
		Name:       m.Name.ValueString(),
		Categories: categories,
		SendTo: sg.SingleSendTo{
			ListIDs:    listIDs,
			SegmentIDs: segmentIDs,
			All:        sendTo.All.ValueBool(),
		},
		EmailConfig: sg.SingleSendEmailConfig{
			Subject:              emailConfig.Subject.ValueString(),
			HTMLContent:          emailConfig.HTMLContent.ValueString(),
			PlainContent:         emailConfig.PlainContent.ValueString(),
			GeneratePlainContent: emailConfig.GeneratePlainContent.ValueBool(),
			DesignID:             emailConfig.DesignID.ValueString(),
			Editor:               emailConfig.Editor.ValueString(),
			CustomUnsubscribeURL: emailConfig.CustomUnsubscribeURL.ValueString(),
			IPPool:               emailConfig.IPPool.ValueString(),
		},
	}
	if !emailConfig.SuppressionGroupID.IsNull() && !emailConfig.SuppressionGroupID.IsUnknown() {
		v := emailConfig.SuppressionGroupID.ValueInt64()
		p.EmailConfig.SuppressionGroupID = &v
	}
	if !emailConfig.SenderID.IsNull() && !emailConfig.SenderID.IsUnknown() {
		v := emailConfig.SenderID.ValueInt64()
		p.EmailConfig.SenderID = &v
	}
	return p, diags
}

// sgOpenTrackPlaceholder is SendGrid's open-tracking placeholder tag. If a
// Single Send's html_content doesn't already contain it, SendGrid silently
// prepends it server-side — confirmed against a live account. Left as-is,
// this crashes every apply with "Provider produced inconsistent result"
// (Terraform sees the plan's literal html_content diverge from what Create
// returns). Since a user placing this exact token at the very start of
// their content on purpose is vanishingly unlikely compared to SendGrid
// having added it, strip it unconditionally when present as a prefix.
const sgOpenTrackPlaceholder = "%sg_open_track%"

func stripInjectedOpenTrackPlaceholder(htmlContent string) string {
	return strings.TrimPrefix(htmlContent, sgOpenTrackPlaceholder)
}

func flattenSingleSend(ctx context.Context, s *sg.SingleSend, m *singleSendResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(s.ID)
	m.Name = types.StringValue(s.Name)
	m.Status = types.StringValue(s.Status)
	diags.Append(setStringList(ctx, &m.Categories, s.Categories)...)

	if m.SendTo == nil {
		m.SendTo = &singleSendToModel{}
	}
	diags.Append(setStringList(ctx, &m.SendTo.ListIDs, s.SendTo.ListIDs)...)
	diags.Append(setStringList(ctx, &m.SendTo.SegmentIDs, s.SendTo.SegmentIDs)...)
	m.SendTo.All = types.BoolValue(s.SendTo.All)

	if m.EmailConfig == nil {
		m.EmailConfig = &singleSendEmailConfigModel{}
	}
	m.EmailConfig.Subject = types.StringValue(s.EmailConfig.Subject)
	m.EmailConfig.HTMLContent = types.StringValue(stripInjectedOpenTrackPlaceholder(s.EmailConfig.HTMLContent))
	m.EmailConfig.PlainContent = types.StringValue(s.EmailConfig.PlainContent)
	m.EmailConfig.GeneratePlainContent = types.BoolValue(s.EmailConfig.GeneratePlainContent)
	m.EmailConfig.DesignID = types.StringValue(s.EmailConfig.DesignID)
	m.EmailConfig.Editor = types.StringValue(s.EmailConfig.Editor)
	m.EmailConfig.CustomUnsubscribeURL = types.StringValue(s.EmailConfig.CustomUnsubscribeURL)
	m.EmailConfig.IPPool = types.StringValue(s.EmailConfig.IPPool)
	if s.EmailConfig.SuppressionGroupID != nil {
		m.EmailConfig.SuppressionGroupID = types.Int64Value(*s.EmailConfig.SuppressionGroupID)
	} else {
		m.EmailConfig.SuppressionGroupID = types.Int64Null()
	}
	if s.EmailConfig.SenderID != nil {
		m.EmailConfig.SenderID = types.Int64Value(*s.EmailConfig.SenderID)
	} else {
		m.EmailConfig.SenderID = types.Int64Null()
	}

	return diags
}

func (r *singleSendResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan singleSendResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wantSendAt := plan.SendAt.ValueString()

	params, diags := singleSendParamsFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.CreateSingleSend(ctx, params)
	if err != nil {
		resp.Diagnostics.AddError("Error creating single send", err.Error())
		return
	}

	// The Single Send now exists remotely as an unscheduled draft. Persist
	// that fact before attempting to schedule it below, so a failure there
	// doesn't leave Terraform unable to track (and destroy) the draft, and
	// so state accurately reflects "not scheduled" if scheduling fails.
	resp.Diagnostics.Append(flattenSingleSend(ctx, s, &plan)...)
	plan.SendAt = types.StringNull()
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if wantSendAt != "" {
		status, err := r.client.ScheduleSingleSend(ctx, s.ID, wantSendAt)
		if err != nil {
			resp.Diagnostics.AddError("Error scheduling single send", err.Error())
			return
		}
		plan.SendAt = types.StringValue(wantSendAt)
		plan.Status = types.StringValue(status)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *singleSendResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state singleSendResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.GetSingleSend(ctx, state.ID.ValueString())
	if err != nil {
		if sg.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading single send", err.Error())
		return
	}

	sendAt := state.SendAt
	resp.Diagnostics.Append(flattenSingleSend(ctx, s, &state)...)
	// send_at is a one-time scheduling instruction, not returned by the API
	// in a directly comparable form; preserve what's already in state.
	state.SendAt = sendAt
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *singleSendResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state singleSendResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wantSendAt := plan.SendAt.ValueString()
	curSendAt := state.SendAt.ValueString()

	params, diags := singleSendParamsFromModel(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	s, err := r.client.UpdateSingleSend(ctx, state.ID.ValueString(), params)
	if err != nil {
		resp.Diagnostics.AddError("Error updating single send", err.Error())
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(flattenSingleSend(ctx, s, &plan)...)
	// The content update above is confirmed-applied; checkpoint send_at and
	// status at their last-known-good (unchanged) values before attempting
	// the separate schedule/unschedule call below, so a failure there
	// doesn't commit an unapplied schedule change to state.
	plan.SendAt = state.SendAt
	plan.Status = state.Status
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if wantSendAt != curSendAt {
		if wantSendAt == "" {
			if err := r.client.UnscheduleSingleSend(ctx, state.ID.ValueString()); err != nil {
				resp.Diagnostics.AddError("Error unscheduling single send", err.Error())
				return
			}
			plan.SendAt = types.StringNull()
			plan.Status = types.StringValue("draft")
		} else {
			status, err := r.client.ScheduleSingleSend(ctx, state.ID.ValueString(), wantSendAt)
			if err != nil {
				resp.Diagnostics.AddError("Error scheduling single send", err.Error())
				return
			}
			plan.SendAt = types.StringValue(wantSendAt)
			plan.Status = types.StringValue(status)
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
}

func (r *singleSendResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state singleSendResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteSingleSend(ctx, state.ID.ValueString()); err != nil && !sg.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting single send", err.Error())
	}
}

func (r *singleSendResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

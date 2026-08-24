package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

var _ datasource.DataSource = &domainAuthenticationDataSource{}

func NewDomainAuthenticationDataSource() datasource.DataSource {
	return &domainAuthenticationDataSource{}
}

type domainAuthenticationDataSource struct {
	client *sg.Client
}

type domainAuthenticationDataSourceModel struct {
	ID                types.String `tfsdk:"id"`
	Domain            types.String `tfsdk:"domain"`
	Subdomain         types.String `tfsdk:"subdomain"`
	Username          types.String `tfsdk:"username"`
	IPs               types.List   `tfsdk:"ips"`
	CustomSPF         types.Bool   `tfsdk:"custom_spf"`
	Default           types.Bool   `tfsdk:"default"`
	AutomaticSecurity types.Bool   `tfsdk:"automatic_security"`
	Region            types.String `tfsdk:"region"`
	Valid             types.Bool   `tfsdk:"valid"`
	DNSRecords        types.Map    `tfsdk:"dns_records"`
}

func (d *domainAuthenticationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_authentication"
}

func (d *domainAuthenticationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves an existing SendGrid authenticated domain by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The authenticated domain's numeric ID.",
			},
			"domain": schema.StringAttribute{
				Computed:    true,
				Description: "The domain being authenticated.",
			},
			"subdomain": schema.StringAttribute{
				Computed:    true,
				Description: "The subdomain used for this authenticated domain's DNS records.",
			},
			"username": schema.StringAttribute{
				Computed:    true,
				Description: "The username this authenticated domain is associated with.",
			},
			"ips": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "IP addresses included in this domain's custom SPF record.",
			},
			"custom_spf": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this domain uses a custom SPF record.",
			},
			"default": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this is the default authenticated domain.",
			},
			"automatic_security": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether SendGrid automatically manages this domain's SPF record and DKIM keys.",
			},
			"region": schema.StringAttribute{
				Computed:    true,
				Description: "The region of the domain.",
			},
			"valid": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether SendGrid has validated the required DNS records for this domain.",
			},
			"dns_records": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "The DNS records required to authenticate this domain.",
			},
		},
	}
}

func (d *domainAuthenticationDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	client, diags := clientFromProviderData(req.ProviderData)
	resp.Diagnostics.Append(diags...)
	d.client = client
}

func (d *domainAuthenticationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config domainAuthenticationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := d.client.GetDomainAuthentication(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading domain authentication", err.Error())
		return
	}

	config.ID = types.StringValue(idFromInt64(domain.ID))
	config.Domain = types.StringValue(domain.Domain)
	config.Subdomain = types.StringValue(domain.Subdomain)
	config.Username = types.StringValue(domain.Username)
	config.CustomSPF = types.BoolValue(domain.CustomSPF)
	config.Default = types.BoolValue(domain.Default)
	config.AutomaticSecurity = types.BoolValue(domain.AutomaticSecurity)
	config.Valid = types.BoolValue(domain.Valid)
	config.Region = types.StringValue(domain.Region)

	resp.Diagnostics.Append(setStringList(ctx, &config.IPs, domain.IPs)...)

	records := map[string]string{
		"mail_cname_type": domain.DNS.MailCNAME.Type,
		"mail_cname_host": domain.DNS.MailCNAME.Host,
		"mail_cname_data": domain.DNS.MailCNAME.Data,
		"dkim1_type":      domain.DNS.DKIM1.Type,
		"dkim1_host":      domain.DNS.DKIM1.Host,
		"dkim1_data":      domain.DNS.DKIM1.Data,
		"dkim2_type":      domain.DNS.DKIM2.Type,
		"dkim2_host":      domain.DNS.DKIM2.Host,
		"dkim2_data":      domain.DNS.DKIM2.Data,
	}
	dnsMap, diags := types.MapValueFrom(ctx, types.StringType, records)
	resp.Diagnostics.Append(diags...)
	config.DNSRecords = dnsMap

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

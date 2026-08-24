package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	sg "github.com/twilio-internal/terraform-provider-sendgrid/internal/sendgrid"
)

// Ensure the implementation satisfies the expected interfaces.
var _ provider.Provider = &SendgridProvider{}

// SendgridProvider is the Terraform provider implementation for SendGrid.
type SendgridProvider struct {
	// version is set by the build process via goreleaser ldflags, and is
	// used to set the User-Agent header on outgoing API requests.
	version string
}

// SendgridProviderModel describes the provider configuration block.
type SendgridProviderModel struct {
	APIKey     types.String `tfsdk:"api_key"`
	Host       types.String `tfsdk:"host"`
	OnBehalfOf types.String `tfsdk:"on_behalf_of"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SendgridProvider{version: version}
	}
}

func (p *SendgridProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "sendgrid"
	resp.Version = p.version
}

func (p *SendgridProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with the SendGrid v3 API.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "SendGrid API key. Can also be set via the `SENDGRID_API_KEY` environment variable.",
			},
			"host": schema.StringAttribute{
				Optional: true,
				Description: "SendGrid API host. Defaults to `https://api.sendgrid.com`; set to " +
					"`https://api.eu.sendgrid.com` for EU regional accounts. Can also be set via the " +
					"`SENDGRID_API_HOST` environment variable.",
			},
			"on_behalf_of": schema.StringAttribute{
				Optional: true,
				Description: "Sets the `on-behalf-of` header on every request, causing all requests to act " +
					"as the given subuser or account (e.g. `account-id <account-id>`). Can also be set via the " +
					"`SENDGRID_ON_BEHALF_OF` environment variable. Can be overridden per-resource with that " +
					"resource's own `on_behalf_of` attribute, where supported.",
			},
		},
	}
}

func (p *SendgridProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config SendgridProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := config.APIKey.ValueString()
	if apiKey == "" {
		apiKey = os.Getenv("SENDGRID_API_KEY")
	}
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing SendGrid API Key",
			"The provider requires an API key, set via the `api_key` provider attribute or the "+
				"`SENDGRID_API_KEY` environment variable.",
		)
		return
	}

	host := config.Host.ValueString()
	if host == "" {
		host = os.Getenv("SENDGRID_API_HOST")
	}

	onBehalfOf := config.OnBehalfOf.ValueString()
	if onBehalfOf == "" {
		onBehalfOf = os.Getenv("SENDGRID_ON_BEHALF_OF")
	}

	client := sg.NewClient(apiKey, sg.WithHost(host), sg.WithOnBehalfOf(onBehalfOf))

	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *SendgridProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAPIKeyResource,
		NewSubuserResource,
		NewDomainAuthenticationResource,
		NewTemplateResource,
		NewTemplateVersionResource,
		NewEventWebhookResource,
	}
}

func (p *SendgridProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAPIKeyDataSource,
		NewSubuserDataSource,
		NewDomainAuthenticationDataSource,
		NewTemplateDataSource,
		NewEventWebhookDataSource,
	}
}

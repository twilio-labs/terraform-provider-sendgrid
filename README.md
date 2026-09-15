# Terraform Provider for SendGrid

A Terraform provider for managing [SendGrid](https://sendgrid.com) (Twilio) resources, built on the
[Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).


## Resources

**Access & Security**
- [`sendgrid_api_key`](docs/resources/api_key.md)

**Account Management**
- [`sendgrid_subuser`](docs/resources/subuser.md)

**Sender Authentication**
- [`sendgrid_domain_authentication`](docs/resources/domain_authentication.md)

**Templates**
- [`sendgrid_template`](docs/resources/template.md)
- [`sendgrid_template_version`](docs/resources/template_version.md)

**Webhooks**
- [`sendgrid_event_webhook`](docs/resources/event_webhook.md)

**Marketing Campaigns**
- [`sendgrid_list`](docs/resources/list.md)
- [`sendgrid_custom_field`](docs/resources/custom_field.md)
- [`sendgrid_segment`](docs/resources/segment.md)
- [`sendgrid_single_send`](docs/resources/single_send.md)

**Tracking & Mail Settings**
- [`sendgrid_click_tracking_settings`](docs/resources/click_tracking_settings.md)
- [`sendgrid_open_tracking_settings`](docs/resources/open_tracking_settings.md)
- [`sendgrid_subscription_tracking_settings`](docs/resources/subscription_tracking_settings.md)
- [`sendgrid_google_analytics_settings`](docs/resources/google_analytics_settings.md)

**IP Management**
- [`sendgrid_ip_pool`](docs/resources/ip_pool.md)

Each resource has a matching data source of the same name, except `sendgrid_template_version` and the four
singleton Tracking & Mail Settings resources (there's nothing to look up by ID for an account-wide setting).
See [docs/](docs/) for full reference documentation, and
[docs/guides/authentication.md](docs/guides/authentication.md) for provider configuration (including EU data
residency and On-Behalf-Of subuser access).

## Development

Requirements:

- [Go](https://golang.org/doc/install) >= 1.24
- [Terraform](https://www.terraform.io/downloads.html) >= 1.11

```sh
go install
```

Or run everything CI runs in one shot (build, vet, lint, unit tests, docs generation):

```sh
make check
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full workflow

### Configuration for local development

Point Terraform at a locally-built copy of the provider using a
[dev override](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers)
in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "twilio-labs/sendgrid" = "/path/to/your/GOBIN"
  }
  direct {}
}
```

Then set credentials via environment variables (or the equivalent provider block attributes — see
[docs/guides/authentication.md](docs/guides/authentication.md)):

| Environment variable    | Purpose                                                                 |
| ------------------------ | ------------------------------------------------------------------------ |
| `SENDGRID_API_KEY`       | Required. The SendGrid API key used to authenticate all requests.       |
| `SENDGRID_API_HOST`      | Optional. Defaults to `https://api.sendgrid.com`; set for EU regional accounts. |
| `SENDGRID_ON_BEHALF_OF`  | Optional. Acts on behalf of a subuser or parent account on every request. |

### Testing

Unit tests (no credentials required):

```sh
go test ./...
```

Acceptance tests make real API calls against a live SendGrid account and will create/destroy real resources.
They're gated behind `TF_ACC` and require a real API key:

```sh
TF_ACC=1 SENDGRID_API_KEY=SG.xxxx go test ./internal/provider/... -v
```

### Generating documentation

Docs under `docs/` are generated from resource/data-source schema descriptions, `examples/`, and
`templates/` via [`tfplugindocs`](https://github.com/hashicorp/terraform-plugin-docs). After changing a
schema description or adding an example, regenerate with:

```sh
go generate ./...
```

## License

[Mozilla Public License v2.0](LICENSE)

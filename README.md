# SendGrid Terraform Provider

A Terraform provider for managing [SendGrid](https://sendgrid.com) (Twilio) resources, built on the
[Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).


## Usage

The provider is published on the [Terraform Registry](https://registry.terraform.io/providers/twilio-labs/sendgrid/latest)
as `twilio-labs/sendgrid`:

```hcl
terraform {
  required_providers {
    sendgrid = {
      source = "twilio-labs/sendgrid"
    }
  }
}

provider "sendgrid" {
  # Or set the SENDGRID_API_KEY environment variable.
  api_key = var.sendgrid_api_key
}
```

See [docs/guides/authentication.md](docs/guides/authentication.md) for all provider configuration options
(including EU data residency and On-Behalf-Of subuser access), and
[docs/guides/migrating-from-other-providers.md](docs/guides/migrating-from-other-providers.md) if you're
coming from another SendGrid provider.

## Resources

**Access & Security**
- [`sendgrid_api_key`](docs/resources/api_key.md)

**Account Management**
- [`sendgrid_subuser`](docs/resources/subuser.md)
- [`sendgrid_teammate`](docs/resources/teammate.md)

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
- [`sendgrid_address_whitelist_mail_settings`](docs/resources/address_whitelist_mail_settings.md)
- [`sendgrid_bounce_purge_mail_settings`](docs/resources/bounce_purge_mail_settings.md)
- [`sendgrid_footer_mail_settings`](docs/resources/footer_mail_settings.md)
- [`sendgrid_forward_bounce_mail_settings`](docs/resources/forward_bounce_mail_settings.md)
- [`sendgrid_forward_spam_mail_settings`](docs/resources/forward_spam_mail_settings.md)
- [`sendgrid_template_mail_settings`](docs/resources/template_mail_settings.md)

**IP Management**
- [`sendgrid_ip_address`](docs/resources/ip_address.md)
- [`sendgrid_ip_pool`](docs/resources/ip_pool.md)
- [`sendgrid_ip_warmup`](docs/resources/ip_warmup.md)

**Partner Accounts**
- [`sendgrid_partner_account`](docs/resources/partner_account.md)

Most resources have a matching data source of the same name. The exceptions are `sendgrid_template_version`,
`sendgrid_ip_warmup`, `sendgrid_partner_account`, and the account-wide Tracking & Mail Settings resources
(there's nothing to look up by ID for a singleton setting). There is also a standalone
[`sendgrid_partner_offerings`](docs/data-sources/partner_offerings.md) data source. See [docs/](docs/) for
full reference documentation.

## Development

Requirements:

- [Go](https://golang.org/doc/install) >= 1.26 (see `go.mod`)
- [Terraform](https://www.terraform.io/downloads.html) >= 1.9

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

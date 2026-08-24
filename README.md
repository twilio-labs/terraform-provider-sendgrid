# Terraform Provider for SendGrid

A Terraform provider for managing [SendGrid](https://sendgrid.com) (Twilio) resources, built on the
[Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).

**Status:** early development — not yet published to the Terraform Registry.

## Resources

- [`sendgrid_api_key`](docs/resources/api_key.md)
- [`sendgrid_subuser`](docs/resources/subuser.md)
- [`sendgrid_domain_authentication`](docs/resources/domain_authentication.md)
- [`sendgrid_template`](docs/resources/template.md)
- [`sendgrid_template_version`](docs/resources/template_version.md)
- [`sendgrid_event_webhook`](docs/resources/event_webhook.md)

Each resource has a matching data source of the same name (except `sendgrid_template_version`). See
[docs/](docs/) for full reference documentation, and [docs/guides/authentication.md](docs/guides/authentication.md)
for provider configuration (including EU data residency and On-Behalf-Of subuser access).

More resources (Marketing Campaigns/Contacts, full Mail/Tracking Settings, IP address management) are not yet
implemented — see open issues for the current backlog.

## Development

Requirements:

- [Go](https://golang.org/doc/install) >= 1.24
- [Terraform](https://www.terraform.io/downloads.html) >= 1.11

```sh
go install
```

### Configuration for local development

Point Terraform at a locally-built copy of the provider using a
[dev override](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers)
in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "twilio-internal/sendgrid" = "/path/to/your/GOBIN"
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

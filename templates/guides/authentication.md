---
page_title: "Authentication"
subcategory: ""
description: |-
  Configuring credentials, EU data residency, and On-Behalf-Of access for the SendGrid provider.
---

# Authentication

The SendGrid provider requires an API key to authenticate. It can be supplied either in the provider
configuration block or via environment variables.

```terraform
provider "sendgrid" {
  api_key = var.sendgrid_api_key
}
```

Or, leave `api_key` unset and export `SENDGRID_API_KEY`:

```sh
export SENDGRID_API_KEY="SG.xxxxxxxx.yyyyyyyy"
```

## Provider Configuration Reference

| Attribute      | Environment variable      | Description                                                                                                     |
| -------------- | -------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| `api_key`      | `SENDGRID_API_KEY`         | Required. The SendGrid API key used to authenticate all requests.                                                |
| `host`         | `SENDGRID_API_HOST`        | Optional. Defaults to `https://api.sendgrid.com`. Set to `https://api.eu.sendgrid.com` for EU regional accounts.  |
| `on_behalf_of` | `SENDGRID_ON_BEHALF_OF`    | Optional. Sets the `on-behalf-of` header on every request, acting as a subuser or parent account.                 |

## EU Data Residency

Accounts provisioned in SendGrid's EU region must set `host` (or `SENDGRID_API_HOST`) to
`https://api.eu.sendgrid.com`, or requests will be rejected.

```terraform
provider "sendgrid" {
  api_key = var.sendgrid_api_key
  host    = "https://api.eu.sendgrid.com"
}
```

## Acting On Behalf Of a Subuser or Parent Account

Set `on_behalf_of` to make every request from a parent account act as one of its subusers, or as a specific
customer account:

```terraform
provider "sendgrid" {
  api_key      = var.sendgrid_api_key
  on_behalf_of = "my_subuser_username"
}
```

See SendGrid's [On Behalf Of documentation](https://docs.sendgrid.com/api-reference/how-to-use-the-sendgrid-v3-api/on-behalf-of)
for the full header format, including acting on behalf of a customer account by ID.

## A Note on API Key Scopes

SendGrid silently adds `2fa_required`, `sender_verification_eligible`, and/or `sender_verification_legacy` to
some API keys' scopes regardless of what was requested when creating them via the API. The
[`sendgrid_api_key`](../resources/api_key.md) resource accounts for this and won't show a permanent diff for
these scopes unless you explicitly request them.

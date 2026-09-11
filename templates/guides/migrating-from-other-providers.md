---
page_title: "Migrating from other SendGrid providers"
subcategory: ""
description: |-
  Notes for users coming from Trois-Six/sendgrid or kenzo0107/sendgrid.
---

# Migrating from other SendGrid providers

Several community-maintained Terraform providers for SendGrid exist. This guide covers what's different if
you're migrating an existing configuration.

## From `Trois-Six/sendgrid`

That provider has been archived since May 2022 and is built on the legacy Terraform Plugin SDKv2. Resource
names generally match this provider's (`sendgrid_api_key`, `sendgrid_subuser`, `sendgrid_template`,
`sendgrid_template_version`, `sendgrid_event_webhook`), but `sendgrid_domain_authentication` here replaces
`sendgrid_domain_authentication` there under the same name — no rename needed for that one. Because both
providers use the same resource type names, changing only the `source` in your `required_providers` block and
re-running `terraform init` is usually enough; run `terraform plan` afterward and review any diff carefully, as
attribute-level schema details (defaults, computed fields) differ between the two implementations.

## From `kenzo0107/sendgrid`

That provider covers a broader set of resources than this one currently does (e.g. `sendgrid_design`,
`sendgrid_sso_teammate`). If you depend on a resource this provider doesn't yet implement, keep using
`kenzo0107/sendgrid` for those resources — Terraform supports multiple providers of the same type with
different local names in one configuration:

```terraform
terraform {
  required_providers {
    sendgrid        = { source = "twilio-labs/sendgrid" }
    sendgrid_kenzo  = { source = "kenzo0107/sendgrid" }
  }
}
```

For resources both providers implement, changing `source` and re-running `terraform init` is the same process
as above.

## General migration steps

1. Update the `source` field in your `required_providers` block.
2. Run `terraform init -upgrade`.
3. Run `terraform plan` and review the diff. Schema differences (attribute names, defaults, computed vs.
   optional) between providers can produce a plan even when nothing needs to change remotely — read each
   proposed change before applying.
4. If a resource shows an unwanted destroy/recreate, use
   [`terraform state mv`](https://developer.hashicorp.com/terraform/cli/commands/state/mv) to move it into the
   new provider's state entry rather than letting Terraform recreate it, provided the underlying SendGrid
   object ID stays the same.

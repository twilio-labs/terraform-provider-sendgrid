# Contributing

## Development requirements

- [Go](https://golang.org/doc/install) >= 1.26 (see `go.mod`)
- [Terraform](https://www.terraform.io/downloads.html) >= 1.9
- [golangci-lint](https://golangci-lint.run/welcome/install/) (matches `.golangci.yml`)

## Adding or changing a resource

1. Confirm the exact request/response shape against the SendGrid API before writing code — the OpenAPI spec docs (`https://www.twilio.com/docs/sendgrid/api-reference`) sometimes lag the live API, so prefer testing against a real account when in doubt.
2. Add or update the client method(s) in `internal/sendgrid/`.
3. Add or update the resource/data source in `internal/provider/`, following the existing CRUD/ImportState pattern (see `internal/provider/api_key_resource.go` for a straightforward example).
4. Watch for the two bug classes that have bitten this provider before:
   - An `Optional`-only attribute that gets unconditionally overwritten with an API-derived value (including an empty string/list for an unset field) will crash `terraform apply` with "Provider produced inconsistent result after apply." If a value can come back from the API even when unset in config, the attribute needs `Computed: true` (usually paired with a `UseStateForUnknown` plan modifier).
   - If a resource's `Create`/`Update` makes more than one API call, persist state after each call succeeds — not just at the end — so a failure partway through doesn't orphan a remote resource or commit an unapplied change to state. See `internal/provider/subuser_resource.go`'s `Update` for the pattern.
5. Add a matching data source unless the resource has no reasonable "look up an existing one" use case.
6. Add unit tests for any non-trivial pure logic (request/response shape, model conversion) in the same style as `internal/sendgrid/api_key_test.go` — these run in CI on every PR. Add acceptance tests (`*_test.go` under `internal/provider/`, gated by `resource.Test`) for the resource's full lifecycle — these only run against a live account (see Testing below).
7. Add `examples/resources/sendgrid_<name>/resource.tf` (and `import.sh` if the resource is importable) and a `templates/resources/<name>.md.tmpl` with an appropriate `subcategory`, then run `go generate ./...` to produce `docs/`.
8. Update the resource list in `README.md`.

## Testing

```sh
go test ./...                                                          # unit tests, no credentials needed
TF_ACC=1 SENDGRID_API_KEY=SG.xxxx go test -v ./internal/provider/...   # acceptance tests, real account required
```

Acceptance tests create and destroy real resources against whatever account `SENDGRID_API_KEY` points to. Use a non-production account.

## Before opening a PR

```sh
go build ./... && go vet ./... && golangci-lint run && go test ./... && go generate ./...
```

If `go generate ./...` changes anything under `docs/`, commit that too — CI fails the build if generated docs don't match what's committed.

## Submitting and reviewing changes

1. Open a pull request against `main`.
2. CI (build, vet, lint, unit tests, docs-drift check) must pass.
3. At least one maintainer review is required before merge.
4. Acceptance tests only run automatically on push to `main`, not on pull requests — forked and
   external PRs don't receive the `SENDGRID_API_KEY` secret, so untrusted branches can't spend a
   live account's quota. As part of review, a maintainer runs the acceptance suite locally (or
   merges to trigger the push-triggered run) for any change that touches provider behavior.

## Reporting issues

Open a GitHub issue with as much detail as you can: the resource/data source involved, your
Terraform configuration (redacted of secrets), and the actual vs. expected behavior.

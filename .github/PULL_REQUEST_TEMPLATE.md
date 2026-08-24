## Summary

<!-- What does this change do, and why? -->

## Test plan

- [ ] `go build ./...`, `go vet ./...`, `golangci-lint run` all clean
- [ ] `go test ./...` passes
- [ ] `go generate ./...` produces no uncommitted diff (docs are in sync with schema)
- [ ] Acceptance tests (`TF_ACC=1 go test ./internal/provider/...`) pass against a real SendGrid account, if this PR touches resource/data source behavior
- [ ] README / docs updated if this PR adds, removes, or changes a resource or data source

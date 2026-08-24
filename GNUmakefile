default: help

.PHONY: build
build: ## Build the provider binary
	go build -v .

.PHONY: install
install: ## Build and install the provider binary to $GOBIN (for use with dev_overrides)
	go install -v .

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: test
test: ## Run unit tests (no credentials required)
	go test -v -cover ./...

.PHONY: testacc
testacc: ## Run acceptance tests against a live SendGrid account (requires TF_ACC=1 and SENDGRID_API_KEY)
	TF_ACC=1 go test -v -cover ./internal/provider/... -timeout 30m

.PHONY: docs
docs: ## Regenerate docs/ from schema + examples/ + templates/
	go generate ./...

.PHONY: fmt
fmt: ## Format Go source
	gofmt -l -w .

.PHONY: check
check: build vet lint test docs ## Run everything CI runs, before opening a PR
	@git diff --exit-code -- docs/ || (echo "docs/ is out of sync — 'go generate ./...' produced changes; review and commit them." && exit 1)

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}'

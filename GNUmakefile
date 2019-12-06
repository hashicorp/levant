default: check test build

.PHONY: tools
tools: ## Install the tools used to test and build
	@echo "==> Installing build tools"
	GO111MODULE=off go get -u github.com/ahmetb/govvv
	GO111MODULE=off go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: build
build: ## Build Levant for development purposes
	@echo "==> Running $@..."
	govvv build -o levant-local . -version local

.PHONY: test
test: ## Run the Levant test suite with coverage
	@echo "==> Running $@..."
	@go test -cover -v -tags -race \
		"$(BUILDTAGS)" $(shell go list ./... |grep -v vendor |grep -v test)

.PHONY: acceptance-test
acceptance-test: ## Run the Levant acceptance tests
	@echo "==> Running $@..."
	go test -timeout 120s github.com/jrasell/levant/test -v

.PHONY: release
release: ## Trigger the release build script
	@echo "==> Running $@..."
	@goreleaser --rm-dist

.PHONY: check
check: ## Run golangci-lint
	@echo "==> Running $@..."
	golangci-lint run buildtime/... client/... command/... helper/... levant/... logging/... scale/... template/... version/...

HELP_FORMAT="    \033[36m%-25s\033[0m %s\n"
.PHONY: help
help: ## Display this usage information
	@echo "Levant make commands:"
	@grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; \
			{printf $(HELP_FORMAT), $$1, $$2}'

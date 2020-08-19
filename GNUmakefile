SHELL = bash
default: lint test build check-mod

GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_DIRTY := $(if $(shell git status --porcelain),+CHANGES)

GO_LDFLAGS := "-X github.com/hashicorp/levant/version.GitCommit=$(GIT_COMMIT)$(GIT_DIRTY)"

.PHONY: tools
tools: ## Install the tools used to test and build
	@echo "==> Installing tools..."
	GO111MODULE=off go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	@echo "==> Done"

.PHONY: build
build:
	@echo "==> Building Levant..."
	@CGO_ENABLED=0 GO111MODULE=on \
	go build \
	-ldflags $(GO_LDFLAGS) \
	-o ./bin/levant
	@echo "==> Done"

.PHONY: test
test: ## Test the source code
	@echo "==> Testing source code..."
	@go test -cover -v -tags -race \
		"$(BUILDTAGS)" $(shell go list ./... |grep -v vendor |grep -v test)

.PHONY: acceptance-test
acceptance-test: ## Run the Levant acceptance tests
	@echo "==> Running $@..."
	go test -timeout 120s github.com/hashicorp/levant/test -v

.PHONY: lint
lint: ## Lint the source code
	@echo "==> Linting source code..."
	@golangci-lint run -j 1
	@echo "==> Done"

.PHONEY: check-mod
check-mod: ## Checks the Go mod is tidy
	@echo "==> Checking Go mod..."
	@GO111MODULE=on go mod tidy
	@if (git status --porcelain | grep -q go.mod); then \
		echo go.mod needs updating; \
		git --no-pager diff go.mod; \
		exit 1; fi
	@echo "==> Done"

HELP_FORMAT="    \033[36m%-25s\033[0m %s\n"
.PHONY: help
help: ## Display this usage information
	@echo "Levant make commands:"
	@grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; \
			{printf $(HELP_FORMAT), $$1, $$2}'

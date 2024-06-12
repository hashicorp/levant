SHELL = bash
default: lint test check-mod dev

GIT_COMMIT := $(shell git rev-parse --short HEAD)
GIT_DIRTY := $(if $(shell git status --porcelain),+CHANGES)

GO_LDFLAGS := "$(GO_LDFLAGS) -X github.com/hashicorp/levant/version.GitCommit=$(GIT_COMMIT)$(GIT_DIRTY)"

.PHONY: tools
tools: ## Install the tools used to test and build
	@echo "==> Installing tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.2
	go install github.com/hashicorp/hcl/v2/cmd/hclfmt@d0c4fa8b0bbc2e4eeccd1ed2a32c2089ed8c5cf1
	@echo "==> Done"


pkg/%/levant: GO_OUT ?= $@
pkg/windows_%/levant: GO_OUT = $@.exe
pkg/%/levant: ## Build Levant for GOOS_GOARCH, e.g. pkg/linux_amd64/levant
	@echo "==> Building $@ with tags $(GO_TAGS)..."
	@CGO_ENABLED=0 \
		GOOS=$(firstword $(subst _, ,$*)) \
		GOARCH=$(lastword $(subst _, ,$*)) \
		go build -trimpath -ldflags $(GO_LDFLAGS) -tags "$(GO_TAGS)" -o "$(GO_OUT)"

.PRECIOUS: pkg/%/levant
pkg/%.zip: pkg/%/levant ## Build and zip Levant for GOOS_GOARCH, e.g. pkg/linux_amd64.zip
	@echo "==> Packaging for $@..."
	zip -j $@ $(dir $<)*

.PHONY: crt
crt:
	@CGO_ENABLED=0 go build -trimpath -ldflags $(GO_LDFLAGS) -tags "$(GO_TAGS)" -o "$(BIN_PATH)"


.PHONY: dev
dev: check ## Build for the current development version
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
	go test -timeout 300s github.com/hashicorp/levant/test -v

.PHONY: check
check: tools lint check-mod ## Lint the source code and check other properties

.PHONY: lint
lint: hclfmt ## Lint the source code
	@echo "==> Linting source code..."
	@golangci-lint run -j 1
	@echo "==> Done"

.PHONY: hclfmt
hclfmt: ## Format HCL files with hclfmt
	@echo "--> Formatting HCL"
	@find . -name '.git' -prune \
					-o -name '*fixtures*' -prune \
	        -o \( -name '*.nomad' -o -name '*.hcl' -o -name '*.tf' \) \
	      -print0 | xargs -0 hclfmt -w
	@if (git status -s | grep -q -e '\.hcl$$' -e '\.nomad$$' -e '\.tf$$'); then echo The following HCL files are out of sync; git status -s | grep -e '\.hcl$$' -e '\.nomad$$' -e '\.tf$$'; exit 1; fi

.PHONY: check-mod
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

.PHONY: version
version:
ifneq (,$(wildcard version/version_ent.go))
	@$(CURDIR)/scripts/version.sh version/version.go version/version_ent.go
else
	@$(CURDIR)/scripts/version.sh version/version.go version/version.go
endif

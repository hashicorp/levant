default: check test build

.PHONY: tools
tools: ## Install the tools used to test and build
	@echo "==> Installing build tools"
	go get github.com/ahmetb/govvv
	go get github.com/alecthomas/gometalinter
	gometalinter --install

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
check: ## Run the gometalinter suite
	@echo "==> Running $@..."
	gometalinter \
			--deadline 10m \
			--vendor \
			--sort="path" \
			--aggregate \
			--disable-all \
			--enable golint \
			--enable-gc \
			--enable goimports \
			--enable misspell \
			--enable vet \
			--enable deadcode \
			--enable varcheck \
			--enable ineffassign \
			--enable structcheck \
			--enable errcheck \
			--enable gofmt \
			./...

HELP_FORMAT="    \033[36m%-25s\033[0m %s\n"
.PHONY: help
help: ## Display this usage information
	@echo "Levant make commands:"
	@grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; \
			{printf $(HELP_FORMAT), $$1, $$2}'

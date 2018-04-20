default: check test build

tools: ## Install the tools used to test and build
	@echo "==> Installing build tools"
	go get github.com/mitchellh/gox
	go get github.com/alecthomas/gometalinter
	gometalinter --install

build: ## Build Levant for development purposes
	@echo "==> Running $@..."
	@go build -tags "$(BUILDTAGS) cgo" -o levant-local .

test: ## Run the Levant test suite with coverage
	@echo "==> Running $@..."
	@go test -cover -v -tags \
		"$(BUILDTAGS) cgo" $(shell go list ./... | grep -v vendor)

release: ## Trigger the release build script
	@echo "==> Running $@..."
	./scripts/build.sh

.PHONY: check
check: ## Run the gometalinter suite
	@echo "==> Running $@..."
	gometalinter \
			--deadline 10m \
			--vendor \
			--sort="path" \
			--aggregate \
			--disable-all \
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

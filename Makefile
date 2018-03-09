# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)
SRCFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")
BUILDTAGS=

.PHONY: clean all fmt vet lint spelling build test install static
.DEFAULT: default

all: clean build fmt lint spelling test vet install

build:
	@echo "==> Running $@..."
	@go build -tags "$(BUILDTAGS) cgo" -o levant-local .

static:
	@echo "==> Running $@..."
	CGO_ENABLED=1 go build -tags "$(BUILDTAGS) cgo static_build" -ldflags "-w -extldflags -static" -o ghb0t .

fmt:
	@echo "==> Running $@..."
	@if [ -n "$$(gofmt -s -l $(SRCFILES) )" ]; then \
		echo 'Please run go fmt on your code before submitting the code for reviewal.' \
		&& exit 1; fi

lint:
	@echo "==> Running $@..."
	@golint ./... | grep -v vendor | tee /dev/stderr

test: fmt lint vet spelling
	@echo "==> Running $@..."
	@go test -cover -v -tags "$(BUILDTAGS) cgo" $(shell go list ./... | grep -v vendor)

vet:
	@echo "==> Running $@..."
	@go list ./... \
	| grep -v /vendor/ \
	| cut -d '/' -f 4- \
	| xargs -n1 \
	go tool vet ;\
	if [ $$? -ne 0 ]; then \
	echo ""; \
	echo "Vet found suspicious constructs. Please check the reported constructs"; \
	echo "and fix them if necessary before submitting the code for reviewal."; \
	fi

clean:
	@echo "==> Running $@..."
	@rm -rf ghb0t

install:
	@echo "==> Running $@..."
	@go install .

release:
	@echo "==> Running $@..."
	./scripts/build.sh

spelling:
	@echo "==> Running $@..."
	@echo $(SRCFILES) |xargs misspell -error ;\
	if [ $$? -ne 0 ]; then \
	echo ""; \
	echo "Misspell found spelling mistakes please fix them before submitting the"; \
	echo "code for reviewal."; \
	exit 1; fi
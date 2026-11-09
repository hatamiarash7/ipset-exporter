GO ?= go
GOFMT ?= gofmt
GOLANGCI_LINT ?= golangci-lint
DOCKER ?= docker
GIT ?= git
AWK ?= awk
SUDO ?= sudo
SETCAP ?= setcap
BIN_DIR ?= bin
CONFIG_FILE ?= config.yml
VERSION ?= $(shell $(GIT) describe --tags --always --dirty 2>/dev/null || echo dev)
REVISION ?= $(shell $(GIT) rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
IMAGE ?= hatamiarash7/ipset-exporter
IMAGE_TAG ?= local
BUILDINFO := github.com/hatamiarash7/ipset-exporter/internal/buildinfo
LDFLAGS ?= -s -w -X $(BUILDINFO).Version=$(VERSION) -X $(BUILDINFO).Revision=$(REVISION) -X $(BUILDINFO).BuildDate=$(BUILD_DATE)

.DEFAULT_GOAL := help

##@ General

.PHONY: help
help: ## Show this help
	@$(AWK) 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} \
		/^[a-zA-Z0-9_-]+:.*##/ { printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)

##@ Development

.PHONY: all
all: build ## Build the exporter (compatibility alias)

.PHONY: build
build: ## Build a static Linux exporter with version metadata
	CGO_ENABLED=0 GOOS=linux $(GO) build -trimpath -ldflags '$(LDFLAGS)' -o "$(BIN_DIR)/ipset-exporter" ./cmd/ipset-exporter

.PHONY: run
run: build ## Build, grant NET_ADMIN, and run with CONFIG_FILE
	$(SUDO) $(SETCAP) cap_net_admin+ep "$(BIN_DIR)/ipset-exporter"
	"./$(BIN_DIR)/ipset-exporter" -config "$(CONFIG_FILE)"

.PHONY: update-deps
update-deps: ## Update direct and indirect dependencies and tidy modules
	$(GO) get -u ./...
	$(GO) mod tidy

##@ Quality

.PHONY: format
format: ## Format Go source files
	$(GOFMT) -w cmd internal pkg

.PHONY: format-check
format-check: ## Check formatting without modifying files
	@files="$$($(GOFMT) -l cmd internal pkg)"; result=$$?; \
		if [ $$result -ne 0 ]; then exit $$result; fi; \
		if [ -n "$$files" ]; then printf '%s\n' "$$files"; exit 1; fi

.PHONY: vet
vet: ## Run Go static analysis
	$(GO) vet ./...

.PHONY: lint
lint: ## Run the installed golangci-lint
	$(GOLANGCI_LINT) run --timeout 3m

.PHONY: test
test: ## Run the test suite without cached results
	$(GO) test -count=1 ./...

.PHONY: test-race
test-race: ## Run tests with the race detector
	CGO_ENABLED=1 $(GO) test -race -count=1 ./...

.PHONY: check
check: format-check vet test ## Check formatting, static analysis, and tests

##@ Docker

.PHONY: docker-build
docker-build: ## Build the container image with version metadata
	$(DOCKER) build --build-arg APP_VERSION="$(VERSION)" \
		--build-arg VCS_REF="$(REVISION)" --build-arg DATE_CREATED="$(BUILD_DATE)" \
		-t "$(IMAGE):$(IMAGE_TAG)" .

##@ Cleanup

.PHONY: clean
clean: ## Remove generated binaries from bin/
	rm -rf ./bin

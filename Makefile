GO=go
BIN_DIR=bin
SOURCES=$(shell find . -name '*.go' -not -name '*_test.go' -not -name "main.go")
LDFLAGS := "-s -w"

all: $(BIN_DIR)/ipset-exporter

format:
	find . -name '*.go' -not -path "./.cache/*" | xargs -n1 $(GO) fmt

check: format
	git diff
	git diff-index --quiet HEAD

lint:
	golangci-lint run --timeout 1m0s

clean:
	rm -rf $(BIN_DIR)

test:
	$(GO) test -v ./...

$(BIN_DIR)/%: cmd/% $(SOURCES)
	$(GO) build -ldflags $(LDFLAGS) -o $@ $</*.go

$(BIN_DIR):
	mkdir -p $@

update-deps:
	$(GO) get -u all
	$(GO) mod tidy

run:
	sudo setcap cap_net_admin+ep $@
	$(BIN_DIR)/ipset-exporter

.PHONY: all format check lint clean test update-deps

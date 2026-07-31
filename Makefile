.PHONY: build install uninstall test e2e-human-gate print-version bump-patch bump-minor bump-major release-patch release-minor release-major

PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
BUILD_DIR ?= .build
BUILD_OUTPUT ?= $(BUILD_DIR)/start-issue
# Release tags are the version source.  A checkout between releases carries the
# nearest tag plus its Git describe suffix, which keeps source builds distinct
# from the last published release for update comparisons.
VERSION ?= $(shell git describe --tags --match 'v[0-9]*' --always --dirty 2>/dev/null | sed 's/^v//')
VERSION := $(if $(strip $(VERSION)),$(VERSION),dev)

build:
	@mkdir -p "$(BUILD_DIR)"
	go build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o "$(BUILD_OUTPUT)" ./cmd/start-issue
	@echo "Built: $(BUILD_OUTPUT)"

install: build
	@mkdir -p "$(BINDIR)"
	install -m 0755 "$(BUILD_OUTPUT)" "$(BINDIR)/start-issue"
	@echo "Installed: $(BINDIR)/start-issue"

uninstall:
	rm -f "$(BINDIR)/start-issue"
	@echo "Removed: $(BINDIR)/start-issue"

test:
	@test -z "$$($$(go env GOROOT)/bin/gofmt -l cmd)"
	go vet ./...
	go test ./...
	python3 scripts/check_memory_bank_index.py --max-depth 4
	git diff --check

e2e-human-gate: build
	@START_ISSUE_E2E_BINARY="$(abspath $(BUILD_OUTPUT))" bash test/e2e/human-gate.sh

print-version:
	@echo "$(VERSION)"

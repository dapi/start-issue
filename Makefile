.PHONY: build install uninstall test

PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
BUILD_DIR ?= .build
BUILD_OUTPUT ?= $(BUILD_DIR)/start-issue

build:
	@mkdir -p "$(BUILD_DIR)"
	@bash scripts/build-start-issue "$(BUILD_OUTPUT)" >/dev/null
	@chmod +x "$(BUILD_OUTPUT)"
	@echo "Built: $(BUILD_OUTPUT)"

install:
	@mkdir -p "$(BINDIR)"
	@set -e; \
	tmpfile="$$(mktemp)"; \
	trap 'rm -f "$$tmpfile"' EXIT; \
	bash scripts/build-start-issue "$$tmpfile" >/dev/null; \
	cp "$$tmpfile" "$(BINDIR)/start-issue"
	@chmod +x "$(BINDIR)/start-issue"
	@echo "Installed: $(BINDIR)/start-issue"

uninstall:
	@rm -f "$(BINDIR)/start-issue"
	@echo "Removed: $(BINDIR)/start-issue"

test:
	bash -n scripts/start-issue
	shellcheck scripts/start-issue scripts/build-start-issue scripts/lib/start_issue/*.sh
	git diff --check
	bats test

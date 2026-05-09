.PHONY: build install uninstall

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

.PHONY: install uninstall

PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin

install:
	@mkdir -p "$(BINDIR)"
	@cp scripts/start-issue "$(BINDIR)/start-issue"
	@chmod +x "$(BINDIR)/start-issue"
	@echo "Installed: $(BINDIR)/start-issue"

uninstall:
	@rm -f "$(BINDIR)/start-issue"
	@echo "Removed: $(BINDIR)/start-issue"

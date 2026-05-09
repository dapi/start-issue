.PHONY: install uninstall

PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin
LIBDIR ?= $(PREFIX)/lib/start-issue

install:
	@mkdir -p "$(BINDIR)"
	@mkdir -p "$(LIBDIR)"
	@cp scripts/start-issue "$(BINDIR)/start-issue"
	@cp scripts/lib/start_issue/*.sh "$(LIBDIR)/"
	@chmod +x "$(BINDIR)/start-issue"
	@echo "Installed: $(BINDIR)/start-issue"
	@echo "Installed libs: $(LIBDIR)"

uninstall:
	@rm -f "$(BINDIR)/start-issue"
	@rm -rf "$(LIBDIR)"
	@echo "Removed: $(BINDIR)/start-issue"
	@echo "Removed: $(LIBDIR)"

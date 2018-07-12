include  ./extras/make/paths.mk

GD2 = glusterd2

BUILDDIR = build
BASH_COMPLETIONDIR = /etc/bash_completion.d

GD2_BIN = $(GD2)
GD2_BUILD = $(BUILDDIR)/$(GD2_BIN)
GD2_INSTALL = $(DESTDIR)$(SBINDIR)/$(GD2_BIN)

CLI_BIN = glustercli
CLI_BUILD = $(BUILDDIR)/$(CLI_BIN)
CLI_INSTALL = $(DESTDIR)$(SBINDIR)/$(CLI_BIN)

CLI_BASH_COMPLETION_GEN_BIN = $(BUILDDIR)/generate_bash_completion
CLI_BASH_COMPLETION_BUILD = $(BUILDDIR)/$(CLI_BIN).sh
CLI_BASH_COMPLETION_INSTALL = $(DESTDIR)$(BASH_COMPLETIONDIR)/$(CLI_BIN).sh

GD2_CONF = $(GD2).toml
GD2CONF_BUILDSCRIPT=./scripts/gen-gd2conf.sh
GD2CONF_BUILD = $(BUILDDIR)/$(GD2_CONF)
GD2CONF_INSTALL = $(DESTDIR)$(SYSCONFDIR)/$(GD2)/$(GD2_CONF)

GD2STATEDIR = $(LOCALSTATEDIR)/$(GD2)
GD2LOGDIR = $(LOGDIR)/$(GD2)
GD2RUNDIR = $(RUNDIR)/$(GD2)

DEPENV ?=

PLUGINS ?= yes
FASTBUILD ?= yes

.PHONY: all build binaries check check-go check-reqs install vendor-update vendor-install verify release check-protoc $(GD2_BIN) $(GD2_BUILD) $(CLI_BIN) $(CLI_BUILD) cli $(GD2_CONF) gd2conf test dist dist-vendor functest

all: build

build: check-go check-reqs vendor-install binaries $(GD2_CONF)
check: check-go check-reqs check-protoc
binaries: $(GD2_BIN) $(CLI_BIN)

check-go:
	@./scripts/check-go.sh
	@echo

check-protoc:
	@./scripts/check-protoc.sh
	@echo

check-reqs:
	@./scripts/check-reqs.sh
	@echo

$(GD2_BIN): $(GD2_BUILD) gd2conf
$(GD2_BUILD):
	@PLUGINS=$(PLUGINS) FASTBUILD=$(FASTBUILD) BASE_PREFIX=$(BASE_PREFIX) ./scripts/build.sh glusterd2
	@echo

$(CLI_BIN) cli: $(CLI_BUILD)
$(CLI_BUILD):
	@FASTBUILD=$(FASTBUILD) GD2_STATE_DIR=$(GD2STATEDIR) ./scripts/build.sh  glustercli
	@FASTBUILD=$(FASTBUILD) ./scripts/build.sh glustercli/generate_bash_completion
	@./$(CLI_BASH_COMPLETION_GEN_BIN) $(CLI_BASH_COMPLETION_BUILD)
	@echo

$(GD2_CONF) gd2conf:
	@GD2=$(GD2) GD2STATEDIR=$(GD2STATEDIR) GD2LOGDIR=$(GD2LOGDIR) \
		GD2RUNDIR=$(GD2RUNDIR) $(GD2CONF_BUILDSCRIPT)

install:
	install -D $(GD2_BUILD) $(GD2_INSTALL)
	install -D $(CLI_BUILD) $(CLI_INSTALL)
	install -D -m 0644 $(GD2CONF_BUILD) $(GD2CONF_INSTALL)
	install -D -m 0644 $(CLI_BASH_COMPLETION_BUILD) $(CLI_BASH_COMPLETION_INSTALL)
	@echo

vendor-update:
	@echo Updating vendored packages
	@$(DEPENV) dep ensure -update
	@echo

vendor-install:
	@echo Installing vendored packages
	@$(DEPENV) dep ensure
	@echo

test: check-reqs
	@./test.sh $(TESTOPTIONS)

functest: check-reqs
	@go test ./e2e -v -functest

release: build
	@./scripts/release.sh

dist:
	@DISTDIR=$(DISTDIR) SIGN=$(SIGN) ./scripts/dist.sh

dist-vendor: vendor-install
	@VENDOR=yes DISTDIR=$(DISTDIR) SIGN=$(SIGN) ./scripts/dist.sh

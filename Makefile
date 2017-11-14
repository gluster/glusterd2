include  ./extras/make/paths.mk

GD2 = glusterd2

BUILDDIR = build

GD2_BIN = $(GD2)
GD2_BUILD = $(BUILDDIR)/$(GD2_BIN)
GD2_INSTALL = $(SBINDIR)/$(GD2_BIN)

CLI_BIN = glustercli
CLI_BUILD = $(BUILDDIR)/$(CLI_BIN)
CLI_INSTALL = $(SBINDIR)/$(CLI_BIN)

GD2_CONF = $(GD2).toml
GD2CONF_BUILDSCRIPT=./scripts/gen-gd2conf.sh
GD2CONF_BUILD = $(BUILDDIR)/$(GD2_CONF)
GD2CONF_INSTALL = $(SYSCONFDIR)/$(GD2)/$(GD2_CONF)

TEMPLATESDIR = bin/glusterd2/volgen/templates
TEMPLATES_INSTALL = $(DATADIR)/$(GD2)/templates

GD2STATEDIR = $(LOCALSTATEDIR)/$(GD2)
GD2LOGDIR = $(LOGDIR)/$(GD2)
GD2TEMPLATESDIR = $(TEMPLATES_INSTALL)

PLUGINS ?= yes

.PHONY: all build check check-go check-reqs install vendor-update vendor-install verify release check-protoc $(GD2_BIN) $(GD2_BUILD) $(CLI_BIN) $(CLI_BUILD) cli $(GD2_CONF) gd2conf test dist dist-vendor

all: build

build: check-go check-reqs vendor-install glusterd2 glustercli glusterd2.toml
check: check-go check-reqs check-protoc

check-go:
	@./scripts/check-go.sh
	@echo

check-protoc:
	@./scripts/check-protoc.sh
	@echo

check-reqs:
	@./scripts/check-reqs.sh
	@echo

$(GD2_BIN): $(GD2_BUILD)
$(GD2_BUILD):
	@PLUGINS=$(PLUGINS) ./scripts/build.sh glusterd2
	@echo

$(CLI_BIN) cli: $(CLI_BUILD)
$(CLI_BUILD):
	@./scripts/build.sh glustercli
	@echo

$(GD2_CONF) gd2conf: $(GD2CONF_BUILD)
$(GD2CONF_BUILD):
	@GD2STATEDIR=$(GD2STATEDIR) GD2LOGDIR=$(GD2LOGDIR) GD2TEMPLATESDIR=$(GD2TEMPLATESDIR) $(GD2CONF_BUILDSCRIPT)

install:
	install -D $(GD2_BUILD) $(GD2_INSTALL)
	install -D $(CLI_BUILD) $(CLI_INSTALL)
	install -D $(GD2CONF_BUILD) $(GD2CONF_INSTALL)
	install -D -t $(TEMPLATES_INSTALL) $(TEMPLATESDIR)/*.graph
	@echo

vendor-update:
	@echo Updating vendored packages
	@dep ensure -update
	@echo

vendor-install:
	@echo Installing vendored packages
	@dep ensure
	@echo

verify: check-reqs
	@./scripts/lint-check.sh
	@gometalinter -D gotype -E gofmt --errors --deadline=5m -j 4 --vendor

test:
	@go test $$(go list ./... | sed '/e2e/d;/vendor/d')
	@go test ./e2e -v -functest

release: build
	@./scripts/release.sh

dist:
	@./scripts/dist.sh

dist-vendor: vendor-install
	@VENDOR=yes ./scripts/dist.sh

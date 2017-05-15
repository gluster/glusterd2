GOPATH := $(shell go env GOPATH)
GOBIN := '$(GOPATH)/bin'
PLUGINS ?= yes

.PHONY: all build check check-go check-reqs install vendor-update vendor-install verify glusterd2 release check-protoc

all: build

build: check-go check-reqs vendor-install glusterd2

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

glusterd2:
	@PLUGINS=$(PLUGINS) ./scripts/build.sh
	@echo

install: check-go check-reqs vendor-install
	@PLUGINS=$(PLUGINS) ./scripts/build.sh $(GOBIN)
	@echo Setting CAP_SYS_ADMIN for glusterd2 \(requires sudo\)
	sudo setcap cap_sys_admin+ep $(GOBIN)/glusterd2
	@echo

vendor-update:
	@echo Updating vendored packages
	@glide update --strip-vendor
	@echo

vendor-install:
	@echo Installing vendored packages
	@glide install --strip-vendor
	@echo

verify: check-reqs
	@./scripts/lint-check.sh
	@gometalinter -D gotype -E gofmt --errors --deadline=5m -j 4 $$(glide nv)

test:
	@go test -tags 'novirt noaugeas' $$(glide nv)

release: check-go check-reqs vendor-install
	@./scripts/release.sh

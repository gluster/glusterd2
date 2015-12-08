.PHONY: all build check check-go check-reqs install vendor-update verify

all: build

build: check vendor-update glusterd2

check: check-go check-reqs

check-go:
	@./scripts/check-go.sh
	@echo

check-reqs:
	@./scripts/check-reqs.sh
	@echo

glusterd2:
	@echo Building GlusterD-2.0
	@GO15VENDOREXPERIMENT=1 go build
	@echo

install:
	@echo Building and installing GlusterD-2.0
	@GO15VENDOREXPERIMENT=1 go install
	@echo

vendor-update:
	@echo Updating vendored packages
	@GO15VENDOREXPERIMENT=1 glide -q up
	@echo

verify: check-reqs
	@GO15VENDOREXPERIMENT=1 gometalinter -D gotype -E gofmt --errors --deadline=1m $$(GO15VENDOREXPERIMENT=1 glide nv)

test:
	@GO15VENDOREXPERIMENT=1 go test $$(GO15VENDOREXPERIMENT=1 glide nv)

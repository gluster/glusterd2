VERSION := $(shell bash ./scripts/pkg-version --full)
LDFLAGS := '-X github.com/gluster/glusterd2/gdctx.GlusterdVersion=$(VERSION)'

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
	@GO15VENDOREXPERIMENT=1 go build -ldflags $(LDFLAGS)
	@echo

install: check vendor-update
	@echo Building and installing GlusterD-2.0
	@GO15VENDOREXPERIMENT=1 go install -ldflags $(LDFLAGS)
	@echo Setting CAP_SYS_ADMIN for glusterd2 \(requires sudo\)
	sudo setcap cap_sys_admin+ep $$GOPATH/bin/glusterd2
	@echo

vendor-update:
	@echo Updating vendored packages
	@GO15VENDOREXPERIMENT=1 glide install
	@echo

verify: check-reqs
	@GO15VENDOREXPERIMENT=1 gometalinter -D gotype -E gofmt --errors --deadline=5m -j 4 $$(GO15VENDOREXPERIMENT=1 glide nv)

test:
	@GO15VENDOREXPERIMENT=1 go test $$(GO15VENDOREXPERIMENT=1 glide nv)

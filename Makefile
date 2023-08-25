PULUMI_TEST_ORG   ?= $(shell pulumi whoami)
PULUMI_TEST_OWNER ?= ${PULUMI_TEST_ORG}
PULUMI_LIVE_TEST  ?= false
export PULUMI_TEST_ORG
export PULUMI_TEST_OWNER

CONCURRENCY       := 10
SHELL := sh

GO                          := go

.phony: .EXPORT_ALL_VARIABLES
.EXPORT_ALL_VARIABLES:

default: ensure build

install::
	${GO} install ./cmd/...

clean::
	rm -f ./bin/*

ensure::
	${GO} mod download

.phony: lint
lint:: lint-copyright lint-golang
lint-golang:
	golangci-lint run
lint-copyright:
	pulumictl copyright

build:: ensure
	${GO} build -p ${CONCURRENCY} ./...

test:: build
	${GO} test --timeout 30m -short -count 1 -parallel ${CONCURRENCY} ./...

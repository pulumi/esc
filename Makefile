VERSION := $(if ${PULUMI_VERSION},${PULUMI_VERSION},$(shell ./scripts/pulumi-version.sh))

CONCURRENCY := 10
SHELL := sh

GO := go

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

.phony: format
format:
	find . -iname "*.go" -print0 | xargs -r0 gofmt -s -w

build:: ensure
	${GO} install -ldflags "-X github.com/pulumi/esc/cmd/internal/version.Version=${VERSION}" ./cmd/esc

build_debug:: ensure
	${GO} install -gcflags="all=-N -l" -ldflags "-X github.com/pulumi/esc/cmd/internal/version.Version=${VERSION}" ./cmd/esc

test:: build
	${GO} test --timeout 30m -short -count 1 -parallel ${CONCURRENCY} ./...

test_cover:: build
	${GO} test --timeout 30m -count 1 -coverpkg=github.com/pulumi/esc/... -race -coverprofile=coverage.out -parallel ${CONCURRENCY} ./...

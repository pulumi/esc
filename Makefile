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

.PHONY: generate_go_client_sdk
generate_go_client_sdk:
	GO_POST_PROCESS_FILE="/usr/local/bin/gofmt -w" openapi-generator-cli generate -i ./sdk/swagger.yaml -p packageName=esc_sdk,withGoMod=false,isGoSubmodule=true,userAgent=esc-sdk/go/${VERSION} -t ./sdk/templates/go -g go -o ./sdk/go --git-repo-id esc --git-user-id pulumi

.PHONY: generate_ts_client_sdk
generate_ts_client_sdk:
	TS_POST_PROCESS_FILE="/usr/local/bin/prettier --write" openapi-generator-cli generate -i ./sdk/swagger.yaml -p npmName=@pulumi/esc-sdk,userAgent=esc-sdk/ts/${VERSION} -t ./sdk/templates/typescript --enable-post-process-file -g typescript-axios -o ./sdk/typescript/esc/raw  --git-repo-id esc --git-user-id pulumi
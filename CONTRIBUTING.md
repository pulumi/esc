# Contributing to Pulumi ESC

## Development setup

### Prerequisites
- Go 1.24+ (see `go.mod` for exact version)
- [golangci-lint](https://golangci-lint.run/welcome/install-local/) v1.64+
- [pulumictl](https://github.com/pulumi/pulumictl) (for copyright checks)

### Build and test
```sh
make build        # Build the esc binary
make test         # Run tests (-short, parallel)
make test_cover   # Run full tests with race detection and coverage
make lint         # Run all linters
make format       # Format all Go files
make verify       # Format + lint + test (pre-commit check)
```

### Submitting changes
1. Create a branch from `main`.
2. Make your changes. Keep PRs focused — one concern per PR.
3. Run `make verify` and fix any failures.
4. Add a changelog entry to `CHANGELOG_PENDING.md` if the change is user-facing.
5. Open a PR. Fill in all sections of the PR template.
6. Include evidence of test execution in the PR body.

### Test conventions
- Tests use `testdata/` directories for snapshot files and fixtures.
- If your change alters output, update snapshot files with `PULUMI_ACCEPT=true make test` and carefully review the diff.
- Use `go test -run TestName ./path/to/package/` to run a single test.

## Performing a release

We use [goreleaser](https://goreleaser.com/intro/) for automating releases.
To cut a new release, create a commit that:
- Copy the entries in [CHANGELOG_PENDING](./CHANGELOG_PENDING.md) into [CHANGELOG](./CHANGELOG.md).
  CHANGELOG_PENDING is used to generate the release notes. After releasing, the following commit can clear the changes from pending.
- Bumps the version in the [.version](./.version) file, which is used to stamp the version into the binary.
- Tag the commit with a version tag in the format vX.X.X, to trigger the [release automation](./.github/workflows/publish-release.yaml).

## Performing a release

We use [goreleaser](https://goreleaser.com/intro/) for automating releases. 
To cut a new release, create a commit that:
- Moves the entries in [CHANGELOG_PENDING](./CHANGELOG_PENDING.md) into [CHANGELOG](./CHANGELOG.md), (which is used to generate the release notes)
- Bumps the version in the [.version](./.version) file, (which is used to stamp the version into the binary)
- Tag the commit with a version tag in the format vX.X.X, (to trigger the [release automation](./.github/workflows/publish-release.yaml))

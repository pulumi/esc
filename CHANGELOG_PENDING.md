### Improvements

- Coerce non-string scalars passed to `esc env set --secret` to strings
  [#353](https://github.com/pulumi/esc/pull/353)

- `esc env get --show-secrets` now shows secrets from imported environments.
  [#355](https://github.com/pulumi/esc/pull/355)

### Bug Fixes

- Fix a panic in fetching current credentials when the access key had expired.
  [#368](https://github.com/pulumi/esc/pull/368)

### Breaking changes

- The minimum Go version supported is now 1.21.
  [#379](https://github.com/pulumi/esc/pull/379)
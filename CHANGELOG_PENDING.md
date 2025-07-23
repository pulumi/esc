### Improvements

- `esc env set` now supports --file parameter to read content from a file or stdin [#556](https://github.com/pulumi/esc/pull/556)
- `--draft` flag for `esc env set`, `esc env edit`, `esc env versions rollback` to create a change request rather than updating directly. **Warning: this feature is in preview, limited to specific orgs, and subject to change.**
  [#552](https://github.com/pulumi/esc/pull/552)

### Bug Fixes

- Fix decryption error with keys with dashes
  [#559](https://github.com/pulumi/esc/pull/559)

### Breaking changes

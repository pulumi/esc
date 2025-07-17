### Improvements

- `esc run` expects environment to be passed before `--`
  [#545](https://github.com/pulumi/esc/pull/546)
- `esc env set` uses a more readable YAML format when setting a key in an empty map  
  [#548](https://github.com/pulumi/esc/pull/548)
- `--draft` flag for `esc env set`, `esc env edit`, `esc env versions rollback` to create a change request rather than updating directly. **Warning: this feature is in preview, limited to specific orgs, and subject to change.**
  [#552](https://github.com/pulumi/esc/pull/552)

### Bug Fixes

- Fix `esc version`
  [#541](https://github.com/pulumi/esc/pull/541)

### Breaking changes

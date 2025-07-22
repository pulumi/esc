### Improvements

- `esc run` expects environment to be passed before `--`
  [#545](https://github.com/pulumi/esc/pull/546)
- `esc env set` uses a more readable YAML format when setting a key in an empty map  
  [#548](https://github.com/pulumi/esc/pull/548)
- `esc env set` now supports --file parameter to read content from a file or stdin [#556](https://github.com/pulumi/esc/pull/556)

### Bug Fixes

- Fix decryption error with keys with dashes
 [#559](https://github.com/pulumi/esc/pull/559)

### Breaking changes

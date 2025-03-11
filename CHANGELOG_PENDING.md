### Improvements

- Add `--string` flag to `env set` to treat the given value as a string.
  [#467](https://github.com/pulumi/esc/pull/467)
- Add proper return code to list environments when organization doesn't exist
  [#484](https://github.com/pulumi/esc/pull/484)

### Bug Fixes

### Breaking changes

- It is now a syntax error to call a builtin function incorrectly.
  [449](https://github.com/pulumi/esc/pull/449)

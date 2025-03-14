### Improvements

- Add `--string` flag to `env set` to treat the given value as a string.
  [#467](https://github.com/pulumi/esc/pull/467)
- Add proper return code to list environments when organization doesn't exist
  [#484](https://github.com/pulumi/esc/pull/484)

### Bug Fixes

- Adding cascading secrets to `unexport` method
  [#488](https://github.com/pulumi/esc/pull/488)

### Breaking changes

- It is now a syntax error to call a builtin function incorrectly.
  [449](https://github.com/pulumi/esc/pull/449)

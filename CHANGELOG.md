CHANGELOG
=========

## 0.5.4

### Bug Fixes

- Do not panic when `env set` is passed an empty value.
  [#110](https://github.com/pulumi/esc/pull/110)
  
- Fix behavior for `esc login` when no existing credentials are present
  [#111](https://github.com/pulumi/esc/pull/111)

## 0.5.3

### Bug Fixes

- Fix behavior for `esc login` when no backend is provided
  [#105](https://github.com/pulumi/esc/pull/105)

## 0.5.2

### Improvements

- Add a `-f` flag to `esc env init` that allows the specification of the initial definition
  for an environment.
  [#95](https://github.com/pulumi/esc/pull/95)

### Bug Fixes

- Fix panics that could occur when no credentials are available or credentials were invalid.
  [#93](https://github.com/pulumi/esc/pull/93)

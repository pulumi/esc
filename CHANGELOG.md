CHANGELOG
=========

## 0.6.2

### Bug Fixes

- Add support for yaml format as output from `esc open`.
  [#204](https://github.com/pulumi/esc/pull/204)

## 0.6.1

### Bug Fixes

- Fix a nil pointer dereference in Syntax.NodeError.
  [#180](https://github.com/pulumi/esc/pull/180)

- Mark nested structures as secret if the JSON string is secret.
  [#191](https://github.com/pulumi/esc/pull/191)

## 0.6.0

### Improvements

- Include paths in diagnostics.
  [#157](https://github.com/pulumi/esc/pull/157)
  
- Support secret elision in definitions via encryption and decryption
  [#155](https://github.com/pulumi/esc/pull/155)

- Support `--show-secrets` in `esc env get` to display secrets in plaintext.
  [#163](https://github.com/pulumi/esc/pull/163)

## 0.5.7

### Improvements

- Add two new builtins, `fn::fromBase64` and `fn::fromJSON`. The former decodes a base64-encoded
  string into a binary string and the latter decodes a JSON string into a value.
  [#117](https://github.com/pulumi/esc/pull/117)
- Add support for temporary file projection in run and open commands.
  [#141](https://github.com/pulumi/esc/pull/141)
  [#151](https://github.com/pulumi/esc/pull/151)
- Support null, boolean, and number values in environment variables.
  [#151](https://github.com/pulumi/esc/pull/151)

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

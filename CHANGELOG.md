CHANGELOG
=========

## 0.8.3

### Improvements

- Propagate current and root env name to providers.
  [#264](https://github.com/pulumi/esc/pull/264)

### Bug Fixes

- Specify pulumi access token per command run.
  [#263](https://github.com/pulumi/esc/pull/263)

## 0.8.2

### Improvements

- Document the CLI's REST API client.
  [#257](https://github.com/pulumi/esc/pull/257)

## 0.8.1

### Improvements

- Handle anonymous environments when injecting it to the context
  [#250](https://github.com/pulumi/esc/pull/250)

- Export context schema for editor autocompletion
  [#252](https://github.com/pulumi/esc/pull/252)

## 0.8.0

### Improvements

- Allow evaluation of environments with parse errors.
  [#222](https://github.com/pulumi/esc/pull/222)

- Return a properly-merged root schema from environment evaluation.
  [#229](https://github.com/pulumi/esc/pull/229)

- Improve property accessor diagnostics.
  [#230](https://github.com/pulumi/esc/pull/230)

- Populate source positions for property accessors in single-line flow scalars.
  [#231](https://github.com/pulumi/esc/pull/231)

- Provide more accurate accessor diagnostic positions.
  [#238](https://github.com/pulumi/esc/pull/238)

- Add support for execution context interpolation.
  [#239](https://github.com/pulumi/esc/pull/239)
  
## 0.7.0

### Bug Fixes

- Fix merging of already-merged values.
  [#213](https://github.com/pulumi/esc/pull/213)

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

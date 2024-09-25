CHANGELOG
=========

## 0.10.0

### Improvements

- Add commands to manage environment tags.
  [#345](https://github.com/pulumi/esc/pull/345)

- Coerce non-string scalars passed to `esc env set --secret` to strings
  [#353](https://github.com/pulumi/esc/pull/353)

- `esc env get --show-secrets` now shows secrets from imported environments.
  [#355](https://github.com/pulumi/esc/pull/355)

- Add support for projects.
  [#369](https://github.com/pulumi/esc/pull/369)

- Add deprecation warning for legacy environment name format (<org>/<env> or <env>) in favor of <project>/<env> or <org>/<project>/<env>.
  [#375](https://github.com/pulumi/esc/pull/375)

- Add clone environment command.
  [#376](https://github.com/pulumi/esc/pull/376)

- Add project filter flag to env ls command.
  [#382](https://github.com/pulumi/esc/pull/382)


### Bug Fixes

- Fix a panic in fetching current credentials when the access key had expired.
  [#368](https://github.com/pulumi/esc/pull/368)

### Breaking Changes

- The minimum Go version supported is now 1.21.
  [#379](https://github.com/pulumi/esc/pull/379)

## 0.9.1

### Improvements

- Add a command to retract a specific revision of an environment.
  [#330](https://github.com/pulumi/esc/pull/330)

- Move the `rollback` command under the `version` command for consistency.
  [#331](https://github.com/pulumi/esc/pull/331)

### Bug Fixes

- Buffer output to the system pager in paged commands.
  [#327](https://github.com/pulumi/esc/pull/327)

## 0.9.0

### Improvements

- Add support for getting or opening environments at specific revisions/tags.
  [#275](https://github.com/pulumi/esc/pull/275)

- Add support for listing the revisions to an environment.
  [#277](https://github.com/pulumi/esc/pull/277)

- Add support for managing version tags.
  [#283](https://github.com/pulumi/esc/pull/283)

- Add support for displaying changed between environment revisions.
  [#295](https://github.com/pulumi/esc/pull/295)
  
- Finalize command tree for version management.
  [#304](https://github.com/pulumi/esc/pull/304)

- Add support to `esc env edit` for reading the edited environment definition from a file.
  [#308](https://github.com/pulumi/esc/pull/308)

- Add support for rolling back to a specific version of an environment.
  [#305](https://github.com/pulumi/esc/pull/305)

- Add a new ESC SDK for Go and Typescript
  [#271](https://github.com/pulumi/esc/pull/271)

- Add revision field to GetEnvironment and UpdateEnvironment client functions
  [#313](https://github.com/pulumi/esc/issues/313)

### Bug Fixes

- Ensure that redacted output is flushed in `esc run`
  [#280](https://github.com/pulumi/esc/pull/280/files)

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

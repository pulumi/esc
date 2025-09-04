# CHANGELOG

## 0.18.0

### Improvements

- Added support for `fn::concat` built-in function to concatenate arrays.
  [#582](https://github.com/pulumi/esc/pull/582)

## 0.17.0

### Improvements

- Move `--draft` out of preview for `env edit`, `env set`, `env version rollback`, `env run` and `env open`.
  [#566](https://github.com/pulumi/esc/pull/566)
- Internal improvements
  [#569](https://github.com/pulumi/esc/pull/569)

### Bug Fixes

- Fix broken change request URL that links to web console when using `--draft` with `env set`, `env edit` and `env version rollback`.
  [#571](https://github.com/pulumi/esc/pull/571)

## 0.16.0

### Improvements

- Update a draft via `--draft=<change-request-id>` for `env edit`, `env set`, and `env version rollback`. **Warning: this feature is in preview, limited to specific orgs, and subject to change.**
  [#565](https://github.com/pulumi/esc/pull/565)
- Open a draft via `--draft=<change-request-id>` for `env run` and `env open`. **Warning: this feature is in preview, limited to specific orgs, and subject to change.**
  [#565](https://github.com/pulumi/esc/pull/565)

## 0.15.0

### Improvements

- `esc env set` now supports --file parameter to read content from a file or stdin [#556](https://github.com/pulumi/esc/pull/556)
- `--draft` flag for `esc env set`, `esc env edit`, `esc env versions rollback` to create a change request rather than updating directly. **Warning: this feature is in preview, limited to specific orgs, and subject to change.**
  [#552](https://github.com/pulumi/esc/pull/552)

### Bug Fixes

- Fix decryption error with keys with dashes
  [#559](https://github.com/pulumi/esc/pull/559)

## 0.14.3

### Improvements

- `esc run` expects environment to be passed before `--`
  [#545](https://github.com/pulumi/esc/pull/546)
- `esc env set` uses a more readable YAML format when setting a key in an empty map
  [#548](https://github.com/pulumi/esc/pull/548)

### Bug Fixes

- Fix `esc version`
 [#541](https://github.com/pulumi/esc/pull/541)

## 0.13.2

### Bug Fixes

- handle nil in MakeSecret
 [#518](https://github.com/pulumi/esc/pull/518)

## 0.13.1

### Improvements

- Updated to go 1.23

### Bug Fixes

- cmd/esc/cli/env.go: Modified the writeYAMLEnvironmentDiagnostics function to instantiate hcl.NewDiagnosticTextWriter with a width of 0 initially, and then conditionally reinstantiate it with the specified width if it is greater than 0, addressing gosec G115. [#494](https://github.com/pulumi/esc/pull/494)
- No longer error when decrypting invalid secrets outside of values top-level key
  [#491](https://github.com/pulumi/esc/pull/491)
- Make CLI prefer environment variable `PULUMI_BACKEND_URL` over account backend URL
  [#477](https://github.com/pulumi/esc/pull/477)
- Adding cascading secrets into `NewSecret` method
  [#488](https://github.com/pulumi/esc/pull/488)

## 0.13.0

### Improvements

- Add `--string` flag to `env set` to treat the given value as a string.
  [#467](https://github.com/pulumi/esc/pull/467)
- Add proper return code to list environments when organization doesn't exist
  [#484](https://github.com/pulumi/esc/pull/484)

### Breaking changes

- It is now a syntax error to call a builtin function incorrectly.
  [449](https://github.com/pulumi/esc/pull/449)

## 0.12.0

### Improvements

- Fix diagnostic messages when updating environment with invalid definition
  [#422](https://github.com/pulumi/esc/pull/422)
- Introduce support for rotating static credentials via `fn::rotate` providers [432](https://github.com/pulumi/esc/pull/432)
- Add the `rotate` CLI command
  [#433](https://github.com/pulumi/esc/pull/433)
- Add ability to pass specific paths to rotate with the `rotate` CLI command
  [#440](https://github.com/pulumi/esc/pull/440)
- Introduce inline environment reference syntax
  [#443](https://github.com/pulumi/esc/pull/443)
- Introduce rotateOnly inputs
  [#444](https://github.com/pulumi/esc/pull/444)
- Release rotate environment CLI command
  [#459](https://github.com/pulumi/esc/pull/459)

## 0.11.1

### Improvements

- Add `--definition` flag to `esc env get` to output definition
  [#416](https://github.com/pulumi/esc/pull/416)

## 0.11.0

### Improvements

- Improve evaluation performance and memory footprint.
  [#392](https://github.com/pulumi/esc/pull/392)

- Improve login error message when credentials file is missing or invalid.
  [#404](https://github.com/pulumi/esc/pull/404)

### Bug Fixes

- Fix panic when object keys are not strings.
  [#406](https://github.com/pulumi/esc/pull/406)

### Breaking changes

- `schema`: `ObjectBuilder.Properties` and `Record` now take a `MapBuilder` in order to avoid copies.
  [#392](https://github.com/pulumi/esc/pull/392)

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

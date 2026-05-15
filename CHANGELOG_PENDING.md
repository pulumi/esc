### Improvements

- Add `esc env webhook` to manage environment webhooks
  [#647](https://github.com/pulumi/esc/pull/647)

- Add `esc env schedule` to manage environment scheduled actions
  [#646](https://github.com/pulumi/esc/pull/646)

- Add `new` as an alias for `esc env init`
  [#644](https://github.com/pulumi/esc/pull/644)

- Surface warnings when editing environments with the CLI
  [#631](https://github.com/pulumi/esc/pull/631)

- Add `esc env referrer list` (alias `ls`) to list entities that reference an environment
  [#645](https://github.com/pulumi/esc/pull/645)

- Migrate golangci-lint to v2 and enable `staticcheck`; minor reword of
  a handful of CLI error messages as a result
  [#648](https://github.com/pulumi/esc/pull/648)

- Consolidate `SilenceUsage` / `SilenceErrors` on the `esc` root command,
  matching the `pulumi/pulumi` pattern; the CLI now prints errors itself
  rather than relying on cobra
  [#654](https://github.com/pulumi/esc/pull/654)

- Add `esc env provider {aws-login,azure-login,gcp-login} {static,oidc}`
  commands to add login providers to an environment. `static` writes a
  static-credentials block; `oidc` writes a federated-identity block that
  references cloud-side OIDC resources you provision separately (e.g.
  with Pulumi). Each supports `--create` to create the target environment
  if it does not already exist

### Bug Fixes

### Breaking changes

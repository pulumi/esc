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

- Add `--output {text,json}` to the read-mostly `esc env` commands
  (`env ls`, `env version history`, `env version tag ls`,
  `env tag ls/get`, `env schedule list/get/history`, `env referrer list`,
  `env webhook list/get`, `env webhook delivery list`, `env open-request`,
  `env settings get`) for piping into `jq` / scripts; the default `text`
  rendering is unchanged
  [#656](https://github.com/pulumi/esc/pull/656)

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

- Render `esc env schedule list`, `esc env schedule history`,
  `esc env referrer list`, and `esc env version tag ls` as tables for
  consistency with the other `esc env ... list` commands and with
  `pulumi/pulumi`
  [#655](https://github.com/pulumi/esc/pull/655)

### Bug Fixes

### Breaking changes

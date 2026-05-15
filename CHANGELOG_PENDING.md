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

- Render `esc env schedule list`, `esc env schedule history`,
  `esc env referrer list`, and `esc env version tag ls` as tables for
  consistency with the other `esc env ... list` commands and with
  `pulumi/pulumi`
  [#655](https://github.com/pulumi/esc/pull/655)

### Bug Fixes

### Breaking changes

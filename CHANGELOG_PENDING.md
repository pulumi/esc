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

### Bug Fixes

### Breaking changes

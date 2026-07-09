### Improvements

- Add support for `fn::toYAML` built-in function that encodes a value as YAML.
  [#642](https://github.com/pulumi/esc/pull/642)

- Report the new revision number when an environment definition is updated (e.g. `esc env edit`,
  `esc env provider`, `esc env version rollback`).
  [#666](https://github.com/pulumi/esc/pull/666)

- Print a retirement notice on every `esc` command. The standalone ESC CLI has been retired;
  install the Pulumi CLI and use the `pulumi env` subcommand instead. See
  https://www.pulumi.com/docs/install.
  [#669](https://github.com/pulumi/esc/pull/669)

### Bug Fixes

### Breaking changes

### Improvements

- Added support for `fn::split` built-in function to split strings into arrays.
  [#281](https://github.com/pulumi/esc/issues/281)
- Add native support for OIDC token exchange when logging into Pulumi Cloud. Run `esc login --help` for more
  information. [#607](https://github.com/pulumi/esc/pull/607)
- Add `esc setup aws` command to configure AWS OIDC integration for Pulumi ESC. This command creates the
  necessary AWS resources (OIDC provider, IAM role, policy attachment) and optionally creates an ESC
  environment with the OIDC configuration.

### Bug Fixes

### Breaking changes

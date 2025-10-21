### Improvements

- Added support for Open Approvals [#592](https://github.com/pulumi/esc/pull/592)
- Added deletion protection for environments:
  - Use `esc env settings set [<org-name>/][<project-name>/]<environment-name> deletion-protected true` to enable deletion protection
  - Use `esc env settings get [<org-name>/][<project-name>/]<environment-name> [deletion-protected]` to check the current status
  - When enabled, environments cannot be deleted until protection is disabled
  - Deletion protection is disabled by default for new environments

### Bug Fixes

### Breaking changes

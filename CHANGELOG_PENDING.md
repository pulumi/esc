### Improvements

- Added deletion protection for environments:
  - Use `esc env settings set [<org-name>/][<project-name>/]<environment-name> deletion-protected true` to enable deletion protection
  - Use `esc env settings get [<org-name>/][<project-name>/]<environment-name> [deletion-protected]` to check the current status
  - When enabled, environments cannot be deleted until protection is disabled
  - Deletion protection is disabled by default for new environments

### Bug Fixes

- Environment declarations are now returned even in the face of syntax errors.

### Breaking changes

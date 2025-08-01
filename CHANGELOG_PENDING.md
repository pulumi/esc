### Improvements

* Added ESC-specific environment variables to `esc run` command:
  * `PULUMI_ESC_ORG`
  * `PULUMI_ESC_PROJECT`
  * `PULUMI_ESC_ENVIRONMENT`

  These environment variables can be excluded with the `--exclude-env-vars` flag.

  These environment variables allow to detect whether the command executed by `esc run` is running in an ESC environment.

### Bug Fixes

### Breaking changes

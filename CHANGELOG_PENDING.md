### Improvements

* `env run` in interaactive mode replaces itself using the `exec` command, reducing memory usage and
  enabling using `esc` to replace the current shell with one with an opened environment.

### Bug Fixes

* Fix a nil pointer dereference in Syntax.NodeError. [#180](https://github.com/pulumi/esc/pull/180)

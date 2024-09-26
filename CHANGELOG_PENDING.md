### Improvements

- Improve evaluation performance and memory footprint.
  [#392](https://github.com/pulumi/esc/pull/392)

- Prompt for the value if the value arg is not passed.
  [#394](https://github.com/pulumi/esc/pull/394)

### Bug Fixes

### Breaking changes

- `schema`: `ObjectBuilder.Properties` and `Record` now take a `MapBuilder` in order to avoid copies.
  [#392](https://github.com/pulumi/esc/pull/392)

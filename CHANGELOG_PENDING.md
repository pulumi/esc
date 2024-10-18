### Improvements

- Improve evaluation performance and memory footprint.
  [#392](https://github.com/pulumi/esc/pull/392)

- Improve login error message when credentials file is missing or invalid.
  [#404](https://github.com/pulumi/esc/pull/404)

### Bug Fixes

- Fix panic when object keys are not strings.
  [#406](https://github.com/pulumi/esc/pull/406)

### Breaking changes

- `schema`: `ObjectBuilder.Properties` and `Record` now take a `MapBuilder` in order to avoid copies.
  [#392](https://github.com/pulumi/esc/pull/392)

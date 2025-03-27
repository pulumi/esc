### Improvements

- Updated to go 1.23

### Bug Fixes

- cmd/esc/cli/env.go: Modified the writeYAMLEnvironmentDiagnostics function to instantiate hcl.NewDiagnosticTextWriter with a width of 0 initially, and then conditionally reinstantiate it with the specified width if it is greater than 0, addressing gosec G115. [#494](https://github.com/pulumi/esc/pull/494)
- No longer error when decrypting invalid secrets outside of values top-level key
  [#491](https://github.com/pulumi/esc/pull/491)
- Make CLI prefer environment variable `PULUMI_BACKEND_URL` over account backend URL
  [#477](https://github.com/pulumi/esc/pull/477)
- Adding cascading secrets into `NewSecret` method
  [#488](https://github.com/pulumi/esc/pull/488)

### Breaking changes

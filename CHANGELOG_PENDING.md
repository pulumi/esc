### Improvements

- eval: `Value.UnmarshalJSON` now decodes the entire subtree with a single
  shared `json.Decoder`, removing the per-level double scan that made
  decoding cost O(size × depth) on deeply merged payloads.
- eval: introduce `eval.EvalOptions` and `eval.TraceMode`. The default
  `TraceModeCollapsed` keeps only the immediate `Trace.Base` parent on each
  value, bounding serialized opened-payload size to O(1) per leaf regardless
  of import-merge depth.

### Bug Fixes

### Breaking changes

- eval: `EvalEnvironment`, `CheckEnvironment`, and `RotateEnvironment` now
  take a final `EvalOptions` argument. Existing call sites should pass
  `eval.EvalOptions{}` to keep the new safe default, or
  `eval.EvalOptions{TraceMode: eval.TraceModeFull}` to preserve the prior
  full-chain serialization behavior.

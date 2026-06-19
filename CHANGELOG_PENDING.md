### Improvements

- eval: introduce `eval.EvalOptions` and `eval.TraceMode`. `TraceModeNone`
  exports each value without its `Trace.Base` merge-history chain, bounding
  serialized opened-payload size by its logical content instead of by
  import-merge depth. The chain is omitted directly during export, so the
  dropped data is never allocated. The default, `TraceModeFull`, is the zero
  value and preserves the entire chain, so no consumer of `Trace.Base` (such as
  the `esc env get` provenance view) regresses without opting in.

### Bug Fixes

### Breaking changes

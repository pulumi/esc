// Copyright 2026, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package eval

// TraceMode controls how much of each value's Trace.Base merge-history chain
// the eval entry points retain in their result.
//
// The chain is only useful to a consumer that walks it end-to-end (e.g. the
// `esc env get` provenance view, which runs on the Check path). A consumer that
// never reads Trace.Base — such as the service storing opened environments —
// pays for the chain without using it, and because object merges set Trace.Base
// on every produced value the cost grows with import-merge depth. TraceModeNone
// lets such callers drop it so opened payloads stay bounded by their logical
// content.
type TraceMode int

const (
	// TraceModeFull preserves the entire Trace.Base chain. It is the zero value,
	// so callers passing EvalOptions{} keep the historical behavior and no
	// existing consumer of the chain regresses.
	TraceModeFull TraceMode = iota

	// TraceModeNone drops Trace.Base entirely: each value is exported without its
	// merge-history chain. Cheapest payload; safe only when no consumer walks the
	// chain.
	TraceModeNone
)

// EvalOptions configures an evaluation. New fields must default to today's
// behavior when zero-valued so callers passing EvalOptions{} get the historical
// behavior without reasoning about the option surface.
type EvalOptions struct {
	// TraceMode selects how much merge-history to retain on each Value. The base
	// chain is built (or not) directly during export, so TraceModeNone never
	// allocates the dropped chain in the first place.
	TraceMode TraceMode
}

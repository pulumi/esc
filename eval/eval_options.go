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

import "github.com/pulumi/esc"

// TraceMode controls how much of each value's Trace.Base merge-history chain
// is retained in the result returned by the eval entry points.
//
// The zero value is TraceModeCollapsed: it is the safe default that keeps
// debug-friendly per-value provenance bounded to O(1) per leaf, avoiding the
// O(size * depth) payload blowup that occurs when many imports merge into a
// shared map.
type TraceMode int

const (
	// TraceModeCollapsed keeps each value's immediate parent (one level of
	// Trace.Base) and drops the rest of the chain. This is the zero value so
	// existing callers get the safe default without source changes.
	TraceModeCollapsed TraceMode = iota

	// TraceModeFull preserves the entire Trace.Base chain. Intended for
	// debugging / provenance views that need the full merge history.
	TraceModeFull

	// TraceModeNone drops Trace.Base entirely. Cheapest payload; loses all
	// merge provenance.
	TraceModeNone
)

// EvalOptions configures an evaluation. New fields must default to today's
// behavior when zero-valued so callers passing EvalOptions{} get a sensible
// default without thinking about the option surface.
type EvalOptions struct {
	// TraceMode selects how much merge-history to retain on each Value.
	TraceMode TraceMode
}

// applyTraceMode rewrites v's Trace.Base chain per mode, recursing through
// the data dimension (array elements, map values). Mutates in place.
func applyTraceMode(v *esc.Value, mode TraceMode) {
	if v == nil || mode == TraceModeFull {
		return
	}

	switch mode {
	case TraceModeNone:
		v.Trace.Base = nil
	case TraceModeCollapsed:
		if v.Trace.Base != nil {
			// Keep the immediate parent but strip everything reachable from
			// it — both its own Base and its nested children's Base chains.
			// Without this, the kept parent's data tree re-introduces deep
			// chains and the collapse buys nothing.
			applyTraceMode(v.Trace.Base, TraceModeNone)
		}
	}

	switch inner := v.Value.(type) {
	case []esc.Value:
		for i := range inner {
			applyTraceMode(&inner[i], mode)
		}
	case map[string]esc.Value:
		for k, child := range inner {
			applyTraceMode(&child, mode)
			inner[k] = child
		}
	}
}

// applyTraceModeToEnvironment walks an evaluated environment and applies the
// given trace mode to every value in Properties and the ExecutionContext.
func applyTraceModeToEnvironment(env *esc.Environment, mode TraceMode) {
	if env == nil || mode == TraceModeFull {
		return
	}
	for k, v := range env.Properties {
		applyTraceMode(&v, mode)
		env.Properties[k] = v
	}
	if env.ExecutionContext != nil {
		for k, v := range env.ExecutionContext.Properties {
			applyTraceMode(&v, mode)
			env.ExecutionContext.Properties[k] = v
		}
	}
}

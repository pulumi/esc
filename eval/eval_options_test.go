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

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/pulumi/esc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chainValue builds a Value whose Trace.Base chain has the given depth. The
// payload at each level is distinct so callers can identify which level was
// kept after rewriting.
func chainValue(depth int) esc.Value {
	if depth <= 0 {
		return esc.Value{}
	}
	v := esc.Value{Value: "level-0"}
	cur := &v
	for i := 1; i < depth; i++ {
		cur.Trace.Base = &esc.Value{Value: "level-" + string(rune('0'+i))}
		cur = cur.Trace.Base
	}
	return v
}

func chainDepth(v *esc.Value) int {
	n := 0
	for cur := v; cur != nil; cur = cur.Trace.Base {
		n++
	}
	return n
}

func TestApplyTraceMode_Full_IsNoOp(t *testing.T) {
	v := chainValue(5)
	applyTraceMode(&v, TraceModeFull)
	assert.Equal(t, 5, chainDepth(&v))
}

func TestApplyTraceMode_None_DropsAllBase(t *testing.T) {
	v := chainValue(5)
	applyTraceMode(&v, TraceModeNone)
	assert.Nil(t, v.Trace.Base)
}

func TestApplyTraceMode_Collapsed_KeepsImmediateParent(t *testing.T) {
	v := chainValue(5)
	applyTraceMode(&v, TraceModeCollapsed)

	// depth 2: the value plus exactly one parent.
	assert.Equal(t, 2, chainDepth(&v))
	assert.NotNil(t, v.Trace.Base)
	assert.Equal(t, "level-1", v.Trace.Base.Value)
	assert.Nil(t, v.Trace.Base.Trace.Base)
}

func TestApplyTraceMode_Collapsed_NilBaseStaysNil(t *testing.T) {
	v := esc.Value{Value: "leaf"}
	applyTraceMode(&v, TraceModeCollapsed)
	assert.Nil(t, v.Trace.Base)
}

func TestApplyTraceMode_None_ClearsBaseInsideMap(t *testing.T) {
	child := chainValue(3)
	v := esc.Value{Value: map[string]esc.Value{"k": child}}

	applyTraceMode(&v, TraceModeNone)

	got := v.Value.(map[string]esc.Value)["k"]
	assert.Nil(t, got.Trace.Base)
}

func TestApplyTraceMode_None_ClearsBaseInsideArray(t *testing.T) {
	child := chainValue(3)
	v := esc.Value{Value: []esc.Value{child}}

	applyTraceMode(&v, TraceModeNone)

	got := v.Value.([]esc.Value)[0]
	assert.Nil(t, got.Trace.Base)
}

func TestApplyTraceMode_Collapsed_ClearsBaseInsideKeptParent(t *testing.T) {
	// Build: leaf -> parent (with its own nested child carrying a deep chain).
	// After Collapsed we keep `parent`, but the parent's nested child's chain
	// must be cleared — otherwise the kept-parent shortcut re-introduces the
	// O(depth) payload through its data dimension.
	nestedChain := chainValue(4)
	parent := esc.Value{Value: map[string]esc.Value{"n": nestedChain}}
	leaf := esc.Value{Value: "leaf", Trace: esc.Trace{Base: &parent}}

	applyTraceMode(&leaf, TraceModeCollapsed)

	// Parent itself is kept (depth 2 from leaf).
	assert.Equal(t, 2, chainDepth(&leaf))

	// But the parent's nested map element must have a zero-length chain.
	keptParent := leaf.Trace.Base
	keptNested := keptParent.Value.(map[string]esc.Value)["n"]
	assert.Nil(t, keptNested.Trace.Base,
		"nested child inside the kept parent must have its base cleared")
}

// TestEvalEnvironment_TraceModeOption proves the EvalOptions param actually
// reaches the evaluation result. Depth-collapse correctness is covered by the
// applyTraceMode unit tests; this test only asserts the option flows through.
func TestEvalEnvironment_TraceModeOption(t *testing.T) {
	basePath := filepath.Join("testdata", "eval", "import-merge")
	envBytes, err := os.ReadFile(filepath.Join(basePath, "env.yaml"))
	require.NoError(t, err)

	execContext, err := esc.NewExecContext(map[string]esc.Value{})
	require.NoError(t, err)

	loadAndEval := func(mode TraceMode) *esc.Value {
		env, _, err := LoadYAMLBytes("import-merge", envBytes)
		require.NoError(t, err)
		result, _ := EvalEnvironment(context.Background(), "import-merge", env, rot128{}, testProviders{},
			&testEnvironments{basePath}, execContext, EvalOptions{TraceMode: mode})
		require.NotNil(t, result)
		v := result.Properties["some_object"]
		return &v
	}

	full := loadAndEval(TraceModeFull)
	assert.NotNil(t, full.Trace.Base, "Full must preserve some_object's Base")

	none := loadAndEval(TraceModeNone)
	assert.Nil(t, none.Trace.Base, "None must drop some_object's Base")

	collapsed := loadAndEval(TraceModeCollapsed)
	assert.NotNil(t, collapsed.Trace.Base, "Collapsed must keep the immediate parent")
	assert.Nil(t, collapsed.Trace.Base.Trace.Base, "Collapsed must drop the grandparent")
}

// inlineEnvironments implements EnvironmentLoader from an in-memory map so
// tests can build arbitrary import topologies without scattering YAML files
// in testdata.
type inlineEnvironments map[string][]byte

func (e inlineEnvironments) LoadEnvironment(_ context.Context, name string) ([]byte, Decrypter, error) {
	src, ok := e[name]
	if !ok {
		return nil, nil, os.ErrNotExist
	}
	return src, rot128{}, nil
}

// TestEvalEnvironment_ZeroOptionsDefaultsToCollapsed locks in the headline
// promise of this change: callers who pass EvalOptions{} get the safe
// (Collapsed) default without thinking about the option surface.
func TestEvalEnvironment_ZeroOptionsDefaultsToCollapsed(t *testing.T) {
	envs := inlineEnvironments{
		"a":    []byte("values:\n  shared: from-a\n"),
		"b":    []byte("imports:\n  - a\nvalues:\n  shared: from-b\n"),
		"root": []byte("imports:\n  - b\nvalues:\n  shared: from-root\n"),
	}
	env, _, err := LoadYAMLBytes("root", envs["root"])
	require.NoError(t, err)

	execContext, err := esc.NewExecContext(map[string]esc.Value{})
	require.NoError(t, err)

	result, _ := EvalEnvironment(context.Background(), "root", env, rot128{}, testProviders{},
		envs, execContext, EvalOptions{})
	require.NotNil(t, result)

	shared := result.Properties["shared"]
	require.NotNil(t, shared.Trace.Base, "Collapsed default must keep the immediate parent")
	assert.Nil(t, shared.Trace.Base.Trace.Base, "Collapsed default must drop the grandparent")
}

// TestEvalEnvironment_Collapsed_DeepImportChain exercises a chain deeper than
// what the testdata import-merge fixture produces, so Full vs Collapsed
// actually diverge in the assertion.
func TestEvalEnvironment_Collapsed_DeepImportChain(t *testing.T) {
	envs := inlineEnvironments{
		"a":    []byte("values:\n  shared: from-a\n"),
		"b":    []byte("imports:\n  - a\nvalues:\n  shared: from-b\n"),
		"c":    []byte("imports:\n  - b\nvalues:\n  shared: from-c\n"),
		"d":    []byte("imports:\n  - c\nvalues:\n  shared: from-d\n"),
		"root": []byte("imports:\n  - d\nvalues:\n  shared: from-root\n"),
	}

	loadAndEval := func(mode TraceMode) esc.Value {
		env, _, err := LoadYAMLBytes("root", envs["root"])
		require.NoError(t, err)
		execContext, err := esc.NewExecContext(map[string]esc.Value{})
		require.NoError(t, err)
		result, _ := EvalEnvironment(context.Background(), "root", env, rot128{}, testProviders{},
			envs, execContext, EvalOptions{TraceMode: mode})
		require.NotNil(t, result)
		return result.Properties["shared"]
	}

	full := loadAndEval(TraceModeFull)
	assert.Greater(t, chainDepth(&full), 2,
		"Full must preserve a chain deeper than 2 for this 5-environment fixture")

	collapsed := loadAndEval(TraceModeCollapsed)
	assert.Equal(t, 2, chainDepth(&collapsed),
		"Collapsed must cap the chain at exactly depth 2")

	none := loadAndEval(TraceModeNone)
	assert.Equal(t, 1, chainDepth(&none),
		"None must reduce the chain to depth 1 (root value only)")
}

// TestApplyTraceModeToEnvironment_AppliesToExecutionContext guards against a
// regression where the strip walks env.Properties but forgets the
// ExecutionContext side, which is also a map of values that can carry
// merge-history chains.
func TestApplyTraceModeToEnvironment_AppliesToExecutionContext(t *testing.T) {
	env := &esc.Environment{
		Properties: map[string]esc.Value{
			"prop": chainValue(4),
		},
		ExecutionContext: &esc.EvaluatedExecutionContext{
			Properties: map[string]esc.Value{
				"ctx": chainValue(4),
			},
		},
	}

	applyTraceModeToEnvironment(env, TraceModeCollapsed)

	got := env.Properties["prop"]
	assert.Equal(t, 2, chainDepth(&got), "Properties must be collapsed")
	gotCtx := env.ExecutionContext.Properties["ctx"]
	assert.Equal(t, 2, chainDepth(&gotCtx), "ExecutionContext.Properties must be collapsed")
}

// TestApplyTraceMode_Collapsed_ArrayElements complements the map case: array
// elements each carry independent Base chains and must each be collapsed.
func TestApplyTraceMode_Collapsed_ArrayElements(t *testing.T) {
	v := esc.Value{Value: []esc.Value{
		chainValue(4),
		chainValue(7),
	}}

	applyTraceMode(&v, TraceModeCollapsed)

	arr := v.Value.([]esc.Value)
	a, b := arr[0], arr[1]
	assert.Equal(t, 2, chainDepth(&a))
	assert.Equal(t, 2, chainDepth(&b))
}

func TestApplyTraceMode_Collapsed_RecursesIntoDataDimension(t *testing.T) {
	// Each map element has its own Base chain. Collapsed should truncate each
	// element to depth 2 independently.
	v := esc.Value{Value: map[string]esc.Value{
		"a": chainValue(4),
		"b": chainValue(6),
	}}

	applyTraceMode(&v, TraceModeCollapsed)

	got := v.Value.(map[string]esc.Value)
	a, b := got["a"], got["b"]
	assert.Equal(t, 2, chainDepth(&a))
	assert.Equal(t, 2, chainDepth(&b))
}

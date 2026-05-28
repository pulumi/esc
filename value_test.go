// Copyright 2023, Pulumi Corporation.
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

package esc

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecret(t *testing.T) {
	cases := []struct {
		name      string
		newSecret func() Value
		expected  Value
	}{
		{
			name:      "bool",
			newSecret: func() Value { return NewSecret(true) },
			expected:  Value{Value: true, Secret: true},
		},
		{
			name:      "number",
			newSecret: func() Value { return NewSecret(json.Number("3.14")) },
			expected:  Value{Value: json.Number("3.14"), Secret: true},
		},
		{
			name:      "string",
			newSecret: func() Value { return NewSecret("hello") },
			expected:  Value{Value: "hello", Secret: true},
		},
		{
			name:      "array",
			newSecret: func() Value { return NewSecret([]Value{NewValue([]Value{NewValue("hello"), NewValue("world")})}) },
			expected:  Value{Value: []Value{{Value: []Value{{Value: "hello", Secret: true}, {Value: "world", Secret: true}}, Secret: true}}, Secret: true},
		},
		{
			name: "object",
			newSecret: func() Value {
				return NewSecret(map[string]Value{"nest": NewValue(map[string]Value{"hello": NewValue("world")})})
			},
			expected: Value{Value: map[string]Value{"nest": {Value: map[string]Value{"hello": {Value: "world", Secret: true}}, Secret: true}}, Secret: true},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := c.newSecret()
			assert.Equal(t, actual, c.expected)
		})
	}
}

func TestValueUnmarshalJSON(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected Value
	}{
		{
			name:     "empty object",
			input:    `{}`,
			expected: Value{},
		},
		{
			name:     "null value",
			input:    `{"value":null}`,
			expected: Value{},
		},
		{
			name:     "bool true",
			input:    `{"value":true}`,
			expected: Value{Value: true},
		},
		{
			name:     "bool false",
			input:    `{"value":false}`,
			expected: Value{Value: false},
		},
		{
			name:     "number decoded as json.Number",
			input:    `{"value":3.14}`,
			expected: Value{Value: json.Number("3.14")},
		},
		{
			name:     "string",
			input:    `{"value":"hello"}`,
			expected: Value{Value: "hello"},
		},
		{
			name:     "secret and unknown flags",
			input:    `{"value":"x","secret":true,"unknown":true}`,
			expected: Value{Value: "x", Secret: true, Unknown: true},
		},
		{
			name:  "array of scalars",
			input: `{"value":[{"value":"a"},{"value":"b"}]}`,
			expected: Value{Value: []Value{
				{Value: "a"},
				{Value: "b"},
			}},
		},
		{
			name:  "object of scalars",
			input: `{"value":{"k1":{"value":"a"},"k2":{"value":"b"}}}`,
			expected: Value{Value: map[string]Value{
				"k1": {Value: "a"},
				"k2": {Value: "b"},
			}},
		},
		{
			name:  "nested map of arrays",
			input: `{"value":{"k":{"value":[{"value":1},{"value":2}]}}}`,
			expected: Value{Value: map[string]Value{
				"k": {Value: []Value{
					{Value: json.Number("1")},
					{Value: json.Number("2")},
				}},
			}},
		},
		{
			name:  "trace.base chain preserved",
			input: `{"value":"leaf","trace":{"base":{"value":"parent","trace":{"base":{"value":"grandparent"}}}}}`,
			expected: Value{
				Value: "leaf",
				Trace: Trace{Base: &Value{
					Value: "parent",
					Trace: Trace{Base: &Value{
						Value: "grandparent",
					}},
				}},
			},
		},
		{
			name:     "unknown top-level fields are ignored",
			input:    `{"value":"x","futureField":{"nested":true},"anotherFuture":42}`,
			expected: Value{Value: "x"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got Value
			require.NoError(t, json.Unmarshal([]byte(c.input), &got))
			assert.Equal(t, c.expected, got)
		})
	}
}

func TestValueUnmarshalJSON_RoundTrip(t *testing.T) {
	// Tree mixes secrets, arrays, maps, scalars, and a Base chain so the
	// rewrite is exercised against a representative shape in one pass.
	orig := Value{
		Value: map[string]Value{
			"arr": {Value: []Value{
				{Value: "s", Secret: true},
				{Value: json.Number("42")},
			}},
			"nested": {
				Value: map[string]Value{"k": {Value: true}},
				Trace: Trace{Base: &Value{Value: "before"}},
			},
			"unk": {Unknown: true},
		},
	}

	data, err := json.Marshal(orig)
	require.NoError(t, err)

	var got Value
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, orig, got)
}

func TestValueUnmarshalJSON_Errors(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"not an object", `123`},
		{"malformed json", `{`},
		{"object with non-string key seems impossible in JSON but truncated input errors", `{"value":`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var got Value
			err := json.Unmarshal([]byte(c.input), &got)
			require.Error(t, err)
		})
	}
}

func TestValueUnmarshalJSON_DeepNesting(t *testing.T) {
	// Build a deeply nested object to confirm the decoder does not blow
	// the stack or mis-handle depth. The original double-unmarshal path
	// re-scans each level; this is a guard against accidental regressions
	// during rewrite work.
	const depth = 100
	input := strings.Repeat(`{"value":{"k":`, depth) + `{"value":"leaf"}` + strings.Repeat(`}}`, depth)

	var got Value
	require.NoError(t, json.Unmarshal([]byte(input), &got))

	cur := &got
	for i := range depth {
		m, ok := cur.Value.(map[string]Value)
		require.Truef(t, ok, "expected map at depth %d, got %T", i, cur.Value)
		next := m["k"]
		cur = &next
	}
	assert.Equal(t, "leaf", cur.Value)
}

// BenchmarkValueUnmarshalJSON_Depth exercises UnmarshalJSON against a Trace.Base
// chain — the shape that produced the May 2026 CPU step when an
// import-heavy environment doubled its opened payload. Running with several
// depths makes super-linear regressions in this hot path visible in the
// benchmark output.
func BenchmarkValueUnmarshalJSON_Depth(b *testing.B) {
	for _, depth := range []int{1, 5, 20, 50} {
		b.Run("depth="+strconv.Itoa(depth), func(b *testing.B) {
			payload := buildTraceChain(depth)
			data, err := json.Marshal(payload)
			require.NoError(b, err)

			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for b.Loop() {
				var v Value
				if err := json.Unmarshal(data, &v); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func buildTraceChain(depth int) Value {
	// Approximate the webflow shape: a map of ~10 string leaves at each level,
	// chained via Trace.Base. The leaf width keeps the per-level work non-trivial
	// without making the benchmark dominated by allocator churn.
	makeLeaf := func() Value {
		m := make(map[string]Value, 10)
		for i := range 10 {
			m["k"+strconv.Itoa(i)] = Value{Value: "value-" + strconv.Itoa(i)}
		}
		return Value{Value: m}
	}

	root := makeLeaf()
	cur := &root
	for range depth - 1 {
		parent := makeLeaf()
		cur.Trace.Base = &parent
		cur = cur.Trace.Base
	}
	return root
}

func TestMakeSecret(t *testing.T) {
	cases := []struct {
		name     string
		value    Value
		expected Value
	}{
		{
			name:     "zero",
			value:    Value{},
			expected: Value{Secret: true},
		},
		{
			name:     "bool",
			value:    NewValue(true),
			expected: Value{Value: true, Secret: true},
		},
		{
			name:     "number",
			value:    NewValue(json.Number("3.14")),
			expected: Value{Value: json.Number("3.14"), Secret: true},
		},
		{
			name:     "string",
			value:    NewValue("hello"),
			expected: Value{Value: "hello", Secret: true},
		},
		{
			name:     "array",
			value:    NewValue([]Value{NewValue([]Value{NewValue("hello"), NewValue("world")})}),
			expected: Value{Value: []Value{{Value: []Value{{Value: "hello", Secret: true}, {Value: "world", Secret: true}}, Secret: true}}, Secret: true},
		},
		{
			name:     "object",
			value:    NewValue(map[string]Value{"nest": NewValue(map[string]Value{"hello": NewValue("world")})}),
			expected: Value{Value: map[string]Value{"nest": {Value: map[string]Value{"hello": {Value: "world", Secret: true}}, Secret: true}}, Secret: true},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := c.value.MakeSecret()
			assert.Equal(t, actual, c.expected)
		})
	}
}

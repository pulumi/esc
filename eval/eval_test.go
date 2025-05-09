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

package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/pgavlin/fx/v2"
	fxm "github.com/pgavlin/fx/v2/maps"
	fxs "github.com/pgavlin/fx/v2/slices"
	"github.com/pgavlin/fx/v2/try"
	"github.com/pulumi/esc"
	"github.com/pulumi/esc/schema"
	"github.com/pulumi/esc/syntax"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func accept() bool {
	return cmdutil.IsTruthy(os.Getenv("PULUMI_ACCEPT"))
}

type errorProvider struct{}

func (errorProvider) Schema() (*schema.Schema, *schema.Schema) {
	return schema.Record(schema.BuilderMap{"why": schema.String()}).Schema(), schema.Always()
}

func (errorProvider) Open(ctx context.Context, inputs map[string]esc.Value, context esc.EnvExecContext) (esc.Value, error) {
	return esc.Value{}, errors.New(inputs["why"].Value.(string))
}

type testSchemaProvider struct{}

func (testSchemaProvider) Schema() (*schema.Schema, *schema.Schema) {
	s := schema.Object().
		Defs(schema.BuilderMap{
			"defRecord": schema.Record(schema.BuilderMap{
				"baz": schema.String().Const("qux"),
			}),
		}).
		Properties(schema.BuilderMap{
			"null":    schema.Null(),
			"boolean": schema.Boolean(),
			"false":   schema.Boolean().Const(false),
			"true":    schema.Boolean().Const(true),
			"number":  schema.Number(),
			"pi":      schema.Number().Const("3.14"),
			"string":  schema.String(),
			"hello":   schema.String().Const("hello"),
			"array":   schema.Array().Items(schema.Always()),
			"tuple":   schema.Tuple(schema.String().Const("hello"), schema.String().Const("world")),
			"map":     schema.Object().AdditionalProperties(schema.Always()),
			"record": schema.Record(schema.BuilderMap{
				"foo": schema.String(),
			}),
			"anyOf": schema.AnyOf(schema.String(), schema.Number()),
			"oneOf": schema.OneOf(schema.String(), schema.Number()),
			"ref":   schema.Ref("#/$defs/defRecord"),

			// Complex cases
			"const-array":  &schema.Schema{Type: "array", Const: []any{"hello", json.Number("42")}},
			"const-object": &schema.Schema{Type: "object", Const: map[string]any{"hello": "world"}},
			"enum":         schema.String().Enum("foo", "bar"),
			"never":        schema.Never(),
			"always":       schema.Always(),
			"double":       schema.Tuple(schema.String(), schema.Number()),
			"triple":       schema.Tuple(schema.String(), schema.Number(), schema.Boolean()),
			"dependentReq": schema.Object().
				Properties(schema.BuilderMap{
					"foo": schema.String(),
					"bar": schema.Number(),
				}).DependentRequired(map[string][]string{"foo": {"bar"}}),
			"multiple":         schema.Number().MultipleOf(json.Number("2")),
			"minimum":          schema.Number().Minimum(json.Number("1")),
			"exclusiveMinimum": schema.Number().ExclusiveMinimum(json.Number("1")),
			"maximum":          schema.Number().Maximum(json.Number("1")),
			"exclusiveMaximum": schema.Number().ExclusiveMaximum(json.Number("1")),
			"minLength":        schema.String().MinLength(1),
			"maxLength":        schema.String().MaxLength(1),
			"pattern":          schema.String().Pattern(`^foo[0-9]+$`),
			"minItems":         schema.Array().MinItems(3),
			"maxItems":         schema.Array().MaxItems(2),
			"minProperties":    schema.Object().MinProperties(1),
			"maxProperties":    schema.Object().MaxProperties(1),
		}).
		Schema()

	return s, s
}

func (testSchemaProvider) Open(ctx context.Context, inputs map[string]esc.Value, context esc.EnvExecContext) (esc.Value, error) {
	return esc.NewValue(inputs), nil
}

type benchProvider struct {
	delay time.Duration
}

func (benchProvider) Schema() (*schema.Schema, *schema.Schema) {
	return schema.Always(), schema.Always()
}

func (p benchProvider) Open(ctx context.Context, inputs map[string]esc.Value, context esc.EnvExecContext) (esc.Value, error) {
	time.Sleep(p.delay)
	return esc.NewValue(p.delay.String()), nil
}

type testProvider struct{}

func (testProvider) Schema() (*schema.Schema, *schema.Schema) {
	return schema.Always(), schema.Always()
}

func (testProvider) Open(ctx context.Context, inputs map[string]esc.Value, context esc.EnvExecContext) (esc.Value, error) {
	return esc.NewValue(inputs), nil
}

type secretWrapperProvider struct{}

func (secretWrapperProvider) Schema() (*schema.Schema, *schema.Schema) {
	return schema.Always(), schema.Always()
}

func (secretWrapperProvider) Open(ctx context.Context, inputs map[string]esc.Value, context esc.EnvExecContext) (esc.Value, error) {
	return esc.Value{Value: inputs, Secret: true}, nil
}

type swapRotator struct{}

func (swapRotator) Schema() (*schema.Schema, *schema.Schema, *schema.Schema) {
	inputSchema := schema.Always()
	stateSchema := schema.Record(schema.BuilderMap{
		"a": schema.String(),
		"b": schema.String(),
	}).Schema()
	outputSchema := schema.Record(schema.BuilderMap{
		"a": schema.String(),
		"b": schema.String(),
	}).Schema()
	return inputSchema, stateSchema, outputSchema
}

func (swapRotator) Open(ctx context.Context, inputs, state map[string]esc.Value, context esc.EnvExecContext) (esc.Value, error) {
	return esc.NewValue(state), nil
}

func (swapRotator) Rotate(ctx context.Context, inputs, state map[string]esc.Value, context esc.EnvExecContext) (esc.Value, error) {
	newState := esc.NewValue(map[string]esc.Value{
		"a": state["b"],
		"b": state["a"],
	})
	return newState, nil
}

type echoRotator struct{}

func (echoRotator) Schema() (*schema.Schema, *schema.Schema, *schema.Schema) {
	inputSchema := schema.Record(schema.BuilderMap{
		"next": schema.String(),
	}).RotateOnly("next").Schema()
	stateSchema := schema.Object().Properties(schema.BuilderMap{
		"current":  schema.String(),
		"previous": schema.String(),
	}).Required("current").Schema()
	outputSchema := stateSchema
	return inputSchema, stateSchema, outputSchema
}

func (echoRotator) Open(ctx context.Context, inputs, state map[string]esc.Value, context esc.EnvExecContext) (esc.Value, error) {
	if !inputs["next"].Unknown && inputs["next"].Value != state["current"].Value {
		return esc.Value{}, fmt.Errorf("oh no! inputs were evaluated outside of rotation")
	}

	return esc.NewValue(state), nil
}

func (echoRotator) Rotate(ctx context.Context, inputs, state map[string]esc.Value, context esc.EnvExecContext) (esc.Value, error) {
	if state == nil {
		state = map[string]esc.Value{}
	}

	next := inputs["next"]
	if next.Unknown {
		return esc.Value{}, fmt.Errorf("oh no! rotateOnly inputs were not evaluated during rotation")
	}
	current := state["current"]

	return esc.NewValue(map[string]esc.Value{
		"current":  next,
		"previous": current,
	}), nil
}

type testProviders struct {
	benchDelay time.Duration
}

func (tp testProviders) LoadProvider(ctx context.Context, name string) (esc.Provider, error) {
	switch name {
	case "error":
		return errorProvider{}, nil
	case "schema":
		return testSchemaProvider{}, nil
	case "test":
		return testProvider{}, nil
	case "secret-wrapper":
		return secretWrapperProvider{}, nil
	case "bench":
		return benchProvider{delay: tp.benchDelay}, nil
	}
	return nil, fmt.Errorf("unknown provider %q", name)
}

func (testProviders) LoadRotator(ctx context.Context, name string) (esc.Rotator, error) {
	switch name {
	case "swap":
		return swapRotator{}, nil
	case "echo":
		return echoRotator{}, nil
	}
	return nil, fmt.Errorf("unknown rotator %q", name)
}

type testEnvironments struct {
	root string
}

func (e *testEnvironments) LoadEnvironment(ctx context.Context, name string) ([]byte, Decrypter, error) {
	bytes, err := os.ReadFile(filepath.Join(e.root, name+".yaml"))
	if err != nil {
		return nil, nil, err
	}
	return bytes, rot128{}, nil
}

type benchEnvironments struct {
	defs  map[string][]byte
	delay time.Duration
}

func newBenchEnvironments(root string, delay time.Duration) (*benchEnvironments, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	defs, err := fxm.TryCollect(try.UnpackAll(fx.Map(
		fxs.FMap(entries, func(e os.DirEntry) (fx.Pair[string, string], bool) {
			name, ok := strings.CutSuffix(e.Name(), ".yaml")
			return fx.Pack(name, filepath.Join(root, e.Name())), ok
		}),
		func(namepath fx.Pair[string, string]) fx.Result[fx.Pair[string, []byte]] {
			name, path := namepath.Unpack()
			bytes, err := os.ReadFile(path)
			if err != nil {
				return fx.Err[fx.Pair[string, []byte]](err)
			}
			return fx.OK(fx.Pack(name, bytes))

		},
	)))
	if err != nil {
		return nil, err
	}

	return &benchEnvironments{defs, delay}, nil
}

func (e *benchEnvironments) LoadEnvironment(ctx context.Context, name string) ([]byte, Decrypter, error) {
	time.Sleep(e.delay)

	bytes, ok := e.defs[name]
	if !ok {
		return nil, nil, os.ErrNotExist
	}
	return bytes, rot128{}, nil
}

func sortEnvironmentDiagnostics(diags syntax.Diagnostics) {
	sort.Slice(diags, func(i, j int) bool {
		di, dj := diags[i], diags[j]
		if di.Subject == nil {
			if dj.Subject == nil {
				return di.Summary < dj.Summary
			}
			return true
		}
		if dj.Subject == nil {
			return false
		}
		if di.Subject.Filename != dj.Subject.Filename {
			return di.Subject.Filename < dj.Subject.Filename
		}
		if di.Subject.Start.Line != dj.Subject.Start.Line {
			return di.Subject.Start.Line < dj.Subject.Start.Line
		}
		return di.Subject.Start.Column < dj.Subject.Start.Column
	})
}

func normalize[T any](t *testing.T, v T) T {
	var decoded T
	marshaled, err := json.Marshal(v)
	require.NoError(t, err)
	dec := json.NewDecoder(bytes.NewReader(marshaled))
	dec.UseNumber()
	err = dec.Decode(&decoded)
	require.NoError(t, err)
	return decoded
}

func TestEval(t *testing.T) {
	type testOverrides struct {
		ShowSecrets     bool     `json:"showSecrets,omitempty"`
		RootEnvironment string   `json:"rootEnvironment,omitempty"`
		Rotate          bool     `json:"rotate,omitempty"`
		RotatePaths     []string `json:"rotatePaths,omitempty"`
	}

	type expectedData struct {
		LoadDiags        syntax.Diagnostics `json:"loadDiags,omitempty"`
		CheckDiags       syntax.Diagnostics `json:"checkDiags,omitempty"`
		Check            *esc.Environment   `json:"check,omitempty"`
		CheckJSON        any                `json:"checkJson,omitempty"`
		EvalDiags        syntax.Diagnostics `json:"evalDiags,omitempty"`
		Eval             *esc.Environment   `json:"eval,omitempty"`
		EvalJSONRedacted any                `json:"evalJsonRedacted,omitempty"`
		EvalJSONRevealed any                `json:"evalJSONRevealed,omitempty"`
		RotateDiags      syntax.Diagnostics `json:"rotateDiags,omitempty"`
		Rotate           *esc.Environment   `json:"rotate,omitempty"`
		RotateJSON       any                `json:"rotateJson,omitempty"`
		RotatePatches    []*Patch           `json:"rotatePatches,omitempty"`
	}

	path := filepath.Join("testdata", "eval")
	entries, err := os.ReadDir(path)
	require.NoError(t, err)
	for _, e := range entries {
		if e.Name() == "bench" {
			continue
		}

		t.Run(e.Name(), func(t *testing.T) {
			basePath := filepath.Join(path, e.Name())
			envPath := filepath.Join(basePath, "env.yaml")
			expectedPath := filepath.Join(basePath, "expected.json")
			overridesPath := filepath.Join(basePath, "overrides.json")

			envBytes, err := os.ReadFile(envPath)
			require.NoError(t, err)

			execContext, err := esc.NewExecContext(map[string]esc.Value{
				"pulumi": esc.NewValue(map[string]esc.Value{
					"user": esc.NewValue(map[string]esc.Value{
						"id": esc.NewValue("USER_123"),
					}),
				}),
			})
			assert.NoError(t, err)

			environmentName := e.Name()
			var overrides testOverrides
			if overridesBytes, err := os.ReadFile(overridesPath); err == nil {
				err = json.Unmarshal(overridesBytes, &overrides)
				require.NoError(t, err)
			}

			if overrides.RootEnvironment != "" {
				environmentName = overrides.RootEnvironment
			}
			showSecrets := overrides.ShowSecrets

			doRotate := overrides.Rotate
			rotatePaths := make([]resource.PropertyPath, len(overrides.RotatePaths))
			for i := range overrides.RotatePaths {
				propertyPath, err := resource.ParsePropertyPath(overrides.RotatePaths[i])
				require.NoError(t, err)
				rotatePaths[i] = propertyPath
			}

			if accept() {
				env, loadDiags, err := LoadYAMLBytes(environmentName, envBytes)
				require.NoError(t, err)
				sortEnvironmentDiagnostics(loadDiags)

				check, checkDiags := CheckEnvironment(context.Background(), environmentName, env, rot128{}, testProviders{},
					&testEnvironments{basePath}, execContext, showSecrets)
				sortEnvironmentDiagnostics(checkDiags)

				actual, evalDiags := EvalEnvironment(context.Background(), environmentName, env, rot128{}, testProviders{},
					&testEnvironments{basePath}, execContext)
				sortEnvironmentDiagnostics(evalDiags)

				var rotated *esc.Environment
				var patches []*Patch
				var rotateDiags syntax.Diagnostics
				var rotationResult *RotationResult
				if doRotate {
					rotated, rotationResult, rotateDiags = RotateEnvironment(context.Background(), environmentName, env, rot128{}, testProviders{},
						&testEnvironments{basePath}, execContext, rotatePaths)
					patches = rotationResult.Patches()
				}

				var checkJSON any
				var evalJSONRedacted any
				var evalJSONRevealed any
				var rotateJSON any
				if check != nil {
					check = normalize(t, check)
					checkJSON = esc.NewValue(check.Properties).ToJSON(true)
				}
				if actual != nil {
					actual = normalize(t, actual)
					evalJSONRedacted = esc.NewValue(actual.Properties).ToJSON(true)
					evalJSONRevealed = esc.NewValue(actual.Properties).ToJSON(false)
				}
				if rotated != nil {
					rotated = normalize(t, rotated)
					rotateJSON = esc.NewValue(rotated.Properties).ToJSON(true)
				}

				bytes, err := json.MarshalIndent(expectedData{
					LoadDiags:        loadDiags,
					CheckDiags:       checkDiags,
					EvalDiags:        evalDiags,
					Check:            check,
					Eval:             actual,
					EvalJSONRedacted: evalJSONRedacted,
					EvalJSONRevealed: evalJSONRevealed,
					CheckJSON:        checkJSON,
					RotateDiags:      rotateDiags,
					Rotate:           rotated,
					RotateJSON:       rotateJSON,
					RotatePatches:    patches,
				}, "", "    ")
				bytes = append(bytes, '\n')
				require.NoError(t, err)

				err = os.WriteFile(expectedPath, bytes, 0600)
				require.NoError(t, err)

				return
			}

			var expected expectedData
			expectedBytes, err := os.ReadFile(expectedPath)
			require.NoError(t, err)
			dec := json.NewDecoder(bytes.NewReader(expectedBytes))
			dec.UseNumber()
			err = dec.Decode(&expected)
			require.NoError(t, err)

			env, diags, err := LoadYAMLBytes(environmentName, envBytes)
			require.NoError(t, err)
			sortEnvironmentDiagnostics(diags)
			require.Equal(t, expected.LoadDiags, diags)

			check, diags := CheckEnvironment(context.Background(), environmentName, env, rot128{}, testProviders{},
				&testEnvironments{basePath}, execContext, showSecrets)
			sortEnvironmentDiagnostics(diags)
			require.Equal(t, expected.CheckDiags, diags)

			actual, diags := EvalEnvironment(context.Background(), environmentName, env, rot128{}, testProviders{},
				&testEnvironments{basePath}, execContext)
			sortEnvironmentDiagnostics(diags)
			require.Equal(t, expected.EvalDiags, diags)

			var rotated *esc.Environment
			if doRotate {
				rotated_, rotationResult, diags := RotateEnvironment(context.Background(), environmentName, env, rot128{}, testProviders{},
					&testEnvironments{basePath}, execContext, rotatePaths)
				var patches []*Patch
				if rotationResult != nil {
					patches = rotationResult.Patches()
				}

				sortEnvironmentDiagnostics(diags)
				require.Equal(t, expected.RotateDiags, diags)

				slices.SortFunc(patches, func(a, b *Patch) int {
					return strings.Compare(a.DocPath, b.DocPath)
				})
				require.Equal(t, expected.RotatePatches, patches)

				rotated = rotated_
			}

			// work around a schema comparison issue due to the 'compiled' field by roundtripping through JSON
			check = normalize(t, check)
			actual = normalize(t, actual)
			rotated = normalize(t, rotated)

			// work around a comparison issue when comparing nil slices/maps against zero-length slices/maps
			if actual != nil {
				evalJSONRedacted := esc.NewValue(actual.Properties).ToJSON(true)
				assert.Equal(t, expected.EvalJSONRedacted, evalJSONRedacted)
				evalJSONRevealed := esc.NewValue(actual.Properties).ToJSON(false)
				assert.Equal(t, expected.EvalJSONRevealed, evalJSONRevealed)

				bytes, err := json.MarshalIndent(evalJSONRevealed, "", "  ")
				require.NoError(t, err)
				t.Logf("eval: %v", string(bytes))
			}

			if check != nil {
				checkJSON := esc.NewValue(check.Properties).ToJSON(true)
				assert.Equal(t, expected.CheckJSON, checkJSON)

				bytes, err := json.MarshalIndent(checkJSON, "", "  ")
				require.NoError(t, err)
				t.Logf("check: %v", string(bytes))
			}

			if rotated != nil {
				rotateJSON := esc.NewValue(check.Properties).ToJSON(true)
				assert.Equal(t, expected.CheckJSON, rotateJSON)

				bytes, err := json.MarshalIndent(rotateJSON, "", "  ")
				require.NoError(t, err)
				t.Logf("rotate: %v", string(bytes))
			}

			assert.Equal(t, expected.Check, check)
			assert.Equal(t, expected.Eval, actual)
			assert.Equal(t, expected.Rotate, rotated)
		})
	}
}

func benchmarkEval(b *testing.B, openDelay, loadDelay time.Duration) {
	basePath := filepath.Join("testdata", "eval", "bench")
	envPath := filepath.Join(basePath, "env.yaml")

	envBytes, err := os.ReadFile(envPath)
	require.NoError(b, err)

	envs, err := newBenchEnvironments(basePath, loadDelay)
	require.NoError(b, err)

	for i := 0; i < b.N; i++ {
		execContext, err := esc.NewExecContext(map[string]esc.Value{
			"pulumi": esc.NewValue(map[string]esc.Value{
				"user": esc.NewValue(map[string]esc.Value{
					"id": esc.NewValue("USER_123"),
				}),
			}),
		})
		require.NoError(b, err)

		environmentName := "bench"

		env, loadDiags, err := LoadYAMLBytes(environmentName, envBytes)
		require.NoError(b, err)
		require.Empty(b, loadDiags)

		_, evalDiags := EvalEnvironment(context.Background(), environmentName, env, rot128{}, testProviders{benchDelay: openDelay},
			envs, execContext)
		require.Empty(b, evalDiags)
	}
}

func BenchmarkEval(b *testing.B) {
	benchmarkEval(b, 0, 0)
}

func BenchmarkEvalOpen(b *testing.B) {
	benchmarkEval(b, 10*time.Millisecond, 0)
}

func BenchmarkEvalEnvLoad(b *testing.B) {
	benchmarkEval(b, 0, 10*time.Millisecond)
}

func BenchmarkEvalAll(b *testing.B) {
	benchmarkEval(b, 10*time.Millisecond, 10*time.Millisecond)
}

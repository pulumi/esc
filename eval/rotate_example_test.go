package eval

import (
	"context"
	"github.com/pulumi/esc"
	"github.com/pulumi/esc/syntax/encoding"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"strings"
	"testing"
)

func TestExampleRotate(t *testing.T) {
	const def = `
values:
  test:
    fn::open::swap:
      rotated:
        a: hunter2
        b: 2retnuh
`
	t.Logf("initial: %s", def)

	const path = "test"
	rotated1 := doRotate(t, []byte(def), path)
	t.Logf("rotated1: %s", rotated1)
	//    rotated1: values:
	//          test:
	//            fn::open::swap:
	//              rotated:
	//                a: 2retnuh
	//                b: hunter2

	rotated2 := doRotate(t, []byte(rotated1), path)
	t.Logf("rotated2: %s", rotated2)
	//    rotated2: values:
	//          test:
	//            fn::open::swap:
	//              rotated:
	//                a: hunter2
	//                b: 2retnuh
}

func TestExampleRotateInterpolation(t *testing.T) {
	const def = `
values:
  admin: foobar
  test:
    fn::open::swap:
      # anything outside of the "rotated" key won't be touched by the rewrite
      managing-user: ${admin}
      rotated:
        a: hunter2
        b: 2retnuh
`
	t.Logf("initial: %s", def)

	const path = "test"
	rotated1 := doRotate(t, []byte(def), path)
	t.Logf("rotated1: %s", rotated1)
	//		rotated1: values:
	//          admin: foobar
	//          test:
	//            fn::open::swap:
	//              # anything outside of the "rotated" key won't be touched by the rewrite
	//              managing-user: ${admin}
	//              rotated:
	//                a: 2retnuh
	//                b: hunter2

	rotated2 := doRotate(t, []byte(rotated1), path)
	t.Logf("rotated2: %s", rotated2)
	//	    rotated2: values:
	//          admin: foobar
	//          test:
	//            fn::open::swap:
	//              # anything outside of the "rotated" key won't be touched by the rewrite
	//              managing-user: ${admin}
	//              rotated:
	//                a: hunter2
	//                b: 2retnuh
}

func TestExampleRotateSecrets(t *testing.T) {
	const def = `
values:
  test:
    fn::open::swap:
      # secret outputs will be wrapped with fn::secret
      rotated:
        a: 
          fn::secret: hunter2
        b: 
          fn::secret: 2retnuh
`
	t.Logf("initial: %s", def)

	const path = "test"
	rotated1 := doRotate(t, []byte(def), path)
	t.Logf("rotated1: %s", rotated1)
	//    rotated1: values:
	//          test:
	//            fn::open::swap:
	//              # secret outputs will be wrapped with fn::secret
	//              rotated:
	//                a:
	//                  fn::secret:
	//                    ciphertext: ZXNjeAAAAAGy8uX07vXoD1eCyQ==
	//                b:
	//                  fn::secret:
	//                    ciphertext: ZXNjeAAAAAHo9e705fKyKo30VQ==

	rotated2 := doRotate(t, []byte(rotated1), path)
	t.Logf("rotated2: %s", rotated2)
	//     rotated2: values:
	//          test:
	//            fn::open::swap:
	//              # secret outputs will be wrapped with fn::secret
	//              rotated:
	//                a:
	//                  fn::secret:
	//                    ciphertext: ZXNjeAAAAAHo9e705fKyKo30VQ==
	//                b:
	//                  fn::secret:
	//                    ciphertext: ZXNjeAAAAAGy8uX07vXoD1eCyQ==
}

func doRotate(t *testing.T, def []byte, rotatedPath string) []byte {
	t.Helper()

	env, diags, err := LoadYAMLBytes("<stdin>", def)
	require.NoError(t, err)
	require.Len(t, diags, 0)

	// open the environment
	providers := testProviders{}
	execContext, err := esc.NewExecContext(nil)
	require.NoError(t, err)
	open, diags := EvalEnvironment(context.Background(), "<stdin>", env, rot128{}, providers, &testEnvironments{}, execContext)
	require.Len(t, diags, 0)

	// look up the expr based on path
	path, err := resource.ParsePropertyPath(rotatedPath)
	require.NoError(t, err)
	expr, ok := getEnvExpr(esc.Expr{Object: open.Exprs}, path)
	require.True(t, ok)
	inputs, ok := getEnvValue(esc.NewValue(open.Properties), path) // fixme: this isn't quite right, because it's the opened result
	require.True(t, ok)

	// examine the expr to find the provider
	require.NotNil(t, expr.Builtin)
	providerName, ok := strings.CutPrefix(expr.Builtin.Name, "fn::open::")
	require.True(t, ok)
	provider, err := providers.LoadProvider(context.Background(), providerName)
	require.NoError(t, err)

	// todo check inputs against schema
	//inputSchema, outputSchema := provider.Schema()
	//err = inputSchema.Compile()
	//require.NoError(t, err)
	//_ = outputSchema
	//v := validator{}
	//v.validateValue(val, inputSchema, validationLoc{x: val})

	// invoke rotator method on provider
	rotator, ok := provider.(esc.Rotator)
	require.True(t, ok)
	outputs, err := rotator.Rotate(context.Background(), inputs.Value.(map[string]esc.Value))
	require.NoError(t, err)

	// convert output to yaml
	outputJson := valueToSecretYAML(outputs)
	outputBytes, err := yaml.Marshal(outputJson)
	require.NoError(t, err)
	var outputNode yaml.Node
	err = yaml.Unmarshal(outputBytes, &outputNode)
	require.NoError(t, err)
	outputNode = *outputNode.Content[0]

	// write result back into env
	// we use the `rotated` key as an output location to avoid clobbering potential interpolation for other inputs
	outputPath := append(path, "fn::open::"+providerName, "rotated")
	var docNode yaml.Node
	err = yaml.Unmarshal([]byte(def), &docNode)
	require.NoError(t, err)
	valuesNode, ok := encoding.YAMLSyntax{Node: &docNode}.Get(resource.PropertyPath{"values"})
	require.True(t, ok)
	_, err = encoding.YAMLSyntax{Node: valuesNode}.Set(nil, outputPath, outputNode)
	require.NoError(t, err)

	newYAML, err := yaml.Marshal(docNode.Content[0])
	require.NoError(t, err)

	encryptedYAML, err := EncryptSecrets(context.Background(), "<stdin>", newYAML, rot128{})
	require.NoError(t, err)

	return encryptedYAML
}

func getEnvExpr(root esc.Expr, path resource.PropertyPath) (*esc.Expr, bool) {
	if len(path) == 0 {
		return &root, true
	}

	switch {
	case root.Builtin != nil:
		key, ok := path[0].(string)
		if !ok {
			return nil, false
		}
		if key != root.Builtin.Name {
			return nil, false
		}
		return getEnvExpr(root.Builtin.Arg, path[1:])
	case root.List != nil:
		index, ok := path[0].(int)
		if !ok || index < 0 || index >= len(root.List) {
			return nil, false
		}
		return getEnvExpr(root.List[index], path[1:])
	case root.Object != nil:
		key, ok := path[0].(string)
		if !ok {
			return nil, false
		}
		v, ok := root.Object[key]
		if !ok {
			return nil, false
		}
		return getEnvExpr(v, path[1:])
	default:
		return nil, false
	}
}

func getEnvValue(root esc.Value, path resource.PropertyPath) (*esc.Value, bool) {
	if len(path) == 0 {
		return &root, true
	}

	switch v := root.Value.(type) {
	case []esc.Value:
		index, ok := path[0].(int)
		if !ok || index < 0 || index >= len(v) {
			return nil, false
		}
		return getEnvValue(v[index], path[1:])
	case map[string]esc.Value:
		key, ok := path[0].(string)
		if !ok {
			return nil, false
		}
		e, ok := v[key]
		if !ok {
			return nil, false
		}
		return getEnvValue(e, path[1:])
	default:
		return nil, false
	}
}

func valueToSecretYAML(v esc.Value) any {
	ret := func() any {
		switch pv := v.Value.(type) {
		case []esc.Value:
			a := make([]any, len(pv))
			for i, v := range pv {
				a[i] = valueToSecretYAML(v)
			}
			return a
		case map[string]esc.Value:
			m := make(map[string]any, len(pv))
			for k, v := range pv {
				m[k] = valueToSecretYAML(v)
			}
			return m
		default:
			return pv
		}
	}()

	// wrap secret values
	if v.Secret {
		return map[string]any{
			"fn::secret": ret,
		}
	}
	return ret
}

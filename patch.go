package esc

import (
	"github.com/pulumi/esc/syntax/encoding"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"gopkg.in/yaml.v3"
)

// Patch represents a value that should be written back to the environment at the given path.
type Patch struct {
	DocPath     string
	Replacement Value
}

// ApplyPatches applies a set of patches values to an environment definition.
// If patch values contain secret values, they will be wrapped with fn::secret.
func ApplyPatches(source []byte, patches []*Patch) ([]byte, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(source, &doc); err != nil {
		return nil, err
	}

	for _, patch := range patches {
		path, err := resource.ParsePropertyPath("values." + patch.DocPath)
		if err != nil {
			return nil, err
		}

		// convert the esc.Value into a yaml node that can be set on the environment
		replacement := valueToSecretJSON(patch.Replacement)
		bytes, err := yaml.Marshal(replacement)
		if err != nil {
			return nil, err
		}
		var yamlValue yaml.Node
		if err := yaml.Unmarshal(bytes, &yamlValue); err != nil {
			return nil, err
		}
		yamlValue = *yamlValue.Content[0]

		_, err = encoding.YAMLSyntax{Node: &doc}.Set(nil, path, yamlValue)
		if err != nil {
			return nil, err
		}
	}

	return yaml.Marshal(doc.Content[0])
}

// valueToSecretJSON converts a Value into a plain-old-JSON value, but secret values are wrapped with fn::secret
func valueToSecretJSON(v Value) any {
	ret := func() any {
		switch pv := v.Value.(type) {
		case []Value:
			a := make([]any, len(pv))
			for i, v := range pv {
				a[i] = valueToSecretJSON(v)
			}
			return a
		case map[string]Value:
			m := make(map[string]any, len(pv))
			for k, v := range pv {
				m[k] = valueToSecretJSON(v)
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

// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"

	"github.com/pulumi/esc/syntax/encoding"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"gopkg.in/yaml.v3"
)

// mergeProviderIntoEnv merges providerNode into the YAML environment definition
// at values.<path>, replacing any existing node at that path. The result is the
// new YAML document bytes.
func mergeProviderIntoEnv(envYAML []byte, path resource.PropertyPath, providerNode *yaml.Node) ([]byte, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("path must contain at least one element")
	}

	var docNode yaml.Node
	if len(envYAML) > 0 {
		if err := yaml.Unmarshal(envYAML, &docNode); err != nil {
			return nil, fmt.Errorf("unmarshaling environment definition: %w", err)
		}
	}
	if docNode.Kind != yaml.DocumentNode {
		docNode = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{}}}
	}

	valuesNode, ok := encoding.YAMLSyntax{Node: &docNode}.Get(resource.PropertyPath{"values"})
	if !ok {
		var err error
		valuesNode, err = encoding.YAMLSyntax{Node: &docNode}.Set(nil, resource.PropertyPath{"values"}, yaml.Node{
			Kind: yaml.MappingNode,
		})
		if err != nil {
			return nil, fmt.Errorf("creating values node: %w", err)
		}
	}

	if _, err := (encoding.YAMLSyntax{Node: valuesNode}).Set(nil, path, *providerNode); err != nil {
		return nil, fmt.Errorf("setting provider at %v: %w", path, err)
	}

	out, err := yaml.Marshal(docNode.Content[0])
	if err != nil {
		return nil, fmt.Errorf("marshaling definition: %w", err)
	}
	return out, nil
}

// secretNode returns a yaml mapping node of the shape `fn::secret: <value>`.
// The value is always emitted as a string scalar (tag !!str), so callers do
// not have to worry about YAML coercing tokens like "true" or "12345" into
// booleans/numbers.
func secretNode(value string) *yaml.Node {
	return &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "fn::secret"},
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: value},
		},
	}
}

func applyProviderUpdate(
	ctx context.Context,
	env *envCommand,
	ref environmentRef,
	draft string,
	path resource.PropertyPath,
	providerNode *yaml.Node,
) error {
	var def []byte
	var tag string
	var err error
	if draft != "" && draft != "new" {
		def, tag, err = env.esc.client.GetEnvironmentDraft(ctx, ref.orgName, ref.projectName, ref.envName, draft)
		if err != nil {
			return fmt.Errorf("getting environment draft definition: %w", err)
		}
	} else {
		def, tag, _, err = env.esc.client.GetEnvironment(ctx, ref.orgName, ref.projectName, ref.envName, "", false)
		if err != nil {
			return fmt.Errorf("getting environment definition: %w", err)
		}
	}

	newYAML, err := mergeProviderIntoEnv(def, path, providerNode)
	if err != nil {
		return err
	}

	diags, err := env.esc.updateEnvironment(ctx, ref, draft, newYAML, tag, "Provider updated.")
	if err != nil {
		return err
	}
	if len(diags) != 0 {
		return env.writePropertyEnvironmentDiagnostics(env.esc.stderr, diags)
	}
	return nil
}

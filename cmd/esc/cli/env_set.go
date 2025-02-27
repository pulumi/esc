// Copyright 2023, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"
	"github.com/ccojocar/zxcvbn-go"
	"github.com/pulumi/esc/syntax/encoding"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"regexp"
	"strconv"
)

func newEnvSetCmd(env *envCommand) *cobra.Command {
	var secret bool
	var plaintext bool
	var filename string
	var inputYaml bool

	cmd := &cobra.Command{
		Use:   "set [<org-name>/][<project-name>/]<environment-name> <path> <value>",
		Args:  cobra.RangeArgs(2, 3),
		Short: "Set a value within an environment.",
		Long: "Set a value within an environment\n" +
			"\n" +
			"This command fetches the current definition for the named environment and modifies a\n" +
			"value within it. The path to the value to set is a Pulumi property path. The value\n" +
			"is interpreted as YAML.\n",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			ref, args, err := env.getExistingEnvRef(ctx, args)
			if err != nil {
				return err
			}
			if ref.version != "" {
				return fmt.Errorf("the set command does not accept versions")
			}
			reqArgs := 2
			if filename != "" {
				// if filename is provided, we only need a path
				reqArgs = 1
			}
			if len(args) < reqArgs {
				return fmt.Errorf("expected a path and a value")
			}

			path, err := resource.ParsePropertyPath(args[0])
			if err != nil {
				return fmt.Errorf("invalid path: %w", err)
			}
			if len(path) == 0 {
				return fmt.Errorf("path must contain at least one element")
			}

			var value []byte
			if filename == "" {
				value = []byte(args[1])
			} else {
				value, err = readFileInput(filename)
				if err != nil {
					return err
				}
			}

			var yamlValue yaml.Node
			if inputYaml {
				if err := yaml.Unmarshal(value, &yamlValue); err != nil {
					return fmt.Errorf("invalid value: %w", err)
				}
				if len(yamlValue.Content) == 0 {
					// This can happen when the value is empty (e.g. when "" is present on the command line). Treat this
					// as the empty string.
					err = yaml.Unmarshal([]byte(`""`), &yamlValue)
					contract.IgnoreError(err)
				}
				yamlValue = *yamlValue.Content[0]
			} else {
				// treat input as a raw string
				yamlValue.SetString(string(value))
			}

			if looksLikeSecret(path, yamlValue) && !secret && !plaintext {
				return fmt.Errorf("value looks like a secret; rerun with --secret to mark it as such, or --plaintext if you meant to leave it as plaintext")
			}
			if secret {
				if yamlValue.Kind == yaml.ScalarNode && yamlValue.Tag != "!!str" {
					err = yaml.Unmarshal([]byte(strconv.Quote(args[1])), &yamlValue)
					if err != nil {
						return fmt.Errorf("internal error decoding value; try surrounding the argument in both single and double quotes (e.g. '\"foo\"') (%w)", err)
					}
					yamlValue = *yamlValue.Content[0]
				}

				bytes, err := yaml.Marshal(map[string]yaml.Node{"fn::secret": yamlValue})
				if err != nil {
					return fmt.Errorf("internal error: marshaling secret: %w", err)
				}
				if err = yaml.Unmarshal(bytes, &yamlValue); err != nil {
					return fmt.Errorf("internal error: marshaling secret: %w", err)
				}
				yamlValue = *yamlValue.Content[0]
			}

			def, tag, _, err := env.esc.client.GetEnvironment(ctx, ref.orgName, ref.projectName, ref.envName, "", false)
			if err != nil {
				return fmt.Errorf("getting environment definition: %w", err)
			}

			var docNode yaml.Node
			if err := yaml.Unmarshal(def, &docNode); err != nil {
				return fmt.Errorf("unmarshaling environment definition: %w", err)
			}
			if docNode.Kind != yaml.DocumentNode {
				docNode = yaml.Node{
					Kind:    yaml.DocumentNode,
					Content: []*yaml.Node{{}},
				}
			}

			if path[0] == "imports" {
				_, err = encoding.YAMLSyntax{Node: &docNode}.Set(nil, path, yamlValue)
			} else {
				valuesNode, ok := encoding.YAMLSyntax{Node: &docNode}.Get(resource.PropertyPath{"values"})
				if !ok {
					valuesNode, err = encoding.YAMLSyntax{Node: &docNode}.Set(nil, resource.PropertyPath{"values"}, yaml.Node{
						Kind: yaml.MappingNode,
					})
					if err != nil {
						return fmt.Errorf("internal error: %w", err)
					}
				}
				_, err = encoding.YAMLSyntax{Node: valuesNode}.Set(nil, path, yamlValue)
			}
			if err != nil {
				return err
			}

			newYAML, err := yaml.Marshal(docNode.Content[0])
			if err != nil {
				return fmt.Errorf("marshaling definition: %w", err)
			}

			diags, err := env.esc.client.UpdateEnvironmentWithProject(ctx, ref.orgName, ref.projectName, ref.envName, newYAML, tag)
			if err != nil {
				return fmt.Errorf("updating environment definition: %w", err)
			}
			if len(diags) != 0 {
				return env.writePropertyEnvironmentDiagnostics(env.esc.stderr, diags)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(
		&secret, "secret", false,
		"true to mark the value as secret")
	cmd.Flags().BoolVar(
		&plaintext, "plaintext", false,
		"true to leave the value in plaintext")
	cmd.Flags().StringVar(
		&filename, "file", "", "read value from file. use `-` to read from stdin.")
	cmd.Flags().BoolVar(
		&inputYaml, "yaml", true, "treat value as a structured yaml node.")

	return cmd
}

// readFileInput reads the full content of filename, or stdin if filename is `-`
func readFileInput(filename string) (input []byte, err error) {
	if filename == "-" {
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("reading stdin: %w", err)
		}
	} else {
		input, err = os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("reading file: %w", err)
		}
	}
	return input, err
}

// keyPattern is the regular expression a configuration key must match before we check (and error) if we think
// it is a password
var keyPattern = regexp.MustCompile("(?i)passwd|pass|password|pwd|secret|token")

const (
	// maxEntropyCheckLength is the maximum length of a possible secret for entropy checking.
	maxEntropyCheckLength = 16
	// entropyThreshold is the total entropy threshold a potential secret needs to pass before being flagged.
	entropyThreshold = 80.0
	// entropyCharThreshold is the per-char entropy threshold a potential secret needs to pass before being flagged.
	entropyPerCharThreshold = 3.0
)

// looksLikeSecret returns true if a configuration value "looks" like a secret. This is always going to be a heuristic
// that suffers from false positives, but is better (a) than our prior approach of unconditionally printing a warning
// for all plaintext values, and (b)  to be paranoid about such things. Inspired by the gas linter and securego project.
func looksLikeSecret(path resource.PropertyPath, n yaml.Node) bool {
	if n.Kind != yaml.ScalarNode || n.Tag != "!!str" {
		return false
	}
	v := n.Value

	key, ok := path[len(path)-1].(string)
	if !ok || !keyPattern.MatchString(key) {
		return false
	}

	if len(v) > maxEntropyCheckLength {
		v = v[:maxEntropyCheckLength]
	}

	// Compute the strength use the resulting entropy to flag whether this looks like a secret.
	info := zxcvbn.PasswordStrength(v, nil)
	entropyPerChar := info.Entropy / float64(len(v))
	return info.Entropy >= entropyThreshold ||
		(info.Entropy >= (entropyThreshold/2) && entropyPerChar >= entropyPerCharThreshold)
}

// Copyright 2023, Pulumi Corporation.

package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pulumi/pulumi/sdk/v3/go/common/diag/colors"
	pulumienv "github.com/pulumi/pulumi/sdk/v3/go/common/env"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
)

func newEnvRmCmd(env *envCommand) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "rm [<org-name>/]<environment-name> [path]",
		Args:  cmdutil.MaximumNArgs(2),
		Short: "Remove an environment or a value from an environment.",
		Long: "Remove an environment or a value from an environment\n" +
			"\n" +
			"This command removes an environment or a value from an environment." +
			"\n" +
			"When removing an environment, the environment will no longer be available\n" +
			"once this command completes.",
		Run: cmdutil.RunFunc(func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			yes = yes || pulumienv.SkipConfirmations.Value()

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			orgName, envName, args, err := env.getEnvName(args)
			if err != nil {
				return err
			}

			// Are we removing the entire environment?
			if len(args) == 0 {
				envSlug := fmt.Sprintf("%v/%v", orgName, envName)

				// Ensure the user really wants to do this.
				prompt := fmt.Sprintf("This will permanently remove the %q environment!", envSlug)
				if !yes && !env.esc.confirmPrompt(prompt, envSlug) {
					return errors.New("confirmation declined")
				}

				err = env.esc.client.DeleteEnvironment(ctx, orgName, envName)
				if err != nil {
					return err
				}

				msg := fmt.Sprintf("%sEnvironment %q has been removed!%s", colors.SpecAttention, envSlug, colors.Reset)
				fmt.Println(env.esc.colors.Colorize(msg))
				return nil
			}

			// Otherwise, we're removing a single value.
			path, err := resource.ParsePropertyPath(args[0])
			if err != nil {
				return fmt.Errorf("invalid path: %w", err)
			}

			def, tag, err := env.esc.client.GetEnvironment(ctx, orgName, envName)
			if err != nil {
				return fmt.Errorf("getting environment definition: %w", err)
			}

			var docNode yaml.Node
			if err := yaml.Unmarshal(def, &docNode); err != nil {
				return fmt.Errorf("unmarshaling environment definition: %w", err)
			}
			if docNode.Kind != yaml.DocumentNode {
				return nil
			}
			valuesNode, ok := yamlNode{&docNode}.get(resource.PropertyPath{"values"})
			if !ok {
				return nil
			}
			err = yamlNode{valuesNode}.delete(nil, path)
			if err != nil {
				return err
			}

			newYAML, err := yaml.Marshal(docNode.Content[0])
			if err != nil {
				return fmt.Errorf("marshaling definition: %w", err)
			}

			diags, err := env.esc.client.UpdateEnvironment(ctx, orgName, envName, newYAML, tag)
			if err != nil {
				return fmt.Errorf("updating environment definition: %w", err)
			}
			if len(diags) != 0 {
				return env.writePropertyEnvironmentDiagnostics(os.Stderr, diags)
			}

			return nil
		}),
	}

	cmd.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "Skip confirmation prompts, and proceed with removal anyway")

	return cmd
}

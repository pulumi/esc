// Copyright 2023, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"
	"io"
	"io/fs"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"github.com/spf13/cobra"
)

func newEnvInitCmd(env *envCommand) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "init [<org-name>/]<environment-name>",
		Args:  cobra.MaximumNArgs(1),
		Short: "Create an empty environment with the given name.",
		Long: "Create an empty environment with the given name, ready for editing\n" +
			"\n" +
			"This command creates an empty environment with the given name. It has no definition,\n" +
			"but afterwards it can be edited using the `edit` command.\n" +
			"\n" +
			"To create an environment in an organization when logged in to the Pulumi Cloud,\n" +
			"prefix the stack name with the organization name and a slash (e.g. 'acmecorp/dev').\n",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			orgName, envName, args, err := env.getEnvName(args)
			if err != nil {
				return err
			}
			_ = args

			var yaml []byte
			switch file {
			case "":
				// OK
			case "-":
				yaml, err = io.ReadAll(env.esc.stdin)
			default:
				yaml, err = fs.ReadFile(env.esc.fs, file)
			}
			if err != nil {
				return fmt.Errorf("reading environment definition: %w", err)
			}

			if err := env.esc.client.CreateEnvironment(ctx, orgName, envName); err != nil {
				return fmt.Errorf("creating environment: %w", err)
			}
			fmt.Fprintln(env.esc.stdout, "Environment created.")
			if len(yaml) != 0 {
				diags, err := env.esc.client.UpdateEnvironment(ctx, orgName, envName, yaml, "")
				if err != nil {
					return fmt.Errorf("updating environment definition: %w", err)
				}
				if len(diags) != 0 {
					err = env.writeYAMLEnvironmentDiagnostics(env.esc.stderr, envName, yaml, diags)
					contract.IgnoreError(err)

					return fmt.Errorf("updating environment definition: too many errors")
				}
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&file,
		"file", "f", "",
		"the file to use to initialize the environment, if any. Pass `-` to read from standard input.")

	return cmd
}

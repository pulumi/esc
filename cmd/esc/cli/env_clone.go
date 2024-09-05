// Copyright 2023, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/spf13/cobra"
)

func newEnvCloneCmd(env *envCommand) *cobra.Command {
	var (
		preserveHistory         bool
		preserveAccess          bool
		preserveEnvironmentTags bool
		preserveRevisionTags    bool
	)

	cmd := &cobra.Command{
		Use:   "clone [<org-name>/]<project-name>/<environment-name> [<project-name>/]<environment-name>",
		Args:  cobra.MaximumNArgs(2),
		Short: "Clone an existing environment into a new environment.",
		Long: "Clone an existing environment into a new environment.\n" +
			"\n" +
			"This command clones an existing environment with the given identifier into a new environment.\n" +
			"If a project is omitted from the new environment identifier the new environment will be created\n" +
			"within the same project as the environment being cloned.\n",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			ref, args, err := env.getNewEnvRef(ctx, args)
			if err != nil {
				return err
			}
			if ref.version != "" {
				return fmt.Errorf("the clone command does not accept versions")
			}

			var destProject string
			destName := args[0]
			destParts := strings.Split(args[0], "/")
			if len(destParts) == 2 {
				destProject = destParts[0]
				destName = destParts[1]
			}

			destEnv := client.CloneEnvironmentRequest{
				Project:                 destProject,
				Name:                    destName,
				PreserveHistory:         preserveHistory,
				PreserveAccess:          preserveAccess,
				PreserveEnvironmentTags: preserveEnvironmentTags,
				PreserveRevisionTags:    preserveRevisionTags,
			}
			if err := env.esc.client.CloneEnvironment(ctx, ref.orgName, ref.projectName, ref.envName, destEnv); err != nil {
				return fmt.Errorf("cloning environment: %w", err)
			}

			if destProject == "" {
				destProject = ref.projectName
			}
			fmt.Fprintf(
				env.esc.stdout,
				"Environment %s/%s/%s cloned into %s/%s/%s.\n",
				ref.orgName, ref.projectName, ref.envName,
				ref.orgName, destProject, destName,
			)
			return nil
		},
	}

	cmd.Flags().BoolVar(&preserveHistory,
		"history", false,
		"preserve history of the environment being cloned")

	cmd.Flags().BoolVar(&preserveAccess,
		"access", false,
		"add the newly cloned environment to the same teams that have access to the origin environment")

	cmd.Flags().BoolVar(&preserveEnvironmentTags,
		"envTags", false,
		"preserve any tags on the environment being cloned")

	cmd.Flags().BoolVar(&preserveRevisionTags,
		"revTags", false,
		"preserve any tags on the environment revisions being cloned")

	return cmd
}

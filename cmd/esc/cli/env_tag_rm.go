// Copyright 2024, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pulumi/esc/cmd/esc/cli/client"
)

func newEnvTagRmCmd(env *envCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm [<org-name>/]<environment-name> <tag-name>",
		Short: "Remove an environment tag.",
		Long: "Remove an environment tag\n" +
			"\n" +
			"This command removes an environment tag using the tag name.\n",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			ref, _, err := env.getEnvRef(args)
			if err != nil {
				return err
			}
			if ref.version != "" {
				return fmt.Errorf("the ls command does not accept versions")
			}

			tagIdentifier := args[1]

			var tagID string
			after := ""
			count := 500
			for {
				options := client.ListEnvironmentTagsOptions{
					After: after,
					Count: &count,
				}
				tags, next, err := env.esc.client.ListEnvironmentTags(ctx, ref.orgName, ref.envName, options)
				if err != nil {
					return err
				}

				after = next
				for _, t := range tags {
					if t.Name == tagIdentifier {
						tagID = t.ID
						break
					}
				}

				if after == "0" {
					break
				}
			}

			if tagID == "" {
				return fmt.Errorf("could not find tag with name %q on environment %q", tagIdentifier, ref.envName)
			}

			err = env.esc.client.DeleteEnvironmentTag(ctx, ref.orgName, ref.envName, tagID)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully deleted environment tag: %v\n", tagIdentifier)
			return nil
		},
	}

	return cmd
}

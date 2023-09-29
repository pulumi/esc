// Copyright 2023, Pulumi Corporation.

package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/pulumi/esc/cmd/esc/internal/client"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
)

func newEnvLsCmd(env *envCommand) *cobra.Command {
	var orgFilter string

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List environments.",
		Long: "List environments\n" +
			"\n" +
			"This command lists environments. All environments you have access to will be listed.\n",
		Args: cmdutil.NoArgs,
		Run: cmdutil.RunFunc(func(cmd *cobra.Command, _ []string) error {
			ctx := context.Background()

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			user := env.esc.account.Username

			continuationToken, allNames := "", []client.OrgEnvironment(nil)
			for {
				names, nextToken, err := env.esc.client.ListEnvironments(ctx, orgFilter, continuationToken)
				if err != nil {
					return fmt.Errorf("listing environments: %w", err)
				}
				for _, name := range names {
					if name.Organization == user {
						name.Organization = ""
					}
					allNames = append(allNames, name)
				}
				if nextToken == "" {
					break
				}
				continuationToken = nextToken
			}

			sort.Slice(allNames, func(i, j int) bool {
				return allNames[i].Organization < allNames[j].Organization || allNames[i].Name < allNames[j].Name
			})

			for _, n := range allNames {
				if n.Organization == "" {
					fmt.Fprintln(env.esc.stdout, n.Name)
				} else {
					fmt.Fprintf(env.esc.stdout, "%v/%v\n", n.Organization, n.Name)
				}
			}

			return nil
		}),
	}

	cmd.PersistentFlags().StringVarP(
		&orgFilter, "organization", "o", "", "Filter returned stacks to those in a specific organization")

	return cmd
}

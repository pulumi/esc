// Copyright 2023, Pulumi Corporation.

package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

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

			user, orgs := env.esc.account.Username, env.esc.account.Organizations
			if orgFilter != "" {
				orgs = []string{orgFilter}
			}

			if len(orgs) > 1 {
				// Swap the single-user org to the front of the list
				for i, org := range orgs {
					if org == user {
						orgs[0], orgs[i] = user, orgs[0]
						break
					}
				}
				// Sort the rest of the orgs lexicographically
				sort.Strings(orgs[1:])
			}

			for _, org := range orgs {
				names, err := env.esc.client.ListEnvironments(ctx, org)
				if err != nil {
					return fmt.Errorf("listing environments: %w", err)
				}
				for _, name := range names {
					if org != user {
						fmt.Printf("%v/%v\n", org, name)
					} else {
						fmt.Println(name)
					}
				}
			}
			return nil
		}),
	}

	cmd.PersistentFlags().StringVarP(
		&orgFilter, "organization", "o", "", "Filter returned stacks to those in a specific organization")

	return cmd
}

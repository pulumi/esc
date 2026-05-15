// Copyright 2023, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"

	"github.com/pulumi/esc/cmd/esc/cli/client"
)

func newEnvLsCmd(env *envCommand) *cobra.Command {
	var (
		orgFilter     string
		projectFilter string
	)

	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List environments.",
		Long: "List environments\n" +
			"\n" +
			"This command lists environments. All environments you have access to will be listed.\n",
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := context.Background()

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			allEnvs, err := env.listEnvironments(ctx, orgFilter, projectFilter)
			if err != nil {
				return err
			}

			sort.Slice(allEnvs, func(i, j int) bool {
				ei, ej := allEnvs[i], allEnvs[j]

				if ei.Organization == ej.Organization {
					if ei.Project == ej.Project {
						return ei.Name < ej.Name
					}
					return ei.Project < ej.Project
				}
				return ei.Organization < ej.Organization
			})

			if len(allEnvs) == 0 {
				return nil
			}

			rows := make([]cmdutil.TableRow, 0, len(allEnvs))
			for _, e := range allEnvs {
				rows = append(rows, cmdutil.TableRow{
					Columns: []string{e.Organization, e.Project, e.Name},
				})
			}
			return cmdutil.FprintTable(env.esc.stdout, cmdutil.Table{
				Headers: []string{"ORGANIZATION", "PROJECT", "NAME"},
				Rows:    rows,
			})
		},
	}

	cmd.PersistentFlags().StringVarP(
		&orgFilter, "organization", "o", "", "Filter returned environments to those in a specific organization")
	cmd.PersistentFlags().StringVarP(
		&projectFilter, "project", "p", "", "Filter returned environments to those in a specific project")

	return cmd
}

func (env *envCommand) listEnvironments(ctx context.Context, orgFilter, projectFilter string) ([]client.OrgEnvironment, error) {
	user := env.esc.account.Username
	continuationToken, allEnvs := "", []client.OrgEnvironment(nil)
	for {
		var envs []client.OrgEnvironment
		var nextToken string
		var err error

		// If orgFilter is specified, use ListOrganizationEnvironments endpoint, so that we receive proper errors
		// like 404 when environment doesn't exist, instead of an empty array
		if orgFilter != "" {
			envs, nextToken, err = env.esc.client.ListOrganizationEnvironments(ctx, orgFilter, continuationToken)
		} else {
			envs, nextToken, err = env.esc.client.ListEnvironments(ctx, continuationToken)
		}

		if err != nil {
			return []client.OrgEnvironment(nil), fmt.Errorf("listing environments: %w", err)
		}
		for _, e := range envs {
			if e.Organization == user {
				e.Organization = ""
			}
			if projectFilter != "" && e.Project != projectFilter {
				continue
			}
			allEnvs = append(allEnvs, e)
		}
		if nextToken == "" {
			break
		}
		continuationToken = nextToken
	}

	return allEnvs, nil
}

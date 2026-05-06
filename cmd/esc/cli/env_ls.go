// Copyright 2023, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

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

			desc, err := env.inferDefaultEnv(ctx)
			if err != nil {
				return fmt.Errorf("configuring default environment: %w", err)
			}
			ref, hasDefaultRef := desc.(environmentRef)

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			// listEnvironments empties Organization for envs in the calling user's namespace, so
			// normalize the inferred ref the same way before comparing.
			if hasDefaultRef && ref.orgName == env.esc.account.Username {
				ref.orgName = ""
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

			for _, e := range allEnvs {
				isDefault := hasDefaultRef && ref.orgName == e.Organization && ref.projectName == e.Project && ref.envName == e.Name
				defaultSuffix := ""
				if isDefault {
					defaultSuffix = " (default)"
				}

				if e.Organization == "" {
					fmt.Fprintf(env.esc.stdout, "%v/%v%v\n", e.Project, e.Name, defaultSuffix)
				} else {
					fmt.Fprintf(env.esc.stdout, "%v/%v/%v%v\n", e.Organization, e.Project, e.Name, defaultSuffix)
				}
			}

			return nil
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

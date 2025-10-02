// Copyright 2025, Pulumi Corporation.

package cli

import (
	"context"

	"github.com/spf13/cobra"
)

func newEnvOpenRequestCmd(envcmd *envCommand) *cobra.Command {
	var grantExpirationSeconds int
	var accessDurationSeconds int

	cmd := &cobra.Command{
		Use:   "open-request [<org-name>/][<project-name>/]<environment-name>[@<version>]",
		Args:  cobra.ExactArgs(1),
		Short: "Create a request for opening a protected environment.",
		Long: "Create a request for opening a protected environment with the given name.\n" +
			"\n" +
			"This command creates a request to open a protected environment. The request must be\n" +
			"approved before the environment can be accessed.\n",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if err := envcmd.esc.getCachedClient(ctx); err != nil {
				return err
			}

			ref, _, err := envcmd.getExistingEnvRef(ctx, args)
			if err != nil {
				return err
			}

			return envcmd.esc.client.CreateEnvironmentOpenRequest(
				ctx,
				ref.orgName,
				ref.projectName,
				ref.envName,
				grantExpirationSeconds,
				accessDurationSeconds,
			)
		},
	}

	cmd.Flags().IntVar(
		&grantExpirationSeconds, "grant-expiration-seconds", 90000,
		"expiration time for the grant in seconds (default: 90000)")
	cmd.Flags().IntVar(
		&accessDurationSeconds, "access-duration-seconds", 259200,
		"duration of access in seconds (default: 259200)")

	return cmd
}

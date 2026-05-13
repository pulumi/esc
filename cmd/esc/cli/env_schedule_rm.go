// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func newEnvScheduleRmCmd(env *envCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm [<org-name>/][<project-name>/]<environment-name> <schedule-id>",
		Short: "Remove an environment scheduled action.",
		Long: "Remove an environment scheduled action\n" +
			"\n" +
			"This command removes the named scheduled action from the environment.\n",
		Args:         cobra.ExactArgs(2),
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
				return errors.New("the rm command does not accept versions")
			}

			scheduleID := args[0]
			if scheduleID == "" {
				return errors.New("schedule ID cannot be empty")
			}

			if err := env.esc.client.DeleteEnvironmentSchedule(ctx, ref.orgName, ref.projectName, ref.envName, scheduleID); err != nil {
				return err
			}

			fmt.Fprintf(env.esc.stdout, "Removed schedule %s from %s/%s/%s\n",
				scheduleID, ref.orgName, ref.projectName, ref.envName)
			return nil
		},
	}

	return cmd
}

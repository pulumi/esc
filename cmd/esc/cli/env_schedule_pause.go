// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func newEnvSchedulePauseCmd(env *envCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause [<org-name>/][<project-name>/]<environment-name> <schedule-id>",
		Short: "Pause an environment scheduled action.",
		Long: "Pause an environment scheduled action\n" +
			"\n" +
			"This command pauses the named scheduled action. The schedule remains attached to\n" +
			"the environment and can be resumed later with `esc env schedule resume`.\n",
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
				return errors.New("the pause command does not accept versions")
			}

			scheduleID := args[0]
			if scheduleID == "" {
				return errors.New("schedule ID cannot be empty")
			}

			if err := env.esc.client.PauseEnvironmentSchedule(ctx, ref.orgName, ref.projectName, ref.envName, scheduleID); err != nil {
				return err
			}

			fmt.Fprintf(env.esc.stdout, "Paused schedule %s for %s/%s/%s\n",
				scheduleID, ref.orgName, ref.projectName, ref.envName)
			return nil
		},
	}

	return cmd
}

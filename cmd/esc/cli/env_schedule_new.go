// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pulumi/esc/cmd/esc/cli/client"
)

func newEnvScheduleNewCmd(env *envCommand) *cobra.Command {
	var (
		cron string
		once string
		path string
	)

	cmd := &cobra.Command{
		Use:   "new [<org-name>/][<project-name>/]<environment-name>",
		Short: "Create a new scheduled action on an environment.",
		Long: "Create a new scheduled action on an environment\n" +
			"\n" +
			"This command schedules a secret rotation against the environment. Use --cron to\n" +
			"schedule a recurring rotation or --once to schedule a single rotation at a\n" +
			"specific time (ISO 8601 / RFC 3339). Use --path to rotate only a subset of the\n" +
			"environment; omit it to rotate the whole environment.\n",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			ref, _, err := env.getExistingEnvRef(ctx, args)
			if err != nil {
				return err
			}
			if ref.version != "" {
				return errors.New("the new command does not accept versions")
			}

			switch {
			case cron == "" && once == "":
				return errors.New("exactly one of --cron or --once must be set")
			case cron != "" && once != "":
				return errors.New("only one of --cron or --once may be set")
			}

			req := client.CreateEnvironmentScheduleRequest{
				ScheduleCron: cron,
				ScheduleOnce: once,
				SecretRotationRequest: &client.CreateEnvironmentSecretRotationScheduleRequest{
					EnvironmentPath: path,
				},
			}

			s, err := env.esc.client.CreateEnvironmentSchedule(ctx, ref.orgName, ref.projectName, ref.envName, req)
			if err != nil {
				return err
			}

			fmt.Fprintf(env.esc.stdout, "Created schedule %s for %s/%s/%s\n",
				s.ID, ref.orgName, ref.projectName, ref.envName)
			return nil
		},
	}

	cmd.Flags().StringVar(&cron, "cron", "", "a cron expression for a recurring schedule")
	cmd.Flags().StringVar(&once, "once", "", "an ISO 8601 timestamp for a one-time schedule")
	cmd.Flags().StringVar(&path, "path", "", "the path within the environment to rotate (default: rotate the whole environment)")

	return cmd
}

// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/pulumi/esc/cmd/esc/cli/client"
)

func newEnvScheduleLsCmd(env *envCommand) *cobra.Command {
	var utc bool

	cmd := &cobra.Command{
		Use:   "ls [<org-name>/][<project-name>/]<environment-name>",
		Short: "List environment scheduled actions.",
		Long: "List environment scheduled actions\n" +
			"\n" +
			"This command lists the scheduled actions configured for the given environment.\n",
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
				return fmt.Errorf("the ls command does not accept versions")
			}

			resp, err := env.esc.client.ListEnvironmentSchedules(ctx, ref.orgName, ref.projectName, ref.envName)
			if err != nil {
				return err
			}

			printSchedules(env.esc.stdout, resp, utcFlag(utc))
			return nil
		},
	}

	cmd.Flags().BoolVar(&utc, "utc", false, "display times in UTC")

	return cmd
}

// printSchedules writes each schedule as a key/value block separated by blank lines.
func printSchedules(stdout io.Writer, resp *client.ListScheduledActionsResponse, utc utcFlag) {
	if resp == nil {
		return
	}
	for i, s := range resp.Schedules {
		if i > 0 {
			fmt.Fprintln(stdout)
		}
		printSchedule(stdout, s, utc)
	}
}

func printSchedule(stdout io.Writer, s client.ScheduledAction, utc utcFlag) {
	fmt.Fprintf(stdout, "ID: %s\n", s.ID)
	fmt.Fprintf(stdout, "Kind: %s\n", s.Kind)
	schedule := s.ScheduleCron
	if schedule == "" {
		schedule = s.ScheduleOnce
	}
	if schedule == "" {
		schedule = "<unknown>"
	}
	fmt.Fprintf(stdout, "Schedule: %s\n", schedule)
	fmt.Fprintf(stdout, "Paused: %t\n", s.Paused)
	fmt.Fprintf(stdout, "Next execution: %s\n", formatScheduleTime(s.NextExecution, utc))
	fmt.Fprintf(stdout, "Last executed: %s\n", formatScheduleTime(s.LastExecuted, utc))
}

// formatScheduleTime parses an ISO 8601 timestamp and re-formats it honouring the --utc flag.
// Empty or unparseable values pass through as "never" and the raw string, respectively.
func formatScheduleTime(s string, utc utcFlag) string {
	if s == "" {
		return "never"
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return utc.time(t).Format(time.RFC3339)
}

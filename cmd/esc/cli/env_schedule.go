// Copyright 2026, Pulumi Corporation.

package cli

import (
	"github.com/spf13/cobra"
)

func newEnvScheduleCmd(env *envCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schedule",
		Short: "Manage environment scheduled actions",
		Long: "Manage environment scheduled actions\n" +
			"\n" +
			"A scheduled action runs against an environment on a cron schedule or at a single\n" +
			"point in time. Today the CLI exposes secret-rotation schedules.\n" +
			"\n" +
			"Subcommands exist for listing, creating, pausing, resuming, and removing schedules.",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}

	cmd.AddCommand(newEnvScheduleLsCmd(env))
	cmd.AddCommand(newEnvScheduleNewCmd(env))
	cmd.AddCommand(newEnvSchedulePauseCmd(env))
	cmd.AddCommand(newEnvScheduleResumeCmd(env))
	cmd.AddCommand(newEnvScheduleRmCmd(env))

	return cmd
}

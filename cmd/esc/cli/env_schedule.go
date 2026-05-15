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
			"Subcommands exist for listing, creating, inspecting, editing, and removing\n" +
			"schedules, as well as viewing their execution history.",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}

	cmd.AddCommand(newEnvScheduleEditCmd(env))
	cmd.AddCommand(newEnvScheduleGetCmd(env))
	cmd.AddCommand(newEnvScheduleHistoryCmd(env))
	cmd.AddCommand(newEnvScheduleListCmd(env))
	cmd.AddCommand(newEnvScheduleNewCmd(env))
	cmd.AddCommand(newEnvScheduleRemoveCmd(env))

	return cmd
}

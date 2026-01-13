// Copyright 2026, Pulumi Corporation.

package cli

import (
	"github.com/spf13/cobra"
)

type setupCommand struct {
	esc *escCommand
}

func newSetupCmd(esc *escCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup cloud integrations for Pulumi ESC",
		Long: "Setup cloud integrations for Pulumi ESC\n" +
			"\n" +
			"This command group provides utilities to configure cloud providers for use with " +
			"Pulumi ESC OIDC authentication.\n" +
			"\n" +
			"Available cloud providers:\n" +
			"  aws     Setup AWS OIDC integration\n",

		Args: cobra.NoArgs,
	}

	setup := &setupCommand{esc: esc}

	cmd.AddCommand(newSetupAWSCmd(setup))

	return cmd
}

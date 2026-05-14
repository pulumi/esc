// Copyright 2026, Pulumi Corporation.

package cli

import (
	"github.com/spf13/cobra"
)

func newEnvProviderCmd(env *envCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage login providers within an environment",
		Long: "Manage login providers within an environment\n" +
			"\n" +
			"Subcommands add cloud-provider login blocks (AWS, Azure, GCP) to an environment.\n" +
			"Only static credentials are supported today; OIDC support will be added later.\n",
		Args: cobra.NoArgs,
	}

	cmd.AddCommand(newEnvProviderAWSLoginCmd(env))
	cmd.AddCommand(newEnvProviderAzureLoginCmd(env))

	return cmd
}

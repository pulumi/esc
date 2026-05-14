// Copyright 2026, Pulumi Corporation.

package cli

import (
	"github.com/spf13/cobra"
)

func newEnvProviderAWSLoginCmd(env *envCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aws-login",
		Short: "Add an AWS login provider to an environment",
		Long: "Add an AWS login provider to an environment\n" +
			"\n" +
			"Subcommands select the authentication mode. Today only `static` is supported;\n" +
			"`oidc` is planned in a follow-up.\n",
		Args: cobra.NoArgs,
	}

	cmd.AddCommand(newEnvProviderAWSLoginStaticCmd(env))

	return cmd
}

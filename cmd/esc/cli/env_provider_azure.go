// Copyright 2026, Pulumi Corporation.

package cli

import (
	"github.com/spf13/cobra"
)

func newEnvProviderAzureLoginCmd(env *envCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azure-login",
		Short: "Add an Azure login provider to an environment",
		Long: "Add an Azure login provider to an environment\n" +
			"\n" +
			"Subcommands select the authentication mode. Today only `static` is supported;\n" +
			"`oidc` is planned in a follow-up.\n",
		Args: cobra.NoArgs,
	}

	cmd.AddCommand(newEnvProviderAzureLoginStaticCmd(env))

	return cmd
}

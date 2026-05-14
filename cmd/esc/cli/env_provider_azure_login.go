// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

func newEnvProviderAzureLoginCmd(env *envCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "azure-login",
		Short: "Add an Azure login provider to an environment",
		Long: "Add an Azure login provider to an environment\n" +
			"\n" +
			"Subcommands select the authentication mode: `static` for static credentials,\n" +
			"`oidc` for federated identity via OpenID Connect.\n",
		Args: cobra.NoArgs,
	}

	cmd.AddCommand(newEnvProviderAzureLoginStaticCmd(env))
	cmd.AddCommand(newEnvProviderAzureLoginOIDCCmd(env))

	return cmd
}

func newEnvProviderAzureLoginStaticCmd(env *envCommand) *cobra.Command {
	var clientSecret string
	var pathStr string
	var draft string
	var create bool

	cmd := &cobra.Command{
		Use:   "static [<org>/][<project>/]<environment-name> <client-id> <tenant-id> <subscription-id>",
		Args:  cobra.RangeArgs(3, 4),
		Short: "Add an Azure static-credentials login provider to an environment",
		Long: "Add an Azure static-credentials login provider to an environment\n" +
			"\n" +
			"Writes an `fn::open::azure-login` block at the configured path under `values`.\n" +
			"`--client-secret`, if provided, is wrapped in `fn::secret`. If a block already\n" +
			"exists at the path it is replaced.\n",
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
				return fmt.Errorf("the provider command does not accept versions")
			}
			if len(args) != 3 {
				return fmt.Errorf("expected <client-id> <tenant-id> <subscription-id>")
			}
			clientID, tenantID, subscriptionID := args[0], args[1], args[2]

			path, err := resource.ParsePropertyPath(pathStr)
			if err != nil {
				return fmt.Errorf("invalid --path: %w", err)
			}

			node := buildAzureLoginStaticNode(clientID, tenantID, subscriptionID, clientSecret)

			if err := ensureProviderEnv(ctx, env, ref, create); err != nil {
				return err
			}
			return applyProviderUpdate(ctx, env, ref, draft, path, node)
		},
	}

	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "optional Azure client secret")
	cmd.Flags().StringVar(&pathStr, "path", "azure.login", "property path under `values` where the provider block is written")
	cmd.Flags().BoolVar(&create, "create", false,
		"create the environment if it does not already exist")
	cmd.Flags().StringVar(&draft, "draft", "",
		"set flag without a value (--draft) to create a draft rather than saving changes directly. --draft=<change-request-id> to update an existing change request.")
	cmd.Flag("draft").NoOptDefVal = "new"

	return cmd
}

// buildAzureLoginStaticNode returns a yaml.Node representing
// `fn::open::azure-login: { ... }`. clientSecret is wrapped in `fn::secret` and
// omitted when empty.
func buildAzureLoginStaticNode(clientID, tenantID, subscriptionID, clientSecret string) *yaml.Node {
	loginContent := []*yaml.Node{
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "clientId"},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: clientID},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "tenantId"},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: tenantID},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "subscriptionId"},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: subscriptionID},
	}
	if clientSecret != "" {
		loginContent = append(loginContent,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "clientSecret"},
			secretNode(clientSecret),
		)
	}

	return &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "fn::open::azure-login"},
			{Kind: yaml.MappingNode, Tag: "!!map", Content: loginContent},
		},
	}
}

func newEnvProviderAzureLoginOIDCCmd(env *envCommand) *cobra.Command {
	var clientSecret string
	var subjectAttributes []string
	var pathStr string
	var draft string
	var create bool

	cmd := &cobra.Command{
		Use:   "oidc [<org>/][<project>/]<environment-name> <client-id> <tenant-id> <subscription-id>",
		Args:  cobra.RangeArgs(3, 4),
		Short: "Add an Azure OIDC login provider to an environment",
		Long: "Add an Azure OIDC login provider to an environment\n" +
			"\n" +
			"Writes an `fn::open::azure-login` block with `oidc: true` at the configured\n" +
			"path under `values`. `--client-secret`, if provided, is wrapped in\n" +
			"`fn::secret`. The Azure federated credential must be provisioned separately\n" +
			"(e.g. with Pulumi). If a block already exists at the path it is replaced.\n",
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
				return fmt.Errorf("the provider command does not accept versions")
			}
			if len(args) != 3 {
				return fmt.Errorf("expected <client-id> <tenant-id> <subscription-id>")
			}
			clientID, tenantID, subscriptionID := args[0], args[1], args[2]

			path, err := resource.ParsePropertyPath(pathStr)
			if err != nil {
				return fmt.Errorf("invalid --path: %w", err)
			}

			node := buildAzureLoginOIDCNode(clientID, tenantID, subscriptionID, clientSecret, subjectAttributes)

			if err := ensureProviderEnv(ctx, env, ref, create); err != nil {
				return err
			}
			return applyProviderUpdate(ctx, env, ref, draft, path, node)
		},
	}

	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "optional Azure client secret")
	cmd.Flags().StringArrayVar(&subjectAttributes, "subject-attribute", nil,
		"OIDC subject attribute to include in the federated token (repeatable)")
	cmd.Flags().StringVar(&pathStr, "path", "azure.login", "property path under `values` where the provider block is written")
	cmd.Flags().BoolVar(&create, "create", false,
		"create the environment if it does not already exist")
	cmd.Flags().StringVar(&draft, "draft", "",
		"set flag without a value (--draft) to create a draft rather than saving changes directly. --draft=<change-request-id> to update an existing change request.")
	cmd.Flag("draft").NoOptDefVal = "new"

	return cmd
}

// buildAzureLoginOIDCNode returns a yaml.Node representing
// `fn::open::azure-login: { ..., oidc: true }`. clientSecret is wrapped in
// `fn::secret` and omitted when empty. subjectAttributes is omitted when empty.
func buildAzureLoginOIDCNode(clientID, tenantID, subscriptionID, clientSecret string, subjectAttributes []string) *yaml.Node {
	loginContent := []*yaml.Node{
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "clientId"},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: clientID},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "tenantId"},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: tenantID},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "subscriptionId"},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: subscriptionID},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "oidc"},
		{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"},
	}
	if clientSecret != "" {
		loginContent = append(loginContent,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "clientSecret"},
			secretNode(clientSecret),
		)
	}
	if len(subjectAttributes) > 0 {
		loginContent = append(loginContent,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "subjectAttributes"},
			stringSequenceNode(subjectAttributes),
		)
	}

	return &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "fn::open::azure-login"},
			{Kind: yaml.MappingNode, Tag: "!!map", Content: loginContent},
		},
	}
}

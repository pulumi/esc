// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
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

func newEnvProviderAWSLoginStaticCmd(env *envCommand) *cobra.Command {
	var sessionToken string
	var pathStr string
	var draft string

	cmd := &cobra.Command{
		Use:   "static [<org>/][<project>/]<environment-name> <access-key-id> <secret-access-key>",
		Args:  cobra.RangeArgs(2, 3),
		Short: "Add an AWS static-credentials login provider to an environment",
		Long: "Add an AWS static-credentials login provider to an environment\n" +
			"\n" +
			"Writes an `fn::open::aws-login` block with static credentials at the configured\n" +
			"path under `values`. The secret access key and session token, if any, are\n" +
			"wrapped in `fn::secret`. If a block already exists at the path it is replaced.\n",
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
			if len(args) != 2 {
				return fmt.Errorf("expected <access-key-id> and <secret-access-key>")
			}
			accessKeyID, secretAccessKey := args[0], args[1]

			path, err := resource.ParsePropertyPath(pathStr)
			if err != nil {
				return fmt.Errorf("invalid --path: %w", err)
			}

			node := buildAWSLoginStaticNode(accessKeyID, secretAccessKey, sessionToken)

			return applyProviderUpdate(ctx, env, ref, draft, path, node)
		},
	}

	cmd.Flags().StringVar(&sessionToken, "session-token", "", "optional AWS session token")
	cmd.Flags().StringVar(&pathStr, "path", "aws.login", "property path under `values` where the provider block is written")
	cmd.Flags().StringVar(&draft, "draft", "",
		"set flag without a value (--draft) to create a draft rather than saving changes directly. --draft=<change-request-id> to update an existing change request.")
	cmd.Flag("draft").NoOptDefVal = "new"

	return cmd
}

// buildAWSLoginStaticNode returns a yaml.Node representing
// `fn::open::aws-login: { static: {...} }`. secretAccessKey and sessionToken
// are wrapped in `fn::secret`. sessionToken is omitted when empty.
func buildAWSLoginStaticNode(accessKeyID, secretAccessKey, sessionToken string) *yaml.Node {
	staticContent := []*yaml.Node{
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "accessKeyId"},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: accessKeyID},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "secretAccessKey"},
		secretNode(secretAccessKey),
	}
	if sessionToken != "" {
		staticContent = append(staticContent,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "sessionToken"},
			secretNode(sessionToken),
		)
	}

	return &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "fn::open::aws-login"},
			{
				Kind: yaml.MappingNode,
				Tag:  "!!map",
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Tag: "!!str", Value: "static"},
					{Kind: yaml.MappingNode, Tag: "!!map", Content: staticContent},
				},
			},
		},
	}
}

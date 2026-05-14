// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
)

// buildGCPLoginNode returns a yaml.Node representing
// `fn::open::gcp-login: { project, accessToken: { accessToken: {fn::secret}, ... } }`.
// serviceAccount and tokenLifetime are omitted when empty.
func buildGCPLoginNode(project int64, accessToken, serviceAccount, tokenLifetime string) *yaml.Node {
	accessTokenContent := []*yaml.Node{
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "accessToken"},
		secretNode(accessToken),
	}
	if serviceAccount != "" {
		accessTokenContent = append(accessTokenContent,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "serviceAccount"},
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: serviceAccount},
		)
	}
	if tokenLifetime != "" {
		accessTokenContent = append(accessTokenContent,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "tokenLifetime"},
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: tokenLifetime},
		)
	}

	return &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
		Content: []*yaml.Node{
			{Kind: yaml.ScalarNode, Tag: "!!str", Value: "fn::open::gcp-login"},
			{
				Kind: yaml.MappingNode,
				Tag:  "!!map",
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Tag: "!!str", Value: "project"},
					{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.FormatInt(project, 10)},
					{Kind: yaml.ScalarNode, Tag: "!!str", Value: "accessToken"},
					{Kind: yaml.MappingNode, Tag: "!!map", Content: accessTokenContent},
				},
			},
		},
	}
}

func newEnvProviderGCPLoginCmd(env *envCommand) *cobra.Command {
	var serviceAccount string
	var tokenLifetime string
	var pathStr string
	var draft string

	cmd := &cobra.Command{
		Use:   "gcp-login [<org>/][<project>/]<environment-name> <project-number> <access-token>",
		Args:  cobra.RangeArgs(2, 3),
		Short: "Add a GCP static-credentials login provider to an environment",
		Long: "Add a GCP static-credentials login provider to an environment\n" +
			"\n" +
			"Writes an `fn::open::gcp-login` block at the configured path under `values`. The\n" +
			"access token is wrapped in `fn::secret`. <project-number> must be the numerical\n" +
			"GCP project ID. If a block already exists at the path it is replaced.\n",
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
				return fmt.Errorf("expected <project-number> and <access-token>")
			}
			project, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid project number %q: must be a positive integer", args[0])
			}
			if project <= 0 {
				return fmt.Errorf("invalid project number %q: must be a positive integer", args[0])
			}
			accessToken := args[1]

			path, err := resource.ParsePropertyPath(pathStr)
			if err != nil {
				return fmt.Errorf("invalid --path: %w", err)
			}

			node := buildGCPLoginNode(project, accessToken, serviceAccount, tokenLifetime)

			return applyProviderUpdate(ctx, env, ref, draft, path, node)
		},
	}

	cmd.Flags().StringVar(&serviceAccount, "service-account", "", "optional GCP service account to impersonate")
	cmd.Flags().StringVar(&tokenLifetime, "token-lifetime", "", "optional lifetime for impersonated credentials, e.g. 1h30m")
	cmd.Flags().StringVar(&pathStr, "path", "gcp.login", "property path under `values` where the provider block is written")
	cmd.Flags().StringVar(&draft, "draft", "",
		"set flag without a value (--draft) to create a draft rather than saving changes directly. --draft=<change-request-id> to update an existing change request.")
	cmd.Flag("draft").NoOptDefVal = "new"

	return cmd
}

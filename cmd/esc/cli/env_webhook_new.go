// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pulumi/esc/cmd/esc/cli/client"
)

func newEnvWebhookNewCmd(env *envCommand) *cobra.Command {
	var (
		url         string
		displayName string
		format      string
		filters     []string
		active      bool
		secret      string
	)

	cmd := &cobra.Command{
		Use:   "new [<org-name>/][<project-name>/]<environment-name> <webhook-name>",
		Short: "Create a new environment webhook.",
		Long: "Create a new environment webhook\n" +
			"\n" +
			"This command attaches a new webhook to the given environment. The webhook will be\n" +
			"delivered to --url whenever the environment changes. Use --filter to limit the set\n" +
			"of events that trigger a delivery; the flag can be repeated.\n",
		Args:         cobra.ExactArgs(2),
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
				return errors.New("the new command does not accept versions")
			}

			webhookName := args[0]
			if webhookName == "" {
				return errors.New("webhook name cannot be empty")
			}
			if url == "" {
				return errors.New("--url is required")
			}

			if displayName == "" {
				displayName = webhookName
			}

			req := client.CreateEnvironmentWebhookRequest{
				Active:           active,
				DisplayName:      displayName,
				Name:             webhookName,
				OrganizationName: ref.orgName,
				PayloadURL:       url,
				Filters:          filters,
				Format:           format,
				Secret:           secret,
			}

			w, err := env.esc.client.CreateEnvironmentWebhook(ctx, ref.orgName, ref.projectName, ref.envName, req)
			if err != nil {
				return err
			}

			fmt.Fprintf(env.esc.stdout, "Created webhook %s for %s/%s/%s\n",
				w.Name, ref.orgName, ref.projectName, ref.envName)
			return nil
		},
	}

	cmd.Flags().StringVar(&url, "url", "", "the payload URL to deliver events to (required)")
	cmd.Flags().StringVar(&displayName, "display-name", "", "the display name (defaults to the webhook name)")
	cmd.Flags().StringVar(&format, "format", "raw", "the payload format")
	cmd.Flags().StringArrayVar(&filters, "filter", nil, "event filters to apply (repeatable)")
	cmd.Flags().BoolVar(&active, "active", true, "whether the webhook is active")
	cmd.Flags().StringVar(&secret, "secret", "", "shared secret used to sign deliveries")

	return cmd
}

// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/pulumi/esc/cmd/esc/cli/client"
)

func newEnvWebhookEditCmd(env *envCommand) *cobra.Command {
	var (
		url           string
		displayName   string
		format        string
		filters       []string
		active        bool
		secret        string
		addFilters    []string
		removeFilters []string
	)

	cmd := &cobra.Command{
		Use:   "edit [<org-name>/][<project-name>/]<environment-name> <webhook-name>",
		Short: "Edit an environment webhook.",
		Long: "Edit an environment webhook\n" +
			"\n" +
			"This command updates one or more fields of the named webhook. Only the fields\n" +
			"specified via flags are sent; everything else is left unchanged.\n" +
			"\n" +
			"--filter replaces the filter list. Use --add-filter and --remove-filter to apply\n" +
			"incremental changes that merge with the existing filters; mixing --filter with\n" +
			"either of those is not allowed.\n",
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
				return errors.New("the edit command does not accept versions")
			}

			webhookName := args[0]
			if webhookName == "" {
				return errors.New("webhook name cannot be empty")
			}

			filterChanged := cmd.Flags().Changed("filter")
			addFilterChanged := cmd.Flags().Changed("add-filter")
			removeFilterChanged := cmd.Flags().Changed("remove-filter")
			if filterChanged && (addFilterChanged || removeFilterChanged) {
				return errors.New("--filter cannot be combined with --add-filter or --remove-filter")
			}

			req := client.UpdateEnvironmentWebhookRequest{}
			if cmd.Flags().Changed("active") {
				v := active
				req.Active = &v
			}
			if cmd.Flags().Changed("display-name") {
				v := displayName
				req.DisplayName = &v
			}
			if cmd.Flags().Changed("url") {
				v := url
				req.PayloadURL = &v
			}
			if cmd.Flags().Changed("format") {
				v := format
				req.Format = &v
			}
			if cmd.Flags().Changed("secret") {
				v := secret
				req.Secret = &v
			}
			if filterChanged {
				v := append([]string(nil), filters...)
				req.Filters = &v
			} else if addFilterChanged || removeFilterChanged {
				existing, err := env.esc.client.GetEnvironmentWebhook(
					ctx, ref.orgName, ref.projectName, ref.envName, webhookName)
				if err != nil {
					return err
				}
				merged := mergeFilters(existing.Filters, addFilters, removeFilters)
				req.Filters = &merged
			}

			w, err := env.esc.client.UpdateEnvironmentWebhook(
				ctx, ref.orgName, ref.projectName, ref.envName, webhookName, req)
			if err != nil {
				return err
			}

			fmt.Fprintf(env.esc.stdout, "Updated webhook %s for %s/%s/%s\n",
				w.Name, ref.orgName, ref.projectName, ref.envName)
			return nil
		},
	}

	cmd.Flags().StringVar(&url, "url", "", "the payload URL to deliver events to")
	cmd.Flags().StringVar(&displayName, "display-name", "", "the display name")
	cmd.Flags().StringVar(&format, "format", "", "the payload format")
	cmd.Flags().StringArrayVar(&filters, "filter", nil, "replace the event filters (repeatable)")
	cmd.Flags().BoolVar(&active, "active", true, "whether the webhook is active")
	cmd.Flags().StringVar(&secret, "secret", "", "shared secret used to sign deliveries")
	cmd.Flags().StringArrayVar(&addFilters, "add-filter", nil, "add an event filter (repeatable)")
	cmd.Flags().StringArrayVar(&removeFilters, "remove-filter", nil, "remove an event filter (repeatable)")

	return cmd
}

// mergeFilters returns existing minus removes, then appends adds (skipping duplicates).
func mergeFilters(existing, adds, removes []string) []string {
	removeSet := map[string]struct{}{}
	for _, r := range removes {
		removeSet[r] = struct{}{}
	}
	out := make([]string, 0, len(existing)+len(adds))
	present := map[string]struct{}{}
	for _, f := range existing {
		if _, drop := removeSet[f]; drop {
			continue
		}
		if _, seen := present[f]; seen {
			continue
		}
		present[f] = struct{}{}
		out = append(out, f)
	}
	for _, a := range adds {
		if _, drop := removeSet[a]; drop {
			continue
		}
		if _, seen := present[a]; seen {
			continue
		}
		present[a] = struct{}{}
		out = append(out, a)
	}
	return out
}

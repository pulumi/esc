// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/pulumi/esc/cmd/esc/cli/client"
)

func newEnvWebhookListCmd(env *envCommand) *cobra.Command {
	var count int

	cmd := &cobra.Command{
		Use:     "list [<org-name>/][<project-name>/]<environment-name>",
		Aliases: []string{"ls"},
		Short:   "List environment webhooks.",
		Long: "[EXPERIMENTAL] List environment webhooks\n" +
			"\n" +
			"This command lists the webhooks attached to the given environment.\n",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			ref, _, err := env.getExistingEnvRef(ctx, args)
			if err != nil {
				return err
			}
			if ref.version != "" {
				return fmt.Errorf("the list command does not accept versions")
			}
			if count < 0 {
				return fmt.Errorf("--count must be non-negative")
			}

			hooks, err := env.esc.client.ListEnvironmentWebhooks(ctx, ref.orgName, ref.projectName, ref.envName)
			if err != nil {
				return err
			}

			if count > 0 && len(hooks) > count {
				hooks = hooks[:count]
			}

			printWebhooks(env.esc.stdout, hooks)
			return nil
		},
	}

	cmd.Flags().IntVar(&count, "count", 0, "the maximum number of webhooks to return (all if unset)")

	return cmd
}

// printWebhooks renders the webhook list as a tab-aligned table.
func printWebhooks(stdout io.Writer, hooks []client.EnvironmentWebhook) {
	if len(hooks) == 0 {
		return
	}
	w := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tDISPLAY NAME\tURL\tACTIVE\tFORMAT")
	for _, h := range hooks {
		format := h.Format
		if format == "" {
			format = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%t\t%s\n", h.Name, h.DisplayName, h.PayloadURL, h.Active, format)
	}
	_ = w.Flush()
}

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

func newEnvWebhookLsCmd(env *envCommand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls [<org-name>/][<project-name>/]<environment-name>",
		Short: "List environment webhooks.",
		Long: "List environment webhooks\n" +
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
				return fmt.Errorf("the ls command does not accept versions")
			}

			hooks, err := env.esc.client.ListEnvironmentWebhooks(ctx, ref.orgName, ref.projectName, ref.envName)
			if err != nil {
				return err
			}

			printWebhooks(env.esc.stdout, hooks)
			return nil
		},
	}

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

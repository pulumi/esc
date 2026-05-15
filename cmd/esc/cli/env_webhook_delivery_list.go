// Copyright 2026, Pulumi Corporation.

package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/pulumi/esc/cmd/esc/cli/client"
)

func newEnvWebhookDeliveryListCmd(env *envCommand) *cobra.Command {
	var (
		utc   bool
		count int
	)

	cmd := &cobra.Command{
		Use:     "list [<org-name>/][<project-name>/]<environment-name> <webhook-name>",
		Aliases: []string{"ls"},
		Short:   "List environment webhook deliveries.",
		Long: "[EXPERIMENTAL] List environment webhook deliveries\n" +
			"\n" +
			"This command lists the deliveries recorded for the named webhook.\n",
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
				return errors.New("the list command does not accept versions")
			}
			if count < 0 {
				return errors.New("--count must be non-negative")
			}

			webhookName := args[0]
			if webhookName == "" {
				return errors.New("webhook name cannot be empty")
			}

			deliveries, err := env.esc.client.ListEnvironmentWebhookDeliveries(
				ctx, ref.orgName, ref.projectName, ref.envName, webhookName)
			if err != nil {
				return err
			}

			if count > 0 && len(deliveries) > count {
				deliveries = deliveries[:count]
			}

			printWebhookDeliveries(env.esc.stdout, deliveries, utcFlag(utc))
			return nil
		},
	}

	cmd.Flags().BoolVar(&utc, "utc", false, "display times in UTC")
	cmd.Flags().IntVar(&count, "count", 0, "the maximum number of deliveries to return (all if unset)")

	return cmd
}

func printWebhookDeliveries(stdout io.Writer, ds []client.EnvironmentWebhookDelivery, utc utcFlag) {
	if len(ds) == 0 {
		return
	}
	w := tabwriter.NewWriter(stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tKIND\tTIMESTAMP\tRESPONSE\tDURATION (ms)")
	for _, d := range ds {
		ts := time.Unix(d.Timestamp, 0)
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\n",
			d.ID, d.Kind, utc.time(ts).Format(time.RFC3339), d.ResponseCode, d.Duration)
	}
	_ = w.Flush()
}

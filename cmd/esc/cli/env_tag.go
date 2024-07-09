// Copyright 2024, Pulumi Corporation.

package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/pulumi/esc/cmd/esc/cli/client"
	"github.com/pulumi/esc/cmd/esc/cli/style"
	"github.com/spf13/cobra"
)

func newEnvTagCmd(env *envCommand) *cobra.Command {
	var utc bool

	cmd := &cobra.Command{
		Use:   "tag [<org-name>/]<environment-name> <name>:<value>",
		Args:  cobra.RangeArgs(1, 3),
		Short: "Manage environment tags",
		Long: "Manage environment tags\n" +
			"\n" +
			"This command creates or updates a tag with the given name on the specified environment.\n" +
			"\n" +
			"Subcommands exist for listing and removing tags.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			ref, args, err := env.getEnvRef(args)
			if err != nil {
				return err
			}
			if ref.version != "" {
				return errors.New("the tag command does not accept versions")
			}

			if len(args) == 0 {
				return errors.New("must specify an environment tag")
			}

			// parse tag argument into name / value variables
			tagParts := strings.Split(args[0], ":")
			if len(tagParts) < 2 {
				return errors.New("must specify both key and value")
			}

			name := tagParts[0]
			value := tagParts[1]

			if name == "" {
				return errors.New("environment tag name cannot be empty")
			}
			if value == "" {
				return errors.New("environment tag value cannot be empty")
			}

			var curTag *client.EnvironmentTag
			after := ""
			count := 500
			for {
				options := client.ListEnvironmentTagsOptions{
					After: after,
					Count: &count,
				}
				tags, next, err := env.esc.client.ListEnvironmentTags(ctx, ref.orgName, ref.envName, options)
				if err != nil {
					return err
				}

				after = next
				for _, t := range tags {
					if t.Name == name {
						curTag = t
						break
					}
				}

				if after == "0" {
					break
				}
			}

			st := style.NewStylist(style.Profile(env.esc.stdout))

			if curTag != nil {
				if curTag.Name == name && curTag.Value == value {
					printTag(env.esc.stdout, st, curTag, utcFlag(utc))
					return nil
				}

				t, err := env.esc.client.UpdateEnvironmentTag(ctx, ref.orgName, ref.envName, curTag.ID, curTag.Name, curTag.Value, name, value)
				if err == nil {
					printTag(env.esc.stdout, st, t, utcFlag(utc))
					return nil
				}
				return err
			}

			t, err := env.esc.client.CreateEnvironmentTag(ctx, ref.orgName, ref.envName, name, value)
			if err != nil {
				return err
			}

			printTag(env.esc.stdout, st, t, utcFlag(utc))

			return nil
		},
	}

	cmd.AddCommand(newEnvTagLsCmd(env))
	cmd.AddCommand(newEnvTagRmCmd(env))

	cmd.Flags().BoolVar(&utc, "utc", false, "display times in UTC")

	return cmd
}

func printTag(stdout io.Writer, st *style.Stylist, tag *client.EnvironmentTag, utc utcFlag) {
	rules := style.Default()

	st.Fprintf(stdout, rules.LinkText, "Name: %v\n", tag.Name)
	st.Fprintf(stdout, rules.LinkText, "Value: %v\n", tag.Value)

	fmt.Fprintf(stdout, "Last updated at %v by ", utc.time(tag.Modified))
	if tag.EditorLogin == "" {
		fmt.Fprintf(stdout, "<unknown>")
	} else {
		fmt.Fprintf(stdout, "%v <%v>", tag.EditorName, tag.EditorLogin)
	}
	fmt.Fprintln(stdout)
}

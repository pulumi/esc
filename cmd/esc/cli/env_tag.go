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
		Use:   "tag [<org-name>/]<environment-name>[/<tag-id>] [<key>]:<value>",
		Args:  cobra.RangeArgs(1, 3),
		Short: "Manage environment tags",
		Long: "Manage environment tags\n" +
			"\n" +
			"This command creates or updates a tag with the given name on the specified environment.\n" +
			"An environments tag key can also be updated by passing in an optional tag ID\n" +
			"along with the updated key / value fields. Otherwise, the key will be used as the\n" +
			"tag identifier and only the value will be updated.\n" +
			"\n" +
			"Subcommands exist for listing and removing tags.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			// parse out tag-id if one has been specified
			// if one has not been specified use the key as unique ID by getting current tags for

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

			// parse out environment name and optional tag ID
			envName, tagID, _ := strings.Cut(ref.envName, "/")

			// parse tag argument into key / value variables
			tagParts := strings.Split(args[0], ":")
			var key string
			var value string
			// key is optional for updates
			if len(tagParts) > 1 {
				key = tagParts[0]
				value = tagParts[1]
			} else {
				value = tagParts[0]
			}

			// if neither a key or tag ID were specified we cannot create or update an environment tag
			if key == "" && tagID == "" {
				return fmt.Errorf("tag with key %v not found", key)
			}

			curKey := key
			curVal := value
			after := ""
			count := 500
			for {
				options := client.ListEnvironmentTagsOptions{
					After: after,
					Count: &count,
				}
				tags, next, err := env.esc.client.ListEnvironmentTags(ctx, ref.orgName, envName, options)
				if err != nil {
					return err
				}

				after = next
				for _, t := range tags {
					if t.Name == key || t.ID == tagID {
						tagID = t.ID
						curKey = t.Name
						curVal = t.Value
						break
					}
				}

				if after == "0" {
					break
				}
			}

			// if can't find tag matching specified key return error
			if curKey == "" && tagID != "" {
				return fmt.Errorf("tag with ID %v not found", tagID)
			}

			st := style.NewStylist(style.Profile(env.esc.stdout))

			t, err := env.esc.client.UpdateEnvironmentTag(ctx, ref.orgName, envName, tagID, curKey, curVal, key, value)
			if err == nil {
				printTag(env.esc.stdout, st, t, utcFlag(utc))
				return nil
			}
			if !client.IsNotFound(err) {
				return err
			}

			// create request
			if key == "" || value == "" {
				return errors.New("environment tags must have both a key and value")
			}

			t, err = env.esc.client.CreateEnvironmentTag(ctx, ref.orgName, envName, key, value)
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
	fmt.Fprintf(stdout, "ID: %v\n", tag.ID)
	fmt.Fprintln(stdout)
}

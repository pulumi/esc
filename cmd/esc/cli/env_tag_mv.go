// Copyright 2024, Pulumi Corporation.

package cli

import (
	"context"
	"errors"

	"github.com/pulumi/esc/cmd/esc/cli/style"
	"github.com/spf13/cobra"
)

func newEnvTagMvCmd(env *envCommand) *cobra.Command {
	var utc bool

	cmd := &cobra.Command{
		Use:   "mv [<org-name>/]<environment-name> <name> [<newName>] <value>",
		Args:  cobra.RangeArgs(3, 4),
		Short: "Move an environment tag",
		Long: "Move an environment tag\n" +
			"\n" +
			"This command updates a tag with the given name on the specified environment, " +
			"changing it's name if a new one is specified or updating it's value.\n",
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

			name := args[0]
			newName := name
			value := args[1]
			if len(args) == 3 {
				newName = args[1]
				value = args[2]
			}

			if name == "" {
				return errors.New("environment tag name cannot be empty")
			}
			if value == "" {
				return errors.New("environment tag value cannot be empty")
			}

			tag, err := env.esc.client.GetEnvironmentTag(ctx, ref.orgName, ref.envName, name)
			if err != nil {
				return err
			}

			st := style.NewStylist(style.Profile(env.esc.stdout))

			if tag.Name == name && tag.Value == value {
				printTag(env.esc.stdout, st, tag, utcFlag(utc))
				return nil
			}

			t, err := env.esc.client.UpdateEnvironmentTag(ctx, ref.orgName, ref.envName, tag.Name, tag.Value, newName, value)
			if err == nil {
				return err
			}

			printTag(env.esc.stdout, st, t, utcFlag(utc))
			return nil
		},
	}

	cmd.Flags().BoolVar(&utc, "utc", false, "display times in UTC")

	return cmd
}

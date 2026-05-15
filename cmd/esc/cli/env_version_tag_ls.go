// Copyright 2023, Pulumi Corporation.

package cli

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"

	"github.com/pulumi/esc/cmd/esc/cli/client"
)

func newEnvVersionTagLsCmd(env *envCommand) *cobra.Command {
	var pagerFlag string
	var utc bool

	cmd := &cobra.Command{
		Use:   "ls [<org-name>/][<project-name>/]<environment-name>",
		Short: "List tagged versions.",
		Long: "List tagged versions\n" +
			"\n" +
			"This command lists an environment's tagged versions.\n",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
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
				return fmt.Errorf("the ls command does not accept versions")
			}
			_ = args

			after := ""
			return env.esc.pager.Run(pagerFlag, env.esc.stdout, env.esc.stderr, func(ctx context.Context, stdout io.Writer) error {
				rows := []cmdutil.TableRow{}
				count := 500
				for {
					options := client.ListEnvironmentRevisionTagsOptions{
						After: after,
						Count: &count,
					}
					tags, err := env.esc.client.ListEnvironmentRevisionTags(ctx, ref.orgName, ref.projectName, ref.envName, options)
					if err != nil {
						return err
					}
					if len(tags) == 0 {
						break
					}
					after = tags[len(tags)-1].Name

					for _, t := range tags {
						rows = append(rows, cmdutil.TableRow{
							Columns: []string{
								t.Name,
								strconv.Itoa(t.Revision),
								utcFlag(utc).time(t.Modified).String(),
								revisionTagEditor(t),
							},
						})
					}
				}
				if len(rows) == 0 {
					return nil
				}
				return cmdutil.FprintTable(stdout, cmdutil.Table{
					Headers: []string{"NAME", "REVISION", "MODIFIED", "EDITOR"},
					Rows:    rows,
				})
			})
		},
	}

	cmd.Flags().StringVar(&pagerFlag, "pager", "", "the command to use to page through the environment's version tags")
	cmd.Flags().BoolVar(&utc, "utc", false, "display times in UTC")

	return cmd
}

func revisionTagEditor(t client.EnvironmentRevisionTag) string {
	if t.EditorLogin == "" {
		return "<unknown>"
	}
	return fmt.Sprintf("%s <%s>", t.EditorName, t.EditorLogin)
}

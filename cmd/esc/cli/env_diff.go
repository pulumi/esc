// Copyright 2023, Pulumi Corporation.

package cli

import (
	"bytes"
	"context"
	"fmt"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"

	"github.com/pulumi/esc/cmd/esc/cli/style"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
)

func newEnvDiffCmd(env *envCommand) *cobra.Command {
	var format string
	var showSecrets bool
	var pathString string

	diff := &envGetCommand{env: env}

	cmd := &cobra.Command{
		Use:   "diff [<org-name>/]<environment-name>[:<revision-or-tag>] [<revision-or-tag>]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Show changes between revisions.",
		Long: "Show changes between revisions\n" +
			"\n" +
			"This command fetches the current definition for the named environment and gets a\n" +
			"value within it. The path to the value to set is a Pulumi property path. The value\n" +
			"is printed to stdout as YAML.\n",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			orgName, envName, revisionOrTag, args, err := env.getEnvName(args)
			if err != nil {
				return err
			}
			baseRevisionOrTag := revisionOrTag
			if baseRevisionOrTag == "" {
				baseRevisionOrTag = "latest"
			}

			tipRevisionOrTag := "latest"
			if len(args) != 0 {
				tipRevisionOrTag = args[0]
			}

			var path resource.PropertyPath
			if pathString != "" {
				path, err = resource.ParsePropertyPath(pathString)
				if err != nil {
					return fmt.Errorf("invalid path: %w", err)
				}
			}

			switch format {
			case "":
				// OK
			case "detailed", "json", "string":
				return diff.diffValue(ctx, orgName, envName, baseRevisionOrTag, tipRevisionOrTag, path, format, showSecrets)
			case "dotenv":
				if len(path) != 0 {
					return fmt.Errorf("output format '%s' may not be used with a property path", format)
				}
				return diff.diffValue(ctx, orgName, envName, baseRevisionOrTag, tipRevisionOrTag, path, format, showSecrets)
			case "shell":
				if len(path) != 0 {
					return fmt.Errorf("output format '%s' may not be used with a property path", format)
				}
				return diff.diffValue(ctx, orgName, envName, baseRevisionOrTag, tipRevisionOrTag, path, format, showSecrets)
			default:
				return fmt.Errorf("unknown output format %q", format)
			}

			baseData, err := diff.getEnvironment(ctx, orgName, envName, baseRevisionOrTag, path, showSecrets)
			if err != nil {
				return err
			}
			if baseData == nil {
				baseData = &envGetTemplateData{}
			}

			tipData, err := diff.getEnvironment(ctx, orgName, envName, tipRevisionOrTag, path, showSecrets)
			if err != nil {
				return err
			}
			if tipData == nil {
				tipData = &envGetTemplateData{}
			}

			baseRef := fmt.Sprintf("%s:%s", envName, baseRevisionOrTag)
			tipRef := fmt.Sprintf("%s:%s", envName, tipRevisionOrTag)
			data := diff.diff(baseRef, baseData, tipRef, tipData)

			var markdown bytes.Buffer
			if err := envDiffTemplate.Execute(&markdown, data); err != nil {
				return fmt.Errorf("internal error: rendering: %w", err)
			}

			if !cmdutil.InteractiveTerminal() {
				fmt.Fprint(diff.env.esc.stdout, markdown.String())
				return nil
			}

			renderer, err := style.Glamour(diff.env.esc.stdout, glamour.WithWordWrap(0))
			if err != nil {
				return fmt.Errorf("internal error: creating renderer: %w", err)
			}
			rendered, err := renderer.Render(markdown.String())
			if err != nil {
				rendered = markdown.String()
			}
			fmt.Fprint(diff.env.esc.stdout, rendered)
			return nil
		},
	}

	cmd.Flags().StringVarP(
		&format, "format", "f", "",
		"the output format to use. May be 'dotenv', 'json', 'yaml', 'detailed', or 'shell'")
	cmd.Flags().BoolVar(
		&showSecrets, "show-secrets", false,
		"Show static secrets in plaintext rather than ciphertext")
	cmd.Flags().StringVar(
		&pathString, "path", "",
		"Show the diff for a specific path")

	return cmd
}

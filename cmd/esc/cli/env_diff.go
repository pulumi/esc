// Copyright 2023, Pulumi Corporation.

package cli

import (
	"bytes"
	"context"
	"fmt"
	"strings"

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
		Use:   "diff [<org-name>/]<environment-name>[@<version>] [@<version>]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Show changes between versions.",
		Long: "Show changes between versions\n" +
			"\n" +
			"This command displays the changes between two versions of an environment.\n" +
			"The first argument is the environment and base version for the diff and\n" +
			"the second argument is the comparison version. If the second argument is\n" +
			"omitted, the 'latest' tag is used.",
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
			baseVersion := revisionOrTag
			if baseVersion == "" {
				baseVersion = "latest"
			}

			compareVersion := "latest"
			if len(args) != 0 {
				compareVersion, _ = strings.CutPrefix(args[0], "@")
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
				return diff.diffValue(ctx, orgName, envName, baseVersion, compareVersion, path, format, showSecrets)
			case "dotenv":
				if len(path) != 0 {
					return fmt.Errorf("output format '%s' may not be used with a property path", format)
				}
				return diff.diffValue(ctx, orgName, envName, baseVersion, compareVersion, path, format, showSecrets)
			case "shell":
				if len(path) != 0 {
					return fmt.Errorf("output format '%s' may not be used with a property path", format)
				}
				return diff.diffValue(ctx, orgName, envName, baseVersion, compareVersion, path, format, showSecrets)
			default:
				return fmt.Errorf("unknown output format %q", format)
			}

			baseData, err := diff.getEnvironment(ctx, orgName, envName, baseVersion, path, showSecrets)
			if err != nil {
				return err
			}
			if baseData == nil {
				baseData = &envGetTemplateData{}
			}

			compareData, err := diff.getEnvironment(ctx, orgName, envName, compareVersion, path, showSecrets)
			if err != nil {
				return err
			}
			if compareData == nil {
				compareData = &envGetTemplateData{}
			}

			baseRef := fmt.Sprintf("%s:%s", envName, baseVersion)
			compareRef := fmt.Sprintf("%s:%s", envName, compareVersion)
			data := diff.diff(baseRef, baseData, compareRef, compareData)

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
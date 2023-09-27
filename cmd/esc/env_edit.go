// Copyright 2023, Pulumi Corporation.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"

	"github.com/spf13/cobra"

	"github.com/pulumi/esc"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
)

type envEditCommand struct {
	env *envCommand

	editorFlag string
}

func newEnvEditCmd(env *envCommand) *cobra.Command {
	edit := &envEditCommand{env: env}

	cmd := &cobra.Command{
		Use:   "edit [<org-name>/]<environment-name>",
		Args:  cmdutil.MaximumNArgs(1),
		Short: "Open an environment for editing.",
		Long: "Open an environment for editing\n" +
			"\n" +
			"This command fetches the current definition for the named environment and opens it\n" +
			"for editing in an editor. The editor defaults to the value of the VISUAL environment\n" +
			"variable. If VISUAL is not set, EDITOR is used. These values are interpreted as\n" +
			"commands to which the name of the temporary file used for the environment is appended.\n",
		Run: cmdutil.RunFunc(func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			editor, err := edit.getEditor()
			if err != nil {
				return err
			}

			if err := env.esc.getCachedClient(ctx); err != nil {
				return err
			}

			orgName, envName, args, err := edit.env.getEnvName(args)
			if err != nil {
				return err
			}
			_ = args

			yaml, tag, err := edit.env.esc.client.GetEnvironment(ctx, orgName, envName)
			if err != nil {
				return fmt.Errorf("getting environment definition: %w", err)
			}

			var checked []byte
			if len(yaml) != 0 {
				env, diags, err := edit.env.esc.client.CheckYAMLEnvironment(ctx, orgName, yaml)
				if err != nil {
					checked = []byte(fmt.Sprintf("# checking environment: %v\n", err))
				} else if len(diags) != 0 {
					var stderr bytes.Buffer
					err = edit.env.writeYAMLEnvironmentDiagnostics(&stderr, envName, yaml, diags)
					contract.IgnoreError(err)
					checked = stderr.Bytes()
				} else {
					var stdout bytes.Buffer
					enc := json.NewEncoder(&stdout)
					enc.SetIndent("", "  ")
					err = enc.Encode(esc.NewValue(env.Properties).ToJSON(false))
					contract.IgnoreError(err)
					checked = stdout.Bytes()
				}
			}

			newYAML, err := editWithYAMLEditor(editor, yaml, checked)
			if err != nil {
				return err
			}

			diags, err := edit.env.esc.client.UpdateEnvironment(ctx, orgName, envName, newYAML, tag)
			if err != nil {
				return fmt.Errorf("updating environment definition: %w", err)
			}
			if len(diags) != 0 {
				return edit.env.writeYAMLEnvironmentDiagnostics(os.Stderr, envName, newYAML, diags)
			}

			return nil
		}),
	}

	cmd.Flags().StringVar(&edit.editorFlag, "editor", "", "the command to use to edit the environment definition")

	return cmd
}

func (edit *envEditCommand) getEditor() ([]string, error) {
	editor := edit.editorFlag

	if editor == "" {
		editor = os.Getenv("VISUAL")
		if editor == "" {
			editor = os.Getenv("EDITOR")
		}
	}

	var args []string
	for {
		editor = strings.TrimLeftFunc(editor, unicode.IsSpace)
		if len(editor) == 0 {
			break
		}

		if editor[0] == '"' {
			var arg strings.Builder
			for i := 1; i != len(editor); {
				c := editor[i]
				if c == '"' {
					editor = editor[i+1:]
					break
				} else if i+1 < len(editor) && c == '\\' && editor[i+1] == '"' {
					arg.WriteByte('"')
					i += 2
				} else {
					arg.WriteByte(editor[i])
					i++
				}
			}
			args = append(args, arg.String())
		} else {
			sep := strings.IndexFunc(editor, unicode.IsSpace)
			if sep == -1 {
				args = append(args, editor)
				break
			}
			args, editor = append(args, editor[:sep]), editor[sep+1:]
		}
	}
	if len(args) == 0 {
		return nil, errors.New("No available editor. Please use the --editor flag or set one of the " +
			"VISUAL or EDITOR environment variables.")
	}

	path, err := exec.LookPath(args[0])
	if err != nil {
		return nil, fmt.Errorf("finding %q on path: %w", args[0], err)
	}

	args[0] = path
	return args, nil
}

func editWithYAMLEditor(editor []string, yaml, checked []byte) ([]byte, error) {
	filename, err := func() (string, error) {
		f, err := os.CreateTemp("", "*.yaml")
		if err != nil {
			return "", err
		}
		defer contract.IgnoreClose(f)

		if _, err = f.Write(yaml); err != nil {
			rmErr := os.Remove(f.Name())
			contract.IgnoreError(rmErr)
			return "", err
		}

		if len(checked) != 0 {
			if len(yaml) != 0 && yaml[len(yaml)-1] != '\n' {
				fmt.Fprintln(f, "")
			}

			fmt.Fprintln(f, "---")
			fmt.Fprintln(f, "# Please edit the environment definition above.")
			fmt.Fprintln(f, "# The object below is the current result of")
			fmt.Fprintln(f, "# evaluating the environment and will not be")
			fmt.Fprintln(f, "# saved.")
			fmt.Fprintln(f, "")
			if _, err = f.Write(checked); err != nil {
				rmErr := os.Remove(f.Name())
				contract.IgnoreError(rmErr)
				return "", err
			}
		}

		return f.Name(), nil
	}()
	if err != nil {
		return nil, fmt.Errorf("writing temporary file: %w", err)
	}
	defer func() {
		err := os.Remove(filename)
		contract.IgnoreError(err)
	}()

	//nolint:gosec
	cmd := exec.Command(editor[0], append(editor[1:], filename)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("editor: %w", err)
	}

	new, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("reading temporary file: %w", err)
	}

	sep := bytes.Index(new, []byte("---"))
	if sep != -1 {
		isDocSep := true
		if sep+len("---") < len(new) && new[sep+len("---")] != '\n' {
			isDocSep = false
		}
		if sep != 0 && new[sep-1] != '\n' {
			isDocSep = false
		}
		if isDocSep {
			new = new[:sep]
		}
	}

	return new, nil
}

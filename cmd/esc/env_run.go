// Copyright 2023, Pulumi Corporation.

package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	aho_corasick "github.com/petar-dambovaliev/aho-corasick"

	"github.com/pulumi/esc"
	"github.com/pulumi/esc/ast"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"github.com/spf13/cobra"
)

func newReplacer(secrets []string) aho_corasick.Replacer {
	// Ignore very short secrets (anything less than 3 characters). Such secrets have low entropy, and redacting them is
	// unlikely to be useful (in the case of _empty_ secrets, including them is harmful, as it causes the redactor
	// to insert "[secret]" after every character).
	d := 0
	for _, s := range secrets {
		if len(s) >= 3 {
			secrets[d], d = s, d+1
		}
	}
	secrets = secrets[:d]

	builder := aho_corasick.NewAhoCorasickBuilder(aho_corasick.Opts{
		MatchKind: aho_corasick.StandardMatch,
	})
	return aho_corasick.NewReplacer(builder.Build(secrets))
}

type redactor struct {
	w        io.Writer
	replacer aho_corasick.Replacer
	line     bytes.Buffer
}

func newRedactor(w io.Writer, replacer aho_corasick.Replacer) *redactor {
	return &redactor{w: w, replacer: replacer}
}

func (w *redactor) Write(b []byte) (int, error) {
	written := 0
	for {
		newline := bytes.IndexByte(b, '\n')
		if newline == -1 {
			n, err := w.line.Write(b)
			contract.IgnoreError(err)

			return written + n, nil
		}

		n := w.line.Len()
		_, err := w.line.Write(b[:newline+1])
		contract.IgnoreError(err)

		redacted := w.replacer.ReplaceAllFunc(w.line.String(), func(m aho_corasick.Match) (string, bool) {
			return "[secret]", true
		})

		if _, err = w.w.Write([]byte(redacted)); err != nil {
			w.line.Truncate(n)
			return written, err
		}
		w.line.Reset()

		b, written = b[newline+1:], written+newline+1
	}
}

func (w *redactor) Close() error {
	if w.line.Len() != 0 {
		bytes := w.line.Bytes()
		w.line.Reset()

		_, err := w.w.Write(bytes)
		return err
	}
	return nil
}

func newEnvRunCmd(envcmd *envCommand) *cobra.Command {
	var interactive bool
	var duration time.Duration

	cmd := &cobra.Command{
		Use:   "run [<org-name>/]<environment-name> [command]",
		Args:  cobra.ArbitraryArgs,
		Short: "Open the environment with the given name and runs a command.",
		Long: "Open the environment with the given name and runs a command\n" +
			"\n" +
			"This command opens the environment with the given name and runs the given command.\n" +
			"If the opened environment contains a top-level 'environmentVariables' object, each\n" +
			"key-value pair in the object is made available to the command as an environment\n" +
			"variable.\n" +
			"\n" +
			"By default, the command to run is assumed to be non-interactive and its output\n" +
			"streams are filtered to remove any secret values. Use the -i flag to run interactive\n" +
			"commands, which will disable filtering.\n",
		Run: cmdutil.RunFunc(func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if err := envcmd.esc.getCachedClient(ctx); err != nil {
				return err
			}

			orgName, envName, args, err := envcmd.getEnvName(args)
			if err != nil {
				return err
			}

			if len(args) == 0 {
				return fmt.Errorf("no command specified")
			}
			command, err := exec.LookPath(args[0])
			if err != nil {
				return fmt.Errorf("resolving command: %w", err)
			}
			args = args[1:]

			env, diags, err := envcmd.openEnvironment(ctx, orgName, envName, duration)
			if err != nil {
				return err
			}
			if len(diags) != 0 {
				return envcmd.writePropertyEnvironmentDiagnostics(os.Stderr, diags)
			}

			var secrets []string

			environ := os.Environ()
			if vars, ok := env.Properties["environmentVariables"].Value.(map[string]esc.Value); ok {
				for k, v := range vars {
					if strValue, ok := v.Value.(string); ok {
						if v.Secret {
							secrets = append(secrets, strValue)
						}

						environ = append(environ, fmt.Sprintf("%v=%v", k, strValue))
					}
				}
			}

			envV := esc.NewValue(env.Properties)
			for i, v := range args {
				interp, diags := ast.Interpolate(v)
				if !diags.HasErrors() {
					var arg strings.Builder
					for _, p := range interp.Parts {
						arg.WriteString(p.Text)
						if p.Value != nil {
							path := make(resource.PropertyPath, len(p.Value.Accessors))
							for i, accessor := range p.Value.Accessors {
								switch accessor := accessor.(type) {
								case *ast.PropertyName:
									path[i] = accessor.Name
								case *ast.PropertySubscript:
									path[i] = accessor.Index
								default:
									contract.Failf("unexpected accessor of type %T", accessor)
								}
							}
							if val, ok := getEnvValue(envV, path); ok {
								str := val.ToString(false)
								if val.Secret {
									secrets = append(secrets, str)
								}

								arg.WriteString(str)
							}
						}
					}
					args[i] = arg.String()
				}
			}

			runCmd := exec.Command(command, args...)
			runCmd.Env = environ

			stdout, stderr := envcmd.esc.stdout, envcmd.esc.stderr
			if !interactive {
				replacer := newReplacer(secrets)
				stdout, stderr = newRedactor(stdout, replacer), newRedactor(stderr, replacer)
			}

			runCmd.Stdin = envcmd.esc.stdin
			runCmd.Stdout = stdout
			runCmd.Stderr = stderr
			return runCmd.Run()
		}),
	}

	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "true to treat the command as interactive and disable output filters")
	cmd.Flags().DurationVarP(&duration, "lifetime", "l", 2*time.Hour, "the lifetime of the opened environment")

	return cmd
}

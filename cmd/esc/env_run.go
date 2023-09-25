// Copyright 2023, Pulumi Corporation.

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pulumi/esc"
	"github.com/pulumi/esc/ast"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"github.com/spf13/cobra"
)

func newEnvRunCmd(envcmd *envCommand) *cobra.Command {
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
			"variable.\n",
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

			environ := os.Environ()
			if vars, ok := env.Properties["environmentVariables"].Value.(map[string]esc.Value); ok {
				for k, v := range vars {
					if strValue, ok := v.Value.(string); ok {
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
								arg.WriteString(val.ToString(false))
							}
						}
					}
					args[i] = arg.String()
				}
			}

			runCmd := exec.Command(command, args...)
			runCmd.Env = environ
			runCmd.Stdin = os.Stdin
			runCmd.Stdout = os.Stdout
			runCmd.Stderr = os.Stderr
			return runCmd.Run()
		}),
	}

	cmd.Flags().DurationVarP(&duration, "lifetime", "l", 2*time.Hour, "the lifetime of the opened environment")

	return cmd
}

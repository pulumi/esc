// Copyright 2023, Pulumi Corporation.

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/pulumi/esc/cmd/esc/internal/client"
	"github.com/pulumi/esc/cmd/esc/internal/workspace"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag/colors"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
)

type Options struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	Colors colors.Colorization
}

type escCommand struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	colors colors.Colorization

	account workspace.Account

	client *client.Client
}

// New creates a new ESC command instance.
func New(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "esc",
		Short: "Pulumi ESC command line",
		Long: "Pulumi ESC - Manage environments, secrets, and configuration\n" +
			"\n" +
			"To begin working with Pulumi ESC, run the `esc env init` command:\n" +
			"\n" +
			"    $ esc env init\n" +
			"\n" +
			"This will prompt you to create a new environment to hold secrets and configuration.\n" +
			"\n" +
			"The most common commands from there are:\n" +
			"\n" +
			"    - esc env get  : Get a property in an environment definition\n" +
			"    - esc env set  : Set a property in an environment definition\n" +
			"    - esc env edit : Edit an environment definition\n" +
			"    - esc env run  : Run a command within the context of an environment\n" +
			"    - esc env open : Open an environment and access its contents\n" +
			"    - esc env ls   : List available environments\n" +
			"\n" +
			"For more information, please visit the project page: https://www.pulumi.com/docs/esc",
	}

	if opts == nil {
		opts = &Options{}
	}

	esc := &escCommand{
		stdin:  valueOrDefault(opts.Stdin, io.Reader(os.Stdin)),
		stdout: valueOrDefault(opts.Stdout, io.Writer(os.Stdout)),
		stderr: valueOrDefault(opts.Stderr, io.Writer(os.Stderr)),
		colors: valueOrDefault(opts.Colors, cmdutil.GetGlobalColorization()),
	}

	cmd.AddCommand(newLoginCmd(esc))
	cmd.AddCommand(newEnvCmd(esc))
	cmd.AddCommand(newVersionCmd(esc))

	return cmd
}

func valueOrDefault[T comparable](v, def T) T {
	var zero T
	if v == zero {
		return def
	}
	return v
}

func (esc *escCommand) confirmPrompt(prompt, name string) bool {
	if prompt != "" {
		fmt.Fprint(esc.stdout,
			esc.colors.Colorize(
				fmt.Sprintf("%s%s%s\n", colors.SpecAttention, prompt, colors.Reset)))
	}

	fmt.Fprint(esc.stdout,
		esc.colors.Colorize(
			fmt.Sprintf("%sPlease confirm that this is what you'd like to do by typing `%s%s%s`:%s ",
				colors.SpecAttention, colors.SpecPrompt, name, colors.SpecAttention, colors.Reset)))

	reader := bufio.NewReader(esc.stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line) == name
}

// Copyright 2023, Pulumi Corporation.

package main

import (
	"fmt"

	"github.com/pulumi/esc/cmd/esc/internal/version"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
	"github.com/spf13/cobra"
)

func newVersionCmd(esc *escCommand) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print esc's version number",
		Args:  cmdutil.NoArgs,
		Run: cmdutil.RunFunc(func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(esc.stdout, "%v\n", version.Version)
			return nil
		}),
	}
}

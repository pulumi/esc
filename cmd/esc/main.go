// Copyright 2023, Pulumi Corporation.

package main

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/pulumi/esc/cmd/esc/cli"
	"github.com/pulumi/esc/cmd/esc/cli/version"
)

// printRetirementBanner prints a notice informing users that the standalone esc CLI has been
// retired in favor of the Pulumi CLI's `pulumi env` subcommand. It is printed to stderr before
// every command so that it never corrupts machine-readable output on stdout.
func printRetirementBanner(w io.Writer) {
	lines := []string{
		"The ESC CLI has been retired.",
		"Download the Pulumi CLI and use the 'pulumi env' subcommand instead.",
		"See: https://www.pulumi.com/docs/iac/cli",
	}

	width := 0
	for _, line := range lines {
		if len(line) > width {
			width = len(line)
		}
	}

	border := strings.Repeat("─", width+2)
	fmt.Fprintf(w, "┌%s┐\n", border)
	for _, line := range lines {
		fmt.Fprintf(w, "│ %-*s │\n", width, line)
	}
	fmt.Fprintf(w, "└%s┘\n", border)
}

// panicHandler displays an emergency error message to the user and a stack trace to
// report the panic.
//
// finished should be set to false when the handler is deferred and set to true as the
// last statement in the scope. This trick is necessary to avoid catching and then
// discarding a panic(nil).
func panicHandler() {
	if panicPayload := recover(); panicPayload != nil {
		stack := string(debug.Stack())
		fmt.Fprintln(os.Stderr, "================================================================================")
		fmt.Fprintln(os.Stderr, "The esc CLI encountered a fatal error. This is a bug!")
		fmt.Fprintln(os.Stderr, "We would appreciate a report: https://github.com/pulumi/esc/issues/")
		fmt.Fprintln(os.Stderr, "Please provide all of the below text in your report.")
		fmt.Fprintln(os.Stderr, "================================================================================")
		fmt.Fprintf(os.Stderr, "esc Version:      %s\n", version.Version)
		fmt.Fprintf(os.Stderr, "Go Version:       %s\n", runtime.Version())
		fmt.Fprintf(os.Stderr, "Go Compiler:      %s\n", runtime.Compiler)
		fmt.Fprintf(os.Stderr, "Architecture:     %s\n", runtime.GOARCH)
		fmt.Fprintf(os.Stderr, "Operating System: %s\n", runtime.GOOS)
		fmt.Fprintf(os.Stderr, "Panic:            %s\n\n", panicPayload)
		fmt.Fprintln(os.Stderr, stack)
		os.Exit(1)
	}
}

func main() {
	printRetirementBanner(os.Stderr)

	err := func() error {
		defer panicHandler()
		return cli.New(nil).Execute()
	}()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

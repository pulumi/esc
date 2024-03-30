// Copyright 2023, Pulumi Corporation.

package pager

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

func runSystemPager(pager string, stdout, stderr io.Writer, f func(context.Context, io.Writer) error) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.Command(pager)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("running pager: %w", err)
	}

	done := make(chan error)
	go func() {
		done <- func() error {
			defer stdin.Close()
			return f(ctx, stdin)
		}()
	}()

	if cmdErr := cmd.Run(); cmdErr != nil {
		return fmt.Errorf("running pager: %w", cmdErr)
	}
	cancel()

	return <-done
}
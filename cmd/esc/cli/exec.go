// Copyright 2023, Pulumi Corporation.

package cli

import (
	"os/exec"
	"syscall"
)

type cmdExec interface {
	LookPath(command string) (string, error)
	Run(cmd *exec.Cmd) error
	Exec(cmd *exec.Cmd) error
}

type defaultCmdExec int

func newCmdExec() cmdExec {
	return defaultCmdExec(0)
}

func (defaultCmdExec) LookPath(command string) (string, error) {
	return exec.LookPath(command)
}

func (defaultCmdExec) Run(cmd *exec.Cmd) error {
	return cmd.Run()
}

func (defaultCmdExec) Exec(cmd *exec.Cmd) error {
	// Exec expects the first argument to be the command name, see execve(2).
	args := append([]string{cmd.Path}, cmd.Args...)
	//nolint:gosec
	return syscall.Exec(cmd.Path, args, cmd.Env)
}

// Package command is used to define a Cmd proxy. This is used
// for testing purposes.
package command

import (
	"os/exec"
)

type ExecCommandFn func(string, ...string) *Cmd

// Cmd is an exec.Cmd with a configurable Run function
type Cmd struct {
	*exec.Cmd
	Run func() error
}

// New creates a Cmd with a real exec.Cmd Run function
func New(c string, args ...string) *Cmd {
	cmd := &Cmd{
		Cmd: exec.Command(c, args...),
	}
	cmd.Run = func() error {
		return cmd.Cmd.Run()
	}
	return cmd
}

// Package command defines variation points to allow alternate
// command runners.
package command

import (
	"os/exec"
)

// ExecCommandFn defines a Cmd factory function. It takes the command to run as
// first parameter then its arguments as a slice.
type ExecCommandFn func(string, ...string) *Cmd

// Cmd is an exec.Cmd with a configurable Run function.
type Cmd struct {
	*exec.Cmd
	Run func() error
}

// New is the default factory that creates a Cmd with an os exec.Cmd Run function.
func New(c string, args ...string) *Cmd {
	cmd := &Cmd{
		Cmd: exec.Command(c, args...),
	}
	cmd.Run = func() error {
		return cmd.Cmd.Run()
	}
	return cmd
}

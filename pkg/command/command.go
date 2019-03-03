package command

import (
	"os/exec"
)

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
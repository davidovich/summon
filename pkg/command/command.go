package command

import (
	"io"
	"os/exec"
)

// Commander describes a subset of a exec.Cmd functionality for testing
type Commander interface {
	SetStdStreams(stdin io.Reader, stdout, stderr io.Writer)
	Run() error
}

type cmd struct {
	*exec.Cmd
}

// New creates a concrete commander
func New(c string, args ...string) Commander {
	return &cmd{exec.Command(c, args...)}
}

func (c *cmd) Run() error {
	return c.Cmd.Run()
}

func (c *cmd) SetStdStreams(stdin io.Reader, stdout, stderr io.Writer) {
	c.Stdin = stdin
	c.Stdout = stdout
	c.Stderr = stderr
}

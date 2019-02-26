package testutil

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/davidovich/summon/pkg/command"
	"github.com/spf13/afero"
)

// GetFs is a function that returns an Fs
var GetFs func() afero.Fs

// SetFs sets an app Fs
var SetFs func(fs afero.Fs)

// ReplaceFs replaces the real filesystem by a memory implementation
func ReplaceFs() func() {
	oldFs := GetFs()
	SetFs(afero.NewMemMapFs())
	return func() {
		SetFs(oldFs)
	}
}

type fakeCommand struct {
	*exec.Cmd
}

func (c *fakeCommand) SetStdStreams(stdin io.Reader, stdout, stderr io.Writer) {
}

func (c *fakeCommand) Run() error {
	return c.Cmd.Run()
}

// FakeExecCommand resturns a fake function which calls into testToCall
// this is used to mock an exec.Cmd
// Adapted from https://npf.io/2015/06/testing-exec-command/
func FakeExecCommand(testToCall string, stdout, stderr io.Writer) func(string, ...string) command.Commander {
	return func(c string, args ...string) command.Commander {
		cs := []string{"-test.run=" + testToCall, "--", c}
		cs = append(cs, args...)
		cmd := &fakeCommand{exec.Command(os.Args[0], cs...)}

		cmd.Stdout = stdout
		cmd.Stderr = stderr
		cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
		return cmd
	}
}

// CleanHelperArgs removes the helper process arguments
func CleanHelperArgs(helperArgs []string) []string {
	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command\n")
		os.Exit(2)
	}

	return args
}

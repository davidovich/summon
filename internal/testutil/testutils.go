package testutil

import (
	"fmt"
	"io"
	"os"
	"os/exec"

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

// FakeExecCommand resturns a fake function which calls into testToCall
// this is used to mock an exec.Cmd
// Adapted from https://npf.io/2015/06/testing-exec-command/
func FakeExecCommand(testToCall string, stdout, stderr io.Writer) func(string, ...string) *exec.Cmd {
	return func(command string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=" + testToCall, "--", command}
		cs = append(cs, args...)
		cmd := exec.Command(os.Args[0], cs...)

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

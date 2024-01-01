package testutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/spf13/afero"

	"github.com/davidovich/summon/pkg/command"
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

// Call is a recording of a fake call
type Call struct {
	Args []string
	Env  []string
	Out  string
}

type FakeCommand struct {
	Fn    func(string, ...string) *command.Cmd
	Calls []Call
}

// FakeExecCommand returns a FakeCommand which can record calls
// this is used to mock an exec.Cmd
// Adapted from https://npf.io/2015/06/testing-exec-command/
func FakeExecCommand(testToCall string) *FakeCommand {

	f := &FakeCommand{}
	f.Fn = func(c string, args ...string) *command.Cmd {
		base := []string{"-test.run=" + testToCall, "--"}
		subcallArgs := append([]string{c}, args...)

		cs := append(base, subcallArgs...)
		cmd := &command.Cmd{
			Cmd: exec.Command(os.Args[0], cs...),
		}
		call := Call{}
		call.Args = subcallArgs

		fmt.Printf("caller exec.Command: %v\n", append([]string{os.Args[0]}, cs...))
		fmt.Printf("osArgs: %v\n", os.Args)
		cmd.Run = func() error {
			stdout := &bytes.Buffer{}
			var savedStdout = cmd.Stdout
			cmd.Stdout = stdout
			call.Env = cmd.Env
			cmd.Env = append(cmd.Env, "GO_WANT_HELPER_PROCESS=1")

			err := cmd.Cmd.Run()
			io.Copy(savedStdout, stdout)
			call.Out = stdout.String()

			if err == nil {
				f.Calls = append(f.Calls, call)
			}
			return err
		}

		return cmd
	}

	return f
}

// IsHelper returns true if a process helper is wanted
func IsHelper() bool {
	return os.Getenv("GO_WANT_HELPER_PROCESS") == "1"
}

func (fc *FakeCommand) GetCalls() []Call {
	return fc.Calls
}

// TestSummonRunHelper is a testing helper for go test subprocess mocking
func TestSummonRunHelper() {
	if IsHelper() {
		defer os.Exit(0)
	}
}

// TestFailRunHelper is a non zero exiting testing helper for go test subprocess mocking.
func TestFailRunHelper() {
	if IsHelper() {
		os.Exit(1)
	}
}

package testutil

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

//Call is a recording of a fake call
type Call struct {
	Args string
	Env  []string
}

// Calls is the array of calls
type Calls struct {
	Calls []Call
}

// FakeExecCommand resturns a fake function which calls into testToCall
// this is used to mock an exec.Cmd
// Adapted from https://npf.io/2015/06/testing-exec-command/
func FakeExecCommand(testToCall string, stdout, stderr io.Writer) func(string, ...string) *command.Cmd {
	calls := 0
	return func(c string, args ...string) *command.Cmd {
		cs := []string{"-test.run=" + testToCall, "--", c}
		cs = append(cs, args...)
		cmd := &command.Cmd{
			Cmd: exec.Command(os.Args[0], cs...),
		}
		if calls == 0 {
			startCall(stdout)
		}
		cmd.Run = func() error {
			if calls > 0 {
				willAppendCall(stdout)
			}
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			cmd.Env = append(cmd.Env, "GO_WANT_HELPER_PROCESS=1")
			err := cmd.Cmd.Run()
			calls++
			return err
		}

		return cmd
	}
}

func startCall(out io.Writer) {
	out.Write([]byte("{\"Calls\":["))
}

func willAppendCall(out io.Writer) {
	out.Write([]byte(","))
}

// WriteCall marshals the executable call with env
func WriteCall(c Call, w io.Writer) error {
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	w.Write(b)
	return nil
}

// GetCalls ends and returns a call sequence
func GetCalls(out io.Reader) (*Calls, error) {
	c := &Calls{}
	buf, _ := ioutil.ReadAll(out)
	if len(buf) == 0 {
		return c, nil
	}
	buf = append(buf, []byte("]}")...)
	err := json.Unmarshal(buf, c)

	return c, err
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

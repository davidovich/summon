package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/summon"
	"github.com/gobuffalo/packr/v2"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestRunCmd(t *testing.T) {
	box := packr.New("test box", "testdata/plain")

	testCases := []struct {
		desc      string
		out       string
		args      []string
		main      *mainCmd
		wantError bool
		noCalls   bool
	}{
		{
			desc: "sub-command",
			args: []string{"echo"},
			out:  "bash echo hello",
		},
		{
			desc:      "no-sub-command",
			wantError: true,
		},
		{
			desc:      "invalid-sub-command",
			args:      []string{"ec"},
			wantError: true,
		},
		{
			desc: "sub-param-passed",
			args: []string{"echo", "--unknown-arg", "last", "params"},
			main: &mainCmd{
				json: "{\"Name\": \"david\"}",
			},
			// this relies on the v0.10.0 version of templated exec
			// see the echo command in testdata/summon.config.yaml
			out:       "bash echo hello david --unknown-arg last params",
			wantError: false,
		},
		{
			desc:      "dry-run",
			args:      []string{"echo", "-n"},
			wantError: false,
			noCalls:   true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			s, _ := summon.New(box)
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			execCommand := testutil.FakeExecCommand("TestSummonRunHelper", stdout, stderr)

			s.Configure(summon.ExecCmd(execCommand))

			if tC.main == nil {
				tC.main = &mainCmd{}
			}
			injectOsArgs := append([]string{"summon", "run"}, tC.args...)
			tC.main.osArgs = &injectOsArgs
			cmd := newRunCmd(false, nil, s, tC.main)
			cmd.SetArgs(tC.args)

			err := cmd.Execute()
			if tC.wantError {
				assert.Error(t, err)
				return
			}

			c, err := testutil.GetCalls(stderr)
			assert.Nil(t, err)
			if tC.noCalls {
				assert.Len(t, c.Calls, 0)
			} else {
				assert.Contains(t, c.Calls[0].Args, tC.out)
			}
		})
	}
}

func TestSummonRunHelper(t *testing.T) {
	if testutil.IsHelper() {
		defer os.Exit(0)

		call := testutil.MakeCall()

		testutil.WriteCall(call)
	}
}

func TestExtractUnknownArgs(t *testing.T) {
	fset := pflag.NewFlagSet("test", pflag.ContinueOnError)

	json := ""
	fset.StringVarP(&json, "json", "j", "{}", "")

	unknown := extractUnknownArgs(fset, []string{"--json", "{}", "--unknown"})
	assert.Equal(t, []string{"--unknown"}, unknown)

	unknown = extractUnknownArgs(fset, []string{"--"})
	assert.Equal(t, []string{"--"}, unknown)

	unknownShort := extractUnknownArgs(fset, []string{"-j", "--unknown"})
	assert.Equal(t, []string{"--unknown"}, unknownShort)

	unknownShort = extractUnknownArgs(fset, []string{"-"})
	assert.Equal(t, []string{"-"}, unknownShort)
}

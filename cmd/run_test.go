package cmd

import (
	"bytes"
	"embed"
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/summon"
)

//go:embed testdata/*
var runCmdTestFS embed.FS

func TestRunCmd(t *testing.T) {

	testCases := []struct {
		desc             string
		callArgsFragment string
		args             []string
		main             *mainCmd
		wantError        bool
		noCalls          bool
	}{
		{
			desc:             "sub-command",
			args:             []string{"run", "echo"},
			callArgsFragment: "bash echo hello",
		},
		{
			desc:      "no-sub-command",
			wantError: true,
		},
		{
			desc:      "invalid-sub-command",
			args:      []string{"run", "ec"},
			wantError: true,
		},
		{
			desc: "sub-param-passed",
			args: []string{"run", "echo", "--unknown-arg", "last", "params"},
			main: &mainCmd{
				json: "{\"Name\": \"david\"}",
			},
			// this relies on the v0.10.0 version of templated exec
			// see the echo command in testdata/summon.config.yaml
			callArgsFragment: "bash echo hello david --unknown-arg last params",
		},
		{
			desc:    "dry-run",
			args:    []string{"run", "echo", "-n"},
			noCalls: true,
		},
		{
			desc:             "run-completion",
			args:             []string{"__complete", "run", "tk", ""},
			callArgsFragment: "a\nb\n",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			s, _ := summon.New(runCmdTestFS)
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			execCommand := testutil.FakeExecCommand("TestSummonRunHelper", stdout, stderr)

			err := s.Configure(summon.ExecCmd(execCommand))
			assert.NoError(t, err)

			if tC.main == nil {
				tC.main = &mainCmd{}
			}
			injectOsArgs := append([]string{"summon"}, tC.args...)
			tC.main.osArgs = &injectOsArgs
			cobraCmd := &cobra.Command{Use: "summon", RunE: func(cmd *cobra.Command, args []string) error {
				return fmt.Errorf("root cmd called")
			}}
			// make sure we dont pass a nil slice to cobra, as this is the
			// zero value. Cobra uses os.Args if args are nil.
			// https://stackoverflow.com/a/44305910/28275
			if tC.args == nil {
				tC.args = make([]string, 0)
			}
			cobraCmd.SetArgs(tC.args)
			err = newRunCmd(true, cobraCmd, s, tC.main)
			assert.NoError(t, err)

			err = cobraCmd.Execute()
			if tC.wantError {
				assert.Error(t, err, "should have generated error: args: %v", tC.args)
				return
			}
			require.NoError(t, err)

			c, err := testutil.GetCalls(stderr)
			assert.Nil(t, err)
			if tC.noCalls {
				assert.Len(t, c.Calls, 0)
			} else {
				require.Greater(t, len(c.Calls), 0, "should have made call")
				assert.Contains(t, c.Calls[0].Args, tC.callArgsFragment)
			}
		})
	}
}

func TestSummonRunHelper(t *testing.T) {
	testutil.TestSummonRunHelper()
}

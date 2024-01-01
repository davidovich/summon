package cmd

import (
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/summon"
)

//go:embed testdata/*
var cmdTestFS embed.FS

func TestRunCmd(t *testing.T) {

	testCases := []struct {
		desc      string
		callArgs  []string
		args      []string
		wantError bool
		noCalls   bool
	}{
		{
			desc:     "sub-command",
			args:     []string{"run", "echo"},
			callArgs: []string{"bash", "echo", "hello "},
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
			desc:      "help-flag-on-root-works",
			args:      []string{"--help"},
			wantError: false,
			noCalls:   true,
		},
		{
			desc: "sub-param-passed",
			args: []string{"run", "--json", "{\"Name\": \"david\"}", "echo", "--unknown-arg", "last", "params"},
			// this relies on the v0.10.0 version of templated exec
			// see the echo command in testdata/summon.config.yaml
			callArgs: []string{"bash", "echo", "hello david", "--unknown-arg", "last", "params"},
		},
		{
			desc:    "dry-run",
			args:    []string{"run", "-n", "echo", "hello"},
			noCalls: true,
		},
		{
			desc: "run-completion",
			args: []string{"__complete", "run", "tk", ""},
			// Note that this test completion has an internal run command to output
			// a \n separated string (fake-make), this explains the leading call aags
			callArgs: []string{"bash", "--norc", "--noprofile", "-c", "echo -e a\nb\n"},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			s, _ := summon.New(cmdTestFS)
			execCommand := testutil.FakeExecCommand("TestSummonRunHelper")

			err := s.Configure(summon.ExecCmd(execCommand.Fn))
			assert.NoError(t, err)

			injectOsArgs := append([]string{"summon"}, tC.args...)
			cobraCmd, err := CreateRootCmd(s, injectOsArgs, summon.MainOptions{})
			require.NoError(t, err)

			err = cobraCmd.Execute()
			if tC.wantError {
				assert.Error(t, err, "should have generated error: args: %v", tC.args)
				return
			}
			require.NoError(t, err)

			c := execCommand.GetCalls()
			if tC.noCalls {
				assert.Len(t, c, 0)
			} else {
				require.Greater(t, len(c), 0, "should have made call")
				assert.Equal(t, tC.callArgs, c[0].Args)
			}
		})
	}
}

func TestSummonRunHelper(t *testing.T) {
	testutil.TestSummonRunHelper()
}

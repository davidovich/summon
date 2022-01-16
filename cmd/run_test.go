package cmd

import (
	"bytes"
	"embed"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/summon"
)

//go:embed testdata/*
var runCmdTestFS embed.FS

func TestRunCmd(t *testing.T) {

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
		{
			desc:      "run-completion",
			args:      []string{"__complete", "tk", ""},
			wantError: false,
			out:       "a\nb\n",
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
			injectOsArgs := append([]string{"summon", "run"}, tC.args...)
			tC.main.osArgs = &injectOsArgs
			cmd, _ := newRunCmd(true, nil, s, tC.main)
			cmd.SetArgs(tC.args)

			err = cmd.Execute()
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
	testutil.TestSummonRunHelper()
}

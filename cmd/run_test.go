package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/summon"
	"github.com/gobuffalo/packr/v2"
	"github.com/stretchr/testify/assert"
)

func TestRunCmd(t *testing.T) {
	box := packr.New("test box", "testdata")

	testCases := []struct {
		desc	string
		out string
		args []string
		wantError bool
	}{
		{
			desc: "sub-command",
			args: []string{"echo"},
			out: "bash echo hello",
		},
		{
			desc: "no-sub-command",
			wantError: true,
		},
		{
			desc: "invalid-sub-command",
			args: []string{"ec"},
			wantError: true,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			s, _ := summon.New(box)
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			execCommand := testutil.FakeExecCommand("TestSummonRunHelper", stdout, stderr)

			s.Configure(summon.ExecCmd(execCommand))

			cmd := newRunCmd(s)
			cmd.SetArgs(tC.args)

			err := cmd.Execute()
			if tC.wantError {
				assert.Error(t, err)
				return
			}

			c, err := testutil.GetCalls(stderr)
			assert.Nil(t, err)
			assert.Contains(t, c.Calls[0].Args, tC.out)
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

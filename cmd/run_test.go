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
	s, _ := summon.New(box)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	execCommand := testutil.FakeExecCommand("TestSummonRunHelper", stdout, stderr)

	s.Configure(summon.ExecCmd(execCommand))
	cmd := newRunCmd(s)
	cmd.SetArgs([]string{"echo"})

	cmd.Execute()

	c, err := testutil.GetCalls(stderr)
	assert.Nil(t, err)

	assert.Contains(t, c.Calls[0].Args, "bash echo hello")
}

func TestSummonRunHelper(t *testing.T) {
	if testutil.IsHelper() {
		defer os.Exit(0)

		call := testutil.MakeCall()

		testutil.WriteCall(call)
	}
}

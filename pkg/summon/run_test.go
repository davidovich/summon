package summon

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/command"
	"github.com/gobuffalo/packr/v2"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	stdout := &bytes.Buffer{}
	execCommand = testutil.FakeExecCommand("TestSummonRunHelper", stdout, nil)
	defer func() { execCommand = command.New }()

	box := packr.New("test run box", "testdata")

	s := New(box, Ref("hello"))
	err := s.Run()

	assert.Nil(t, err)
	assert.Equal(t, "python -c print(\"hello from python!\")", stdout.String())
}

func TestSummonRunHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := testutil.CleanHelperArgs(os.Args)
	fmt.Fprintf(os.Stdout, strings.Join(args, " "))
}

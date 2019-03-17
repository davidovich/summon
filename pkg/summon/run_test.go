package summon

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/command"
	"github.com/gobuffalo/packr/v2"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	defer func() { execCommand = command.New }()

	box := packr.New("test run box", "testdata")

	tests := []struct {
		name    string
		helper  string
		ref     string
		expect  string
		wantErr bool
	}{
		{
			name:    "composite-invoker", // python -c
			helper:  "TestSummonRunHelper",
			ref:     "hello",
			expect:  "python -c print(\"hello from python!\")",
			wantErr: false,
		},
		{
			name:    "simple-invoker", // bash
			helper:  "TestSummonRunHelper",
			ref:     "hello-bash",
			expect:  "bash hello.sh",
			wantErr: false,
		},
		{
			name:    "fail",
			ref:     "hello",
			helper:  "TestFailRunHelper",
			wantErr: true,
		},
		{
			name:    "fail-no-ref",
			ref:     "does-not-exist",
			helper:  "TestSummonRunHelper",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			execCommand = testutil.FakeExecCommand(tt.helper, stdout, nil)

			s, _ := New(box, Ref(tt.ref))
			if err := s.Run(); (err != nil) != tt.wantErr {
				t.Errorf("summon.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			c, err := testutil.GetCalls(stdout)
			assert.Nil(t, err)

			if tt.wantErr {
				assert.Len(t, c.Calls, 0)
			} else {
				assert.Equal(t, tt.expect, c.Calls[0].Args)
			}
		})
	}
}

func TestResolveExecUnit(t *testing.T) {

	testCases := []struct {
		desc        string
		execu       execUnit
		expected    execUnit
		wantsCalls  bool
		expectedArg string
		expectedEnv string
	}{
		{
			desc: "gobin",
			execu: execUnit{
				invoker: "gobin",
				target:  "github.com/myitcv/gobin@v0.0.8",
			},
			expected: execUnit{
				invoker: ".summoned/github.com/myitcv/gobin",
				target:  "",
				invOpts: []string{},
			},
			wantsCalls:  true,
			expectedArg: "gobin github.com/myitcv/gobin",
			expectedEnv: "GOBIN=.summoned/github.com/myitcv",
		},
		{
			desc: "non-gobin",
			execu: execUnit{
				invoker: "python",
				target:  "script.py",
			},
			expected: execUnit{
				invoker: "python",
				target:  "script.py",
			}},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			execCommand = testutil.FakeExecCommand("TestSummonRunHelper", stdout, nil)

			s, _ := New(packr.New("t", "testdata"), Dest(".summoned"))
			eu, err := s.resolve(tC.execu, []string{})

			assert.Nil(t, err)
			assert.Equal(t, tC.expected, eu)

			if tC.wantsCalls {
				c, err := testutil.GetCalls(stdout)
				assert.Nil(t, err)
				assert.Len(t, c.Calls, 1)
				assert.Contains(t, c.Calls[0].Env, tC.expectedEnv)
				assert.Contains(t, c.Calls[0].Args, tC.expectedArg)
			}
		})
	}
}

func TestFailRunHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	os.Exit(1)
}

func TestSummonRunHelper(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	call := testutil.Call{
		Args: strings.Join(testutil.CleanHelperArgs(os.Args), " "),
		Env:  os.Environ(),
	}

	testutil.WriteCall(call, os.Stdout)
}

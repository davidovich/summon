package summon

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/gobuffalo/packr/v2"
	"github.com/stretchr/testify/assert"

	"github.com/davidovich/summon/internal/testutil"
)

func TestRun(t *testing.T) {
	box := packr.New("test run box", "testdata")

	tests := []struct {
		name     string
		helper   string
		ref      string
		expect   string
		contains string
		args     []string
		wantErr  bool
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
			name:    "self-reference-invoker", // bash
			helper:  "TestSummonRunHelper",
			ref:     "bash-self-ref",
			expect:  fmt.Sprintf("bash %s", filepath.Join(os.TempDir(), "hello.sh")),
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
		{
			name:    "renderable invoker",
			helper:  "TestSummonRunHelper",
			ref:     "docker",
			expect:  "docker info",
			wantErr: false,
		},
		{
			name:    "args access",
			helper:  "TestSummonRunHelper",
			ref:     "args",
			args:    []string{"a", "b"},
			expect:  "bash args: [a b]",
			wantErr: false,
		},
		{
			name:     "osArgs access",
			helper:   "TestSummonRunHelper",
			ref:      "osArgs",
			contains: "test",
			wantErr:  false,
		},
		{
			name:     "global template render",
			helper:   "TestSummonRunHelper",
			ref:      "templateref",
			contains: "bash 1.2.3",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := New(box, Ref(tt.ref))
			assert.Nil(t, err)

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			s.Configure(ExecCmd(testutil.FakeExecCommand(tt.helper, stdout, stderr)))

			if err := s.Run(Args(tt.args...)); (err != nil) != tt.wantErr {
				t.Errorf("summon.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			c, err := testutil.GetCalls(stderr)
			assert.Nil(t, err)

			if tt.wantErr {
				assert.Len(t, c.Calls, 0)
			} else if tt.expect != "" {
				assert.Equal(t, tt.expect, c.Calls[0].Args)
			} else if tt.contains != "" {
				assert.Contains(t, c.Calls[0].Args, tt.contains)
			}
		})
	}
}

func TestFailRunHelper(t *testing.T) {
	if testutil.IsHelper() {
		os.Exit(1)
	}
}

func TestSummonRunHelper(t *testing.T) {
	if testutil.IsHelper() {
		defer os.Exit(0)

		call := testutil.MakeCall()

		testutil.WriteCall(call)
	}
}

func TestListInvocables(t *testing.T) {
	box := packr.New("test run box", "testdata")

	s, _ := New(box)

	inv := s.ListInvocables()
	assert.ElementsMatch(t, []string{"hello-bash", "bash-self-ref", "docker", "gobin", "gohack", "hello", "args", "osArgs", "templateref"}, inv)
}

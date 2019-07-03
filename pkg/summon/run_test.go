package summon

import (
	"bytes"
	"os"
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
			stderr := &bytes.Buffer{}
			execCommand = testutil.FakeExecCommand(tt.helper, stdout, stderr)

			s, _ := New(box, Ref(tt.ref))
			if err := s.Run(); (err != nil) != tt.wantErr {
				t.Errorf("summon.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			c, err := testutil.GetCalls(stderr)
			assert.Nil(t, err)

			if tt.wantErr {
				assert.Len(t, c.Calls, 0)
			} else {
				assert.Equal(t, tt.expect, c.Calls[0].Args)
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
	assert.ElementsMatch(t, []string{"hello-bash", "gobin", "gohack", "hello"}, inv)
}

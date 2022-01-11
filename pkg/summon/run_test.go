package summon

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/config"
)

func TestRun(t *testing.T) {

	tests := []struct {
		name     string
		helper   string
		ref      string
		expect   []string
		contains []string
		args     []string
		wantErr  bool
	}{
		{
			name:    "composite-invoker", // python -c
			helper:  "TestSummonRunHelper",
			ref:     "hello",
			expect:  []string{"python -c print(\"hello from python!\")"},
			wantErr: false,
		},
		{
			name:    "simple-invoker", // bash
			helper:  "TestSummonRunHelper",
			ref:     "hello-bash",
			expect:  []string{"bash hello.sh"},
			wantErr: false,
		},
		{
			name:    "self-reference-invoker", // bash
			helper:  "TestSummonRunHelper",
			ref:     "bash-self-ref",
			expect:  []string{fmt.Sprintf("bash %s", filepath.Join(os.TempDir(), "hello.sh"))},
			wantErr: false,
		},
		{
			name:   "self-reference-run", // bash
			helper: "TestSubCommandTemplateRunCall",
			ref:    "run-example",
			expect: []string{
				"bash hello.sh",          // run first call (returns "hello from subcmd")
				"bash hello from subcmd", // actual run-example call with args
			},
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
			name:    "renderable-invoker",
			helper:  "TestSummonRunHelper",
			ref:     "docker",
			expect:  []string{"docker info"},
			wantErr: false,
		},
		{
			name:    "args-access",
			helper:  "TestSummonRunHelper",
			ref:     "args",
			args:    []string{"a c", "b"},
			expect:  []string{"bash args: a c b"},
			wantErr: false,
		},
		{
			name:    "one-arg-access-remainder-passed",
			helper:  "TestSummonRunHelper",
			ref:     "one-arg",
			args:    []string{"\"acce ssed\"", "remainder1", "remainder2"},
			expect:  []string{"bash args: \"acce ssed\" remainder1 remainder2"},
			wantErr: false,
		},
		{
			name:    "all-args-access-no-remainder-passed",
			helper:  "TestSummonRunHelper",
			ref:     "all-args",
			args:    []string{"a", "b", "c", "d"},
			expect:  []string{"bash args: a b c d"},
			wantErr: false,
		},
		{
			name:     "osArgs-access",
			helper:   "TestSummonRunHelper",
			ref:      "osArgs",
			contains: []string{"test"},
			wantErr:  false,
		},
		{
			name:     "global-template-render",
			helper:   "TestSummonRunHelper",
			ref:      "templateref",
			contains: []string{"bash 1.2.3"},
			wantErr:  false,
		},
		{
			name:   "run-with-env",
			helper: "TestCommandWithEnvHelper",
			ref:    "run-with-env",
			expect: []string{
				"bash hello.sh",                    // run first call
				"bash env=HELLO=summon subcommand", // actual run-example call with args
			},
		},
		{
			name:   "new-cmd-spec",
			helper: "TestSummonRunHelper",
			ref:    "overrides",
			expect: []string{"bash hello.sh"},
		},
		{
			name:   "new-cmd-spec-subcmd",
			helper: "TestSummonRunHelper",
			ref:    "overrides",
			args:   []string{"subcmd", "another"},
			expect: []string{"bash hello.sh subcmd another"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := New(summonTestFS, Ref(tt.ref))
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
			} else {
				if len(tt.expect) != 0 {
					for i, e := range tt.expect {
						assert.Equal(t, e, c.Calls[i].Args)
					}
				}
				if len(tt.contains) != 0 {
					for i, e := range tt.contains {
						assert.Contains(t, c.Calls[i].Args, e)
					}
				}
			}
		})
	}
}

func TestFailRunHelper(t *testing.T) {
	testutil.TestFailRunHelper()
}

func TestSummonRunHelper(t *testing.T) {
	testutil.TestSummonRunHelper()
}

func TestSubCommandTemplateRunCall(t *testing.T) {
	if testutil.IsHelper() {
		defer os.Exit(0)
		testutil.WriteCall(testutil.MakeCall())

		fmt.Fprint(os.Stdout, "hello from subcmd")
	}
}

func TestCommandWithEnvHelper(t *testing.T) {
	if testutil.IsHelper() {
		defer os.Exit(0)
		testutil.WriteCall(testutil.MakeCall())

		hello := os.Getenv("HELLO")
		fmt.Fprintf(os.Stdout, "env=HELLO=%s subcommand", hello)
	}
}

func TestListInvocables(t *testing.T) {
	configFile := `
version: 1

exec:
  flags:
    config-root: 'CONFIG_ROOT=.'

  invokers:
    echo:
      echo-pwd: ['pwd:', '{{ env "PWD" | base }}']

    docker:
      manifest:
        help: 'render kubernetes manifests in build dir'
        # popArg is used to remove the arg from user input
        cmdArgs: ['manifests/{{ popArg 0 "manifest"}}','{{ template "parseArgs" 1 }}']
        completion: '{{ summon "make list-environments" }}'
`

	testFs := fstest.MapFS{}
	testFs["summon.config.yaml"] = &fstest.MapFile{Data: []byte(configFile)}

	s, err := New(testFs)
	assert.NoError(t, err)

	flags, inv := s.ExecContext()
	handles := []string{}

	for h := range inv {
		handles = append(handles, h)
	}
	assert.ElementsMatch(t, []string{"echo-pwd", "manifest"}, handles)

	assert.IsType(t, config.ArgSliceSpec{}, inv["echo-pwd"].Value)
	assert.IsType(t, config.CmdSpec{}, inv["manifest"].Value)

	assert.Contains(t, flags, "config-root")
}

func TestFlattenStrings(t *testing.T) {
	tests := []struct {
		name string
		args []interface{}
		want []string
	}{
		{
			name: "simple-slice",
			args: []interface{}{"string1", "string2"},
			want: []string{"string1", "string2"},
		},
		{
			name: "empty",
			args: []interface{}{},
			want: []string{},
		},
		{
			name: "slice-of-slice-of-string",
			args: []interface{}{[]string{"elem"}},
			want: []string{"elem"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, FlattenStrings(tt.args...))
		})
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

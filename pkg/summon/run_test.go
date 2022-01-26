package summon

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/lithammer/dedent"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/command"
	"github.com/davidovich/summon/pkg/config"
)

func TestRun(t *testing.T) {

	tests := []struct {
		name      string
		helper    string
		cmd       []string
		enableRun bool
		expect    []string
		contains  []string
		args      []string
		wantErr   bool
	}{
		{
			name:    "composite-invoker", // python -c
			cmd:     []string{"hello"},
			expect:  []string{"python -c print(\"hello from python!\")"},
			wantErr: false,
		},
		{
			name:    "simple-invoker", // bash
			cmd:     []string{"hello-bash"},
			expect:  []string{"bash hello.sh"},
			wantErr: false,
		},
		{
			name:    "self-reference-invoker", // bash
			cmd:     []string{"bash-self-ref"},
			expect:  []string{fmt.Sprintf("bash %s", filepath.Join(os.TempDir(), "hello.sh"))},
			wantErr: false,
		},
		{
			name:   "self-reference-run", // bash
			helper: "TestSubCommandTemplateRunCall",
			cmd:    []string{"run-example"},
			expect: []string{
				"bash hello.sh",          // run first call (returns "hello from subcmd")
				"bash hello from subcmd", // actual run-example call with args
			},
			wantErr: false,
		},
		{
			name:    "fail",
			cmd:     []string{"hello"},
			helper:  "TestFailRunHelper",
			wantErr: true,
		},
		{
			name:    "fail-no-ref",
			cmd:     []string{"does-not-exist"},
			wantErr: true,
		},
		{
			name:    "renderable-invoker",
			cmd:     []string{"docker"},
			expect:  []string{"docker info"},
			wantErr: false,
		},
		{
			name:    "args-access",
			cmd:     []string{"args"},
			args:    []string{"a c", "b"},
			expect:  []string{"bash args: a c b"},
			wantErr: false,
		},
		{
			name:    "one-arg-access-remainder-passed",
			cmd:     []string{"one-arg"},
			args:    []string{"\"acce ssed\"", "remainder1", "remainder2"},
			expect:  []string{"bash args: \"acce ssed\" remainder1 remainder2"},
			wantErr: false,
		},
		{
			name:    "all-args-access-no-remainder-passed",
			cmd:     []string{"all-args"},
			args:    []string{"a", "b", "c", "d"},
			expect:  []string{"bash args: a b c d"},
			wantErr: false,
		},
		{
			name:     "osArgs-access",
			cmd:      []string{"osArgs"},
			contains: []string{"test"},
			wantErr:  false,
		},
		{
			cmd:      []string{"templateref"},
			contains: []string{"bash 1.2.3"},
			wantErr:  false,
		},
		{
			name:   "new-cmd-spec",
			cmd:    []string{"overrides"},
			expect: []string{"bash hello.sh"},
		},
		{
			name:   "new-cmd-spec-subcmd",
			cmd:    []string{"overrides"},
			args:   []string{"subcmd", "another"},
			expect: []string{"bash hello.sh subcmd another"},
		},
		{
			name:      "sub-cmd-with-run-enabled",
			enableRun: true,
			cmd:       []string{"run", "hello-bash"},
			expect:    []string{"bash hello.sh"},
		},
		{
			name:      "args-error-run-enabled",
			enableRun: true,
			cmd:       []string{"run"},
			wantErr:   true,
		},
		{
			name:      "args-invalid-error-run-enabled",
			enableRun: true,
			cmd:       []string{"run", "haha"},
			wantErr:   true,
		},
		{
			name:      "args-valid-run-enabled",
			enableRun: true,
			cmd:       []string{"run"},
			args:      []string{"hello-bash"},
		},
		{
			name: "with-dry-run",
			args: []string{"hello-bash", "-n"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := New(summonTestFS)
			assert.Nil(t, err)

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			if tt.helper == "" {
				tt.helper = "TestSummonRunHelper"
			}

			program := append([]string{"summon"}, tt.cmd...)
			args := append(program, tt.args...)
			s.Configure(ExecCmd(testutil.FakeExecCommand(tt.helper, stdout, stderr)), Args(args...))

			rootCmd := &cobra.Command{Use: "root", Run: func(cmd *cobra.Command, args []string) {}}
			s.ConstructCommandTree(rootCmd, tt.enableRun)

			if _, err := executeCommand(rootCmd); (err != nil) != tt.wantErr {
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

func TestListInvocables(t *testing.T) {
	configFile := `
version: 1

exec:
  flags:
    config-root: 'CONFIG_ROOT=.'

  environments:
    echo:
      echo-pwd: ['pwd:', '{{ env "PWD" | base }}']

    docker:
      manifest:
        help: 'render kubernetes manifests in build dir'
        # popArg is used to remove the arg from user input
        args: ['manifests/{{ arg 0 }}','{{ flag "config-root" }}']
        completion: '{{ summon "make list-environments" }}'
`

	testFs := fstest.MapFS{}
	testFs[config.ConfigFileName] = &fstest.MapFile{Data: []byte(configFile)}

	s, err := New(testFs)
	assert.NoError(t, err)

	flags, handles, err := s.execContext()
	assert.NoError(t, err)
	assert.Contains(t, handles, "echo-pwd")
	assert.Contains(t, handles, "manifest")

	assert.Equal(t,
		[]string{"pwd:", "{{ env \"PWD\" | base }}"},
		FlattenStrings(handles["echo-pwd"].args...))
	assert.Equal(t,
		[]string{"manifests/{{ arg 0 }}", `{{ flag "config-root" }}`},
		FlattenStrings(handles["manifest"].args))

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

type noopWriter struct{}

func (*noopWriter) Write(p []byte) (int, error) { return len(p), nil }

func executeCommand(root *cobra.Command) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(&noopWriter{})

	_, err = root.ExecuteC()

	return buf.String(), err
}

func TestConstructCommandTree(t *testing.T) {
	configFile := dedent.Dedent(`
		version: 1

		exec:
		  flags:
		    config-root: 'CONFIG_ROOT=.'

		  environments:
		    docker:
		      manifest:
		        help: 'render kubernetes manifests in build dir'
		        subCmd:
		          all: [all subcmd]
		        args: ['manifests{{ if args }}/{{arg 0 "manifest"}}{{end}}']
		        completion: 'a-completion'
		      simple: [hello]
		`)

	tests := []struct {
		name           string
		cmd            []string
		withRunCmd     bool
		expectSubArgs  []string
		expectError    bool
		expectComplete []string
	}{
		{
			name:           "without-run",
			cmd:            []string{"manifest"},
			expectComplete: []string{"all", "a-completion"},
		},
		{
			name: "without-run-all",
			cmd:  []string{"manifest", "all"},
		},
		{
			name:           "with-run",
			withRunCmd:     true,
			cmd:            []string{"run", "manifest"},
			expectComplete: []string{"all", "a-completion"},
		},
		{
			name:       "with-run-manifest-all",
			withRunCmd: true,
			cmd:        []string{"run", "manifest", "all"},
		},
		{
			name:           "sub-arg-not-found",
			cmd:            []string{"manifest", "none"},
			expectSubArgs:  []string{"none"},
			expectComplete: []string{"a-completion"},
		},
		{
			name:          "arg-not-found",
			cmd:           []string{"man"},
			expectError:   true,
			expectSubArgs: []string{"man"},
		},
		{
			name:           "only-run",
			cmd:            []string{"run"},
			withRunCmd:     true,
			expectComplete: []string{"manifest", "simple"},
		},
	}
	testFs := fstest.MapFS{}
	testFs[config.ConfigFileName] = &fstest.MapFile{Data: []byte(configFile)}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectSubArgs == nil {
				tt.expectSubArgs = []string{}
			}
			s, err := New(testFs, DryRun(true), Args(append([]string{"prog"}, tt.cmd...)...))
			assert.NoError(t, err)

			rootCmd := cobra.Command{Use: "root", Run: func(cmd *cobra.Command, args []string) {}}
			err = s.ConstructCommandTree(&rootCmd, tt.withRunCmd)
			assert.NoError(t, err)
			if tt.withRunCmd {
				cmd, _, err := rootCmd.Find([]string{"run"})
				assert.NoError(t, err)
				assert.Equal(t, cmd.Use, "run [handle]")
			}

			foundCmd, subArgs, err := rootCmd.Find(tt.cmd)
			if !tt.expectError {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
			assert.NotNil(t, foundCmd)
			assert.Equal(t, tt.expectSubArgs, subArgs)

			// check completion
			completeArgs := []string{cobra.ShellCompNoDescRequestCmd}
			rootCmd.SetArgs(append(completeArgs, append(tt.cmd, "")...))
			out, err := executeCommand(&rootCmd)
			assert.NoError(t, err)
			completeSlice := strings.Split(out, "\n")

			// remove cobra completion control chars
			completeSlice = completeSlice[:len(completeSlice)-2]
			assert.ElementsMatch(t, completeSlice, tt.expectComplete, "computed complete: %v, expected: %v", completeSlice, tt.expectComplete)
		})
	}
}

func TestDuplicateHandles(t *testing.T) {
	configFile := dedent.Dedent(`
		exec:
		  environments:
		    docker:
		      manifest: ['manifests{{ if args }}/{{arg 0 "manifest"}}{{end}}']
		    bash:
		      manifest: [hello]
		`)
	testFs := fstest.MapFS{}
	testFs[config.ConfigFileName] = &fstest.MapFile{Data: []byte(configFile)}

	_, err := New(testFs)
	if assert.Error(t, err) {
		assert.Contains(t, err.Error(),
			fmt.Sprintf("in config %s: cannot have duplicate handles: '%s'", config.ConfigFileName, "manifest"))
	}
}

type flagTest struct {
	name string
	config.Flags
	globalFlags           config.Flags
	cmdSpec               *commandSpec
	userInvocation        []string
	expected              []string
	args                  []string
	withRun               bool
	ensureCobraHelpCalled bool
}

func (ft flagTest) run(t *testing.T) {
	d := Driver{
		configRead: true, // disable config read
		cmdToSpec:  map[*cobra.Command]*commandSpec{},
	}

	cmdSpec := ft.cmdSpec
	if cmdSpec == nil {
		cmdSpec = &commandSpec{
			args:  []interface{}{ft.args},
			flags: ft.Flags,
		}
	}
	cmdSpec.execEnvironment = "program"

	d.globalFlags = ft.globalFlags
	d.handles = handles{"a-handle": cmdSpec}

	programArgs := []string{cmdSpec.execEnvironment}
	if ft.withRun {
		programArgs = append(programArgs, "run")
	}
	programArgs = append(programArgs, "a-handle")
	d.Configure(ExecCmd(func(cmd string, args ...string) *command.Cmd {
		return &command.Cmd{
			Cmd: &exec.Cmd{},
			Run: func() error {
				assert.Equal(t, ft.expected, args)
				return nil
			},
		}
	}), Args(append(programArgs, ft.userInvocation...)...))

	rootCmd := &cobra.Command{Use: "root"}
	helpCalled := false
	rootCmd.SetHelpFunc(func(c *cobra.Command, s []string) {
		helpCalled = true
	})

	err := d.ConstructCommandTree(rootCmd, ft.withRun)
	assert.NoError(t, err)

	_, err = executeCommand(rootCmd)
	assert.NoError(t, err)

	assert.Equal(t, ft.ensureCobraHelpCalled, helpCalled, "cobra help was not called")
}

func TestFlagUsages(t *testing.T) {
	tests := []flagTest{
		{
			name: "happy",
			Flags: config.Flags{
				"a-flag": &config.FlagSpec{
					Effect: "TEMPLATE={{.flag}}",
				},
			},
			userInvocation: []string{"--a-flag", "user-value"},
			expected:       []string{"TEMPLATE=user-value"},
		},
		{
			name: "global",
			globalFlags: config.Flags{
				"global": &config.FlagSpec{
					Effect: "GLOBALFLAGVALUE={{.flag}}",
				},
			},
			userInvocation: []string{"--global", "global-value"},
			expected:       []string{"GLOBALFLAGVALUE=global-value"},
		},
		{
			name: "multiple-flags-in-order",
			Flags: config.Flags{
				"one": &config.FlagSpec{
					Effect: "one={{.flag}}",
				},
				"two": &config.FlagSpec{
					Effect: "two={{.flag}}",
				},
				"three": &config.FlagSpec{
					Effect: "three={{.flag}}",
				},
			},
			userInvocation: []string{"--one", "1", "--two", "2", "--three", "3"},
			expected:       []string{"one=1", "two=2", "three=3"},
		},
		{
			name: "global-and-local",
			Flags: config.Flags{
				"one": &config.FlagSpec{
					Effect: "one={{.flag}}",
				},
			},
			globalFlags: config.Flags{
				"global": &config.FlagSpec{
					Effect: "GLOBALFLAGVALUE={{.flag}}",
				},
			},
			userInvocation: []string{"--one", "1", "--global", "global"},
			expected:       []string{"one=1", "GLOBALFLAGVALUE=global"},
		},
		{
			name: "interspersed",
			Flags: config.Flags{
				"one": &config.FlagSpec{
					Effect: "one={{.flag}}",
				},
			},
			globalFlags: config.Flags{
				"global": &config.FlagSpec{
					Effect: "GLOBALFLAGVALUE={{.flag}}",
				},
			},
			userInvocation: []string{"--one", "1", "another", "--global", "global"},
			expected:       []string{"one=1", "GLOBALFLAGVALUE=global", "another"},
		},
		{
			name: "non-used-flags",
			Flags: config.Flags{
				"one": &config.FlagSpec{
					Effect: "one={{.flag}}",
				},
			},
			userInvocation: []string{"a-arg"},
			expected:       []string{"a-arg"},
		},
		{
			name: "flags-not-duplicated-not-reordered",
			args: []string{"{{ arg 0 }}", "subcmd", `{{ flagValue "one" }}`, "anotherSubCmd"},
			Flags: config.Flags{
				"one": &config.FlagSpec{
					Effect: "one={{.flag}}",
				},
			},
			userInvocation: []string{"a-arg", "--one", "1"},
			expected:       []string{"a-arg", "subcmd", "one=1", "anotherSubCmd"},
		},
		{
			name: "default-values-for-flags",
			Flags: config.Flags{
				"number": &config.FlagSpec{
					Effect:  "number={{.flag}}",
					Default: "1234",
				},
			},
			userInvocation: []string{"--number"},
			expected:       []string{"number=1234"},
		},
		{
			name:     "non-existing-reference-should-not-consume-arg-pos",
			args:     []string{`{{flagValue "inexistant"}}`, "arg"},
			expected: []string{"arg"},
		},
		{
			name:     "empty-array-used-to-insert-empty-arg-pos",
			args:     []string{`[{{flagValue "inexistant"}}]`, "arg"},
			expected: []string{"", "arg"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, test.run)
	}
}

func TestFlagUsages2(t *testing.T) {

	configFile := dedent.Dedent(`
		exec:
		  flags:
		    global-flag:
		      effect: 'global-flag-set'
		      explicit: true
		  environments:
		    program:
		      a-command:
		        args: [-c]
		        flags:
		          user-flag:
		            effect: 'CONVERTED={{.flag}}'
		            help: user-flag allows user to flag something to the callee
		            shorthand: u
		      b-cmd:
		        args: [b-cmd, '{{flagValue "global-flag"}}']
		`)
	testFs := fstest.MapFS{}
	testFs[config.ConfigFileName] = &fstest.MapFile{Data: []byte(configFile)}

	makeDriver := func(expected ...string) *Driver {
		s, err := New(testFs, ExecCmd(func(s1 string, s2 ...string) *command.Cmd {
			return &command.Cmd{
				Cmd: &exec.Cmd{},
				Run: func() error {
					assert.Equal(t, expected, append([]string{s1}, s2...))
					return nil
				},
			}
		}))
		assert.NoError(t, err)
		return s
	}

	t.Run("simple", func(t *testing.T) {
		s := makeDriver("program", "-c", "CONVERTED=user-value")

		rootCmd := &cobra.Command{Use: "root"}
		s.Configure(Args("prog", "a-command", "--user-flag", "user-value"))
		err := s.ConstructCommandTree(rootCmd, false)
		assert.NoError(t, err)

		_, err = executeCommand(rootCmd)
		assert.NoError(t, err)
	})

	t.Run("argSliceSpec", func(t *testing.T) {
		s := makeDriver("program", "b-cmd", "global-flag-set", "arg1", "arg2")
		require.NotNil(t, s)

		rootCmd := &cobra.Command{Use: "root"}
		s.Configure(Args("prog", "b-cmd", "arg1", "arg2", "--global-flag", "user-value"))
		err := s.ConstructCommandTree(rootCmd, false)
		assert.NoError(t, err)

		s.Configure(Args("prog", "arg1", "arg2"))
		_, err = executeCommand(rootCmd)
		assert.NoError(t, err)
	})
}

func TestHelpManagement(t *testing.T) {
	tests := []flagTest{
		{
			name:           "no-defined-help-no-cobra-managed-help",
			userInvocation: []string{"--help", "user-value"},
			expected:       []string{"--help", "user-value"},
			withRun:        true,
		},
		{
			name:                  "cobra-managed-help-on-handle-command",
			userInvocation:        []string{"--help", "user-value"},
			cmdSpec:               &commandSpec{help: "user defined help"},
			expected:              []string{""},
			withRun:               true,
			ensureCobraHelpCalled: true,
		},
		{
			name:           "proxy-managed-help",
			userInvocation: []string{"proxy-sub-command", "--help"},
			expected:       []string{"proxy-sub-command", "--help"},
			withRun:        true,
		},
		{
			name:                  "proxy-interspersed-help",
			userInvocation:        []string{"proxy-command", "subcommand", "--help", "arg"},
			expected:              []string{"proxy-command", "subcommand", "--help", "arg"},
			withRun:               true,
			ensureCobraHelpCalled: false,
		},
		{
			name: "help-used-in-flagValue",
			Flags: config.Flags{
				"one": &config.FlagSpec{
					Effect:  "{{.flag}}",
					Default: "--one",
				},
			},
			args:           []string{"proxy-command", `{{ flagValue "help"}}`, `{{ flagValue "one" }}{{ $swallowargs := args }}`},
			userInvocation: []string{"proxy-command", "--one", "subcommand", "--help", "arg"},
			expected:       []string{"proxy-command", "--help", "--one"},
			withRun:        true,
		},
		{
			name: "help-used-in-flagValue-arg-consumed",
			Flags: config.Flags{
				"one": &config.FlagSpec{
					Effect:  "{{.flag}}",
					Default: "--one",
				},
			},
			args:           []string{"{{ arg 0 }}", `{{ flagValue "help"}}`, `{{ flagValue "one" }}`},
			userInvocation: []string{"proxy-command", "--one", "subcommand", "--help", "arg"},
			expected:       []string{"proxy-command", "--help", "--one", "subcommand", "arg"},
			withRun:        false,
		},
		{
			name: "help-on-user-defined-help-is-cobra-managed",
			cmdSpec: &commandSpec{
				subCmd: map[string]*commandSpec{"sub-command": {
					execEnvironment: "program",
					help:            "sub command help",
				}},
			},
			userInvocation:        []string{"sub-command", "--help"},
			ensureCobraHelpCalled: true,
		},
		{
			name: "help-on-non-user-defined-help-is-passed",
			cmdSpec: &commandSpec{
				subCmd: map[string]*commandSpec{"sub-command": {
					execEnvironment: "program",
				}},
			},
			userInvocation: []string{"sub-command", "--help"},
			expected:       []string{"--help"},
		},
		{
			name: "no-cobra-help-on-user-defined-cmd-with-no-help",
			cmdSpec: &commandSpec{
				subCmd: map[string]*commandSpec{"sub-command": {
					execEnvironment: "program",
					subCmd: map[string]*commandSpec{"sub-sub-cmd": {
						args:            config.ArgSliceSpec{[]string{"sub-command", "sub-sub-cmd"}},
						execEnvironment: "program",
					}},
				}},
			},
			userInvocation: []string{"sub-command", "sub-sub-cmd", "--help"},
			expected:       []string{"sub-command", "sub-sub-cmd", "--help"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, test.run)
	}
}

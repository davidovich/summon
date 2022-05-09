package cmd

import (
	"bytes"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/config"
	"github.com/davidovich/summon/pkg/summon"
)

func makeRootCmd(withoutRun bool, args ...string) (*summon.Driver, *cobra.Command) {
	s, _ := summon.New(cmdTestFS)
	rootCmd, _ := CreateRootCmd(s, append([]string{"summon"}, args...), summon.MainOptions{WithoutRunSubcmd: withoutRun})
	return s, rootCmd
}

func Test_createRootCmd(t *testing.T) {
	defer testutil.ReplaceFs()()

	mockBuildInfo := func(bInfo func() (*debug.BuildInfo, bool)) func() func() {
		return func() func() {
			oldBi := buildInfo
			buildInfo = bInfo
			return func() {
				buildInfo = oldBi
			}
		}
	}

	makeRootCmd := func(args ...string) func(args ...string) *cobra.Command {
		return func(a ...string) *cobra.Command {
			_, c := makeRootCmd(false, args...)
			return c
		}
	}

	tests := []struct {
		name     string
		rootCmd  func(args ...string) *cobra.Command
		expected string
		in       string
		wantErr  bool
		defered  func() func()
	}{
		{
			name:    "no-args-no-all",
			rootCmd: makeRootCmd(),
			wantErr: true,
		},
		{
			name:    "all",
			rootCmd: makeRootCmd("-a"),
		},
		{
			name:     "file",
			rootCmd:  makeRootCmd("b.txt"),
			expected: "overridden_dir/b.txt",
		},
		{
			name:     "completion_run",
			rootCmd:  makeRootCmd("completion"),
			expected: "# bash completion V2 for summon",
		},
		{
			name:    "-v",
			rootCmd: makeRootCmd("-v"),
			expected: `"mod": "example.com/assets",
    "version": "v0.1.0"`, // note 4 spaces indent
			defered: mockBuildInfo(func() (*debug.BuildInfo, bool) {
				bi := &debug.BuildInfo{
					Main: debug.Module{
						Path:    "example.com/assets",
						Version: "v0.1.0",
					},
					Deps: []*debug.Module{
						{
							Path:    "github.com/davidovich/summon",
							Version: "(devel)",
						},
					},
				}
				return bi, true
			}),
		},
		{
			name:    "-v no-build-info",
			rootCmd: makeRootCmd("-v"),
			defered: mockBuildInfo(func() (*debug.BuildInfo, bool) { return nil, false }),
			wantErr: true,
		},
		{
			name:    "exclusive --json and --json-file",
			rootCmd: makeRootCmd("--json", "{}", "--json-file", "-", "summon.config.yaml"),
			wantErr: true,
		},
		{
			name:    "--json-file",
			rootCmd: makeRootCmd("--json-file", "testdata/json-for-template.json", "summon.config.yaml"),
			wantErr: false,
		},
		{
			name:    "--json-file non-existing",
			rootCmd: makeRootCmd("--json-file", "does-not-exist", "summon.config.yaml"),
			wantErr: true,
		},
		{
			name:    "--json-file stdin",
			rootCmd: makeRootCmd("--json-file", "-", "summon.config.yaml"),
			in:      "{}",
			wantErr: false,
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i)+"_"+tt.name, func(t *testing.T) {
			if tt.defered != nil {
				defer tt.defered()()
			}
			b := &bytes.Buffer{}
			rootCmd := tt.rootCmd()
			rootCmd.SetOut(b)
			if tt.in != "" {
				rootCmd.SetIn(strings.NewReader(tt.in))
			}
			if err := rootCmd.Execute(); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.expected != "" {
				assert.Contains(t, b.String(), tt.expected)
			}
		})
	}
}

func Test_RootCmdWithRunnables(t *testing.T) {

	tests := []struct {
		name         string
		args         []string
		expectedCall string
		wantErr      bool
	}{
		{
			name:         "call echo",
			args:         []string{"echo"},
			expectedCall: "bash echo hello",
			wantErr:      false,
		},
		{
			name:         "call hello-bash",
			args:         []string{"hello-bash"},
			expectedCall: "bash hello.sh",
			wantErr:      false,
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i)+"_"+tt.name, func(t *testing.T) {
			s, rootCmd := makeRootCmd(true, tt.args...)

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			execCommand := testutil.FakeExecCommand("TestSummonRunHelper", stdout, stderr)

			s.Configure(summon.ExecCmd(execCommand))

			if err := rootCmd.Execute(); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			c, err := testutil.GetCalls(stderr)
			assert.Nil(t, err)
			assert.Contains(t, c.Calls[0].Args, tt.expectedCall)
		})
	}
}

func Test_mainCmd_run(t *testing.T) {
	defer testutil.ReplaceFs()()

	type fields struct {
		copyAll  bool
		dest     string
		filename string
		driver   *summon.Driver
	}
	tests := []struct {
		name       string
		fields     fields
		out        string
		lsAsOption bool
		wantErr    bool
	}{
		{
			name: "base",
			fields: fields{
				dest:     ".s",
				filename: "a.txt",
				driver:   func() *summon.Driver { s, _ := summon.New(cmdTestFS); return s }(),
			},
			out: ".s/a.txt\n",
		},
		{
			name:       "ls-option",
			lsAsOption: true,
			fields: fields{
				driver: func() *summon.Driver { s, _ := summon.New(cmdTestFS); return s }(),
			},
			out: "a.txt\nb.txt\njson-for-template.json\nsummon.config.yaml\n",
		},
		{
			name: "copyAll",
			fields: fields{
				copyAll: true,
				dest:    ".s",
				driver:  func() *summon.Driver { s, _ := summon.New(cmdTestFS); return s }(),
			},
			out: ".s\n", // note dest dir
		},
		{
			name: "error",
			fields: fields{
				driver: nil,
			},
			wantErr: true,
			out:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mainCmd{
				driver:      tt.fields.driver,
				copyAll:     tt.fields.copyAll,
				dest:        tt.fields.dest,
				filename:    tt.fields.filename,
				listOptions: &listCmdOpts{asOption: tt.lsAsOption, driver: tt.fields.driver},
			}
			b := &bytes.Buffer{}
			m.out = b
			m.listOptions.out = b
			if err := m.run(); (err != nil) != tt.wantErr {
				t.Errorf("mainCmd.run() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.out, b.String())
		})
	}
}

func extractCommands(root *cobra.Command) (commands []string) {
	for _, c := range root.Commands() {
		commands = append(commands, c.Name())
	}
	return commands
}

func TestAssetsAreAlsoCommands(t *testing.T) {
	d, err := summon.New(cmdTestFS)
	assert.NoError(t, err)

	t.Run("assets-are-commands-with-run-cmd", func(t *testing.T) {
		root, err := CreateRootCmd(d, []string{"program"}, summon.MainOptions{})
		assert.NoError(t, err)

		commands := extractCommands(root)
		assert.Subset(t, commands, []string{"a.txt", "b.txt",
			"json-for-template.json", "summon.config.yaml"})
	})

	t.Run("assets-are-commands-without-run-cmd", func(t *testing.T) {
		root, err := CreateRootCmd(d, []string{"program"}, summon.MainOptions{WithoutRunSubcmd: true})
		assert.NoError(t, err)

		commands := extractCommands(root)
		assert.Subset(t, commands, []string{"a.txt", "b.txt",
			"json-for-template.json", "summon.config.yaml"})
	})
}

func TestThatAssetsAreNotCommandsIfConfiguredSo(t *testing.T) {
	configFile := `hideAssetsInHelp: true`

	testFs := fstest.MapFS{}
	testFs["assets/"+config.ConfigFileName] = &fstest.MapFile{Data: []byte(configFile)}

	d, err := summon.New(testFs)
	assert.NoError(t, err)

	t.Run("no-config-file-on-root", func(t *testing.T) {
		root, err := CreateRootCmd(d, []string{"program"}, summon.MainOptions{WithoutRunSubcmd: true})
		assert.NoError(t, err)

		commands := extractCommands(root)
		assert.ElementsMatch(t, []string{"completion"}, commands)
		assert.NotContains(t, commands, config.ConfigFileName)
	})
}

func TestThatAssetsAreNotCommandsAndPassedToRun(t *testing.T) {
	configFile := `hideAssetsInHelp: true
outputdir: "a"`

	testFs := fstest.MapFS{}
	testFs["assets/"+config.ConfigFileName] = &fstest.MapFile{Data: []byte(configFile)}
	testFs["assets/b.txt"] = &fstest.MapFile{Data: []byte("b content")}

	d, err := summon.New(testFs)
	assert.NoError(t, err)

	t.Run("test-json-exclusivity", func(t *testing.T) {
		root, err := CreateRootCmd(d, []string{"program", "--json", "{}", "--json-file", "-", "summon.config.yaml"}, summon.MainOptions{WithoutRunSubcmd: true})
		assert.NoError(t, err)

		err = root.Execute()
		assert.Error(t, err)
	})

	t.Run("file-as-arg", func(t *testing.T) {
		defer testutil.ReplaceFs()()

		root, err := CreateRootCmd(d, []string{"program", "b.txt"}, summon.MainOptions{WithoutRunSubcmd: true})
		assert.NoError(t, err)

		b := &bytes.Buffer{}
		root.SetOut(b)

		err = root.Execute()
		assert.NoError(t, err)

		assert.Contains(t, b.String(), "a/b.txt")
	})

	t.Run("completion", func(t *testing.T) {
		root, err := CreateRootCmd(d, []string{"program", cobra.ShellCompNoDescRequestCmd, ""}, summon.MainOptions{WithoutRunSubcmd: true})
		assert.NoError(t, err)

		b := &bytes.Buffer{}
		root.SetOut(b)

		err = root.Execute()
		assert.NoError(t, err)

		assert.Contains(t, b.String(), "b.txt")
	})
}

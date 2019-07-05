package cmd

import (
	"bytes"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/davidovich/summon/internal/testutil"
	"github.com/davidovich/summon/pkg/summon"
	"github.com/gobuffalo/packr/v2"
)

func Test_createRootCmd(t *testing.T) {
	defer testutil.ReplaceFs()()

	box := packr.New("test box", "testdata")
	box.AddString("a.txt", "a content")
	box.AddString("b.txt", "b content")

	mockBuildInfo := func() func() {
		oldBi := buildInfo
		buildInfo = func() (*debug.BuildInfo, bool) {
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
		}
		return func() {
			buildInfo = oldBi
		}
	}

	makeRootCmd := func(args ...string) *cobra.Command {
		s, _ := summon.New(box)
		rootCmd := CreateRootCmd(s, []string{"summon"})
		rootCmd.SetArgs(args)
		return rootCmd
	}

	tests := []struct {
		name     string
		rootCmd  *cobra.Command
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
			name:    "file",
			rootCmd: makeRootCmd("b.txt"),
		},
		{
			name:     "completion_run",
			rootCmd:  makeRootCmd("completion"),
			expected: "summon_run_hello",
		},
		{
			name:    "-v",
			rootCmd: makeRootCmd("-v"),
			expected: `"mod": "example.com/assets",
    "version": "v0.1.0"`, // note 4 spaces indent
			defered: mockBuildInfo,
		},
		{
			name:    "-v no-build-info",
			rootCmd: makeRootCmd("-v"),
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
			tt.rootCmd.SetOut(b)
			if tt.in != "" {
				tt.rootCmd.SetIn(strings.NewReader(tt.in))
			}
			if err := tt.rootCmd.Execute(); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.expected != "" {
				assert.Contains(t, b.String(), tt.expected)
			}
		})
	}
}

func Test_mainCmd_run(t *testing.T) {
	defer testutil.ReplaceFs()()

	box := packr.New("test box", "")
	box.AddString("a.txt", "a content")
	box.AddString("b.txt", "b content")

	type fields struct {
		copyAll  bool
		dest     string
		filename string
		driver   *summon.Driver
	}
	tests := []struct {
		name    string
		fields  fields
		out     string
		wantErr bool
	}{
		{
			name: "base",
			fields: fields{
				dest:     ".s",
				filename: "a.txt",
				driver:   func() *summon.Driver { s, _ := summon.New(box); return s }(),
			},
			out: ".s/a.txt\n",
		},
		{
			name: "copyAll",
			fields: fields{
				copyAll: true,
				dest:    ".s",
				driver:  func() *summon.Driver { s, _ := summon.New(box); return s }(),
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
				driver:   tt.fields.driver,
				copyAll:  tt.fields.copyAll,
				dest:     tt.fields.dest,
				filename: tt.fields.filename,
			}
			b := &bytes.Buffer{}
			m.out = b
			if err := m.run(); (err != nil) != tt.wantErr {
				t.Errorf("mainCmd.run() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, tt.out, b.String())
		})
	}
}

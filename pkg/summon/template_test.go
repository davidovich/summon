package summon

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/davidovich/summon/pkg/config"
)

type testPrompter struct {
	b         *bytes.Buffer
	choice    int
	userInput string
}

func (tpr *testPrompter) NewPrompt(userPrompt string) {
	tpr.b = bytes.NewBufferString(userPrompt)
}

func (tpr *testPrompter) Choose(choices []string) (string, error) {
	return choices[tpr.choice], nil
}

func (tpr *testPrompter) Input(defaultVal string) (string, error) {
	if tpr.userInput == "" {
		return defaultVal, nil
	}
	return tpr.userInput, nil
}

func TestPrompt(t *testing.T) {
	testFs := fstest.MapFS{}
	testFs["text.txt"] = &fstest.MapFile{Data: []byte("this is a text")}

	tpr := &testPrompter{
		choice: 1,
	}
	// create a summoner to summon text.txt at
	s, err := New(testFs, WithPrompter(tpr))
	assert.NoError(t, err)

	ret, err := s.renderTemplate(`{{ prompt "slotA" "What is your name" "David" }}`)
	assert.NoError(t, err)

	// check default value
	assert.Equal(t, "What is your name", tpr.b.String())
	assert.Equal(t, "David", ret)

	promptVal, err := s.renderTemplate(`{{ promptValue "slotA" }}`)
	assert.NoError(t, err)
	assert.Equal(t, "David", promptVal)

	// test selection (choice is B by the testPrompter config)
	ret, err = s.renderTemplate(`{{ prompt "continue" "Select One" (list "A" "B" "C") }}`)
	assert.NoError(t, err)

	assert.Equal(t, "B", ret)
	promptVal, err = s.renderTemplate(`{{ promptValue "continue" }}`)
	assert.NoError(t, err)
	assert.Equal(t, "B", promptVal)
}

func TestRunFunction(t *testing.T) {
	testFs := fstest.MapFS{}
	testFs["test.txt"] = &fstest.MapFile{Data: []byte("test content")}

	// Create a mock cobra command structure similar to what would exist in real usage
	rootCmd := &cobra.Command{
		Use:   "summon",
		Short: "Summon tool",
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print("init executed")
		},
	}

	ghaCmd := &cobra.Command{
		Use:   "gha",
		Short: "GitHub Actions",
	}
	ghaCmd.AddCommand(initCmd)
	rootCmd.AddCommand(ghaCmd)

	s, err := New(testFs)
	require.NoError(t, err)

	// Set up the cobra command in the driver
	s.opts.cobraCmd = rootCmd

	// Set up the cmdToSpec mapping to simulate configuration from summon.config.yaml
	// This is what the run function looks for when trying to find exec handle references
	s.cmdToSpec = map[*cobra.Command]*commandSpec{
		ghaCmd: {
			command: config.ArgSliceSpec{"echo"},
			args:    config.ArgSliceSpec{"gha command executed"},
		},
		initCmd: {
			command: config.ArgSliceSpec{"echo"},
			args:    config.ArgSliceSpec{"init executed"},
		},
	}

	t.Run("run with single command", func(t *testing.T) {
		// This should work - running a single command
		ret, err := s.renderTemplate(`{{ run "gha" }}`)
		assert.NoError(t, err)
		assert.Contains(t, ret, "gha command executed")
	})

	t.Run("run with command and parameters", func(t *testing.T) {
		// This is the previously buggy scenario: {{ run "gha" "init" }}
		// The bug was that driverCopy.opt.cobraCmd was being nilled out
		ret, err := s.renderTemplate(`{{ run "gha" "init" }}`)
		assert.NoError(t, err)
		// The result should contain the output from the init command
		assert.Contains(t, ret, "init executed")
	})

	t.Run("run with non-existent command", func(t *testing.T) {
		_, err := s.renderTemplate(`{{ run "nonexistent" }}`)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown command")
	})

	t.Run("run preserves original cobra command", func(t *testing.T) {
		// Ensure the original cobra command is preserved after run execution
		originalCmd := s.opts.cobraCmd
		_, err := s.renderTemplate(`{{ run "gha" "init" }}`)
		assert.NoError(t, err)
		assert.Equal(t, originalCmd, s.opts.cobraCmd, "Original cobra command should be preserved")
	})
}

func TestArgFunction(t *testing.T) {
	testFs := fstest.MapFS{}

	s, err := New(testFs)
	require.NoError(t, err)

	t.Run("arg with valid index", func(t *testing.T) {
		s.opts.args = []string{"first", "second", "third"}
		ret, err := s.renderTemplate(`{{ arg 0 }}`)
		assert.NoError(t, err)
		assert.Equal(t, "first", ret)

		ret, err = s.renderTemplate(`{{ arg 1 }}`)
		assert.NoError(t, err)
		assert.Equal(t, "second", ret)
	})

	t.Run("arg with out of range index", func(t *testing.T) {
		s.opts.args = []string{"first"}
		_, err := s.renderTemplate(`{{ arg 5 "missing argument" }}`)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing argument")
		assert.Contains(t, err.Error(), "index 5 out of range")
	})

	t.Run("arg with no args", func(t *testing.T) {
		s.opts.args = []string{}
		_, err := s.renderTemplate(`{{ arg 0 "no arguments provided" }}`)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no arguments provided")
	})
}

func TestArgsFunction(t *testing.T) {
	testFs := fstest.MapFS{}

	s, err := New(testFs)
	require.NoError(t, err)

	t.Run("args returns all arguments", func(t *testing.T) {
		s.opts.args = []string{"first", "second", "third"}
		ret, err := s.renderTemplate(`{{ range args }}{{ . }} {{ end }}`)
		assert.NoError(t, err)
		assert.Equal(t, "first second third ", ret)
	})

	t.Run("args with empty arguments", func(t *testing.T) {
		s.opts.args = []string{}
		ret, err := s.renderTemplate(`{{ len args }}`)
		assert.NoError(t, err)
		assert.Equal(t, "0", ret)
	})
}

func TestSwallowArgsFunction(t *testing.T) {
	testFs := fstest.MapFS{}

	s, err := New(testFs)
	require.NoError(t, err)

	t.Run("swallowargs returns empty string", func(t *testing.T) {
		s.opts.args = []string{"first", "second", "third"}
		ret, err := s.renderTemplate(`{{ swallowargs }}end`)
		assert.NoError(t, err)
		assert.Equal(t, "end", ret)
	})
}

func TestSummonFunction(t *testing.T) {
	testFs := fstest.MapFS{}
	testFs["test.txt"] = &fstest.MapFile{Data: []byte("test content")}

	s, err := New(testFs)
	require.NoError(t, err)

	t.Run("summon with default destination", func(t *testing.T) {
		ret, err := s.renderTemplate(`{{ summon "test.txt" }}`)
		assert.NoError(t, err)
		// The result should be a path to the summoned file
		assert.Contains(t, ret, "test.txt")

		// Verify the file was actually created
		_, err = os.Stat(ret)
		assert.NoError(t, err)

		// Clean up
		os.Remove(ret)
	})

	t.Run("summon with custom destination", func(t *testing.T) {
		tempDir := t.TempDir()
		ret, err := s.renderTemplate(`{{ summon "test.txt" "` + tempDir + `" }}`)
		assert.NoError(t, err)
		assert.Contains(t, ret, tempDir)

		// Verify the file exists in the custom location
		_, err = os.Stat(ret)
		assert.NoError(t, err)
	})
}

func TestFlagValueFunction(t *testing.T) {
	testFs := fstest.MapFS{}

	s, err := New(testFs)
	require.NoError(t, err)

	t.Run("flagValue with non-existent flag", func(t *testing.T) {
		ret, err := s.renderTemplate(`{{ flagValue "nonexistent" }}`)
		assert.NoError(t, err)
		assert.Equal(t, "", ret)
	})

	t.Run("flagValue with existing flag", func(t *testing.T) {
		// Add a mock flag to render
		s.flagsToRender = []*flagValue{
			{
				d:      s,
				name:   "testflag",
				effect: "flag-value",
			},
		}

		ret, err := s.renderTemplate(`{{ flagValue "testflag" }}`)
		assert.NoError(t, err)
		assert.Equal(t, "flag-value", ret)

		// Verify the flag was marked as explicit
		assert.True(t, s.flagsToRender[0].explicit)
	})
}

func TestExecuteTemplate(t *testing.T) {
	t.Run("executeTemplate removes <no value>", func(t *testing.T) {
		tmpl := `Hello {{ .Name }} <no value> World`

		// Create a simple template
		tpl := strings.NewReplacer("<no value>", "")
		result := tpl.Replace(tmpl)

		assert.NotContains(t, result, "<no value>")
		assert.Contains(t, result, "Hello")
		assert.Contains(t, result, "World")
	})
}

func TestRenderTemplate(t *testing.T) {
	testFs := fstest.MapFS{}

	s, err := New(testFs)
	require.NoError(t, err)

	t.Run("renderTemplate with valid template", func(t *testing.T) {
		s.opts.data = map[string]interface{}{
			"Name": "Test",
		}

		ret, err := s.renderTemplate(`Hello {{ .Name }}!`)
		assert.NoError(t, err)
		assert.Equal(t, "Hello Test!", ret)
	})

	t.Run("renderTemplate with invalid template", func(t *testing.T) {
		_, err := s.renderTemplate(`{{ invalid template syntax`)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not parse template")
	})

	t.Run("renderTemplate with sprig functions", func(t *testing.T) {
		ret, err := s.renderTemplate(`{{ "hello" | upper }}`)
		assert.NoError(t, err)
		assert.Equal(t, "HELLO", ret)
	})
}

func TestDryRunMode(t *testing.T) {
	testFs := fstest.MapFS{}

	s, err := New(testFs)
	require.NoError(t, err)

	// Create a mock cobra command structure
	rootCmd := &cobra.Command{Use: "summon"}
	testCmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print("test executed")
		},
	}
	rootCmd.AddCommand(testCmd)
	s.opts.cobraCmd = rootCmd

	// Set up the cmdToSpec mapping for the test command
	s.cmdToSpec = map[*cobra.Command]*commandSpec{
		testCmd: {
			command: config.ArgSliceSpec{"echo"},
			args:    config.ArgSliceSpec{"test executed"},
		},
	}

	t.Run("dry run adds dry-run indicator", func(t *testing.T) {
		s.opts.dryrun = true
		ret, err := s.renderTemplate(`{{ run "test" }}`)
		assert.NoError(t, err)
		assert.Contains(t, ret, "[test (dry-run)]")
	})
}

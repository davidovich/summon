package config

import (
	"testing"

	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigReader(t *testing.T) {
	config := dedent.Dedent(`
    version: 1
    exec:
      handles:
        hello: [python, -c, print("hello")]
        echo:
          cmd: [bash]
          flags:
            special-wrapper: 'happy new year: {{ .flag }}'
          help: 'this is an echo adapter'
          args: ['{{ flagValue "--special-wrapper" }}', '{{ arg 1 ""}}', '{{ arg 0 "" }}']

        with-flag:
          cmd: []
          flags:
            flag-desc:
              effect: "a"
              shorthand: "f"
              help: "flag-help"

    `)

	c := Config{}
	err := c.Unmarshal([]byte(config))

	require.Nil(t, err)
	args := c.Exec.ExecEnv["hello"].Value.(ArgSliceSpec)
	assert.Equal(t, "python", args[0])

	cmdSpec := c.Exec.ExecEnv["echo"].Value.(CmdDesc)
	assert.Equal(t, `{{ flagValue "--special-wrapper" }}`, cmdSpec.Args[0])

	assert.IsType(t, "", cmdSpec.Flags["special-wrapper"].Value)

	cmdSpecWithFlags := c.Exec.ExecEnv["with-flag"].Value.(CmdDesc)
	assert.IsType(t, FlagSpec{}, cmdSpecWithFlags.Flags["flag-desc"].Value)
}

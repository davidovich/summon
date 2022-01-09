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
      invokers:
        python -c:
          hello: [print("hello")]
        bash:
          echo:
            flags:
              special-wrapper: '{{ happy new year: . }}'
            help: 'this is an echo adapter'
            cmdArgs: ['{{ flag "--special-wrapper" }}', '{{ arg 1 ""}}', '{{ arg 0 "" }}']

          with-flag:
            flags:
              flag-desc:
                effect: "a"
                shorthand: "f"
                help: "flag-help"

    `)

	c := Config{}
	err := c.Unmarshal([]byte(config))

	require.Nil(t, err)
	args := c.Exec.Invokers["python -c"]["hello"].Value.(ArgSliceSpec)
	assert.Equal(t, "print(\"hello\")", args[0])

	cmdSpec := c.Exec.Invokers["bash"]["echo"].Value.(CmdSpec)
	assert.Equal(t, "{{ flag \"--special-wrapper\" }}", cmdSpec.CmdArgs[0])

	assert.IsType(t, "", cmdSpec.Flags["special-wrapper"].Value)

	cmdSpecWithFlags := c.Exec.Invokers["bash"]["with-flag"].Value.(CmdSpec)
	assert.IsType(t, FlagSpec{}, cmdSpecWithFlags.Flags["flag-desc"].Value)
}

package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestConfigWriter(t *testing.T) {
	c := Config{
		Version: 1,
		Executables: map[string]Executable{
			"go": {
				"gobin":  []interface{}{"github.com/myitcv/gobin"},
				"gohack": []interface{}{"github.com/rogppepe/gohack"},
			},
			"bash":      {"hello-bash": []interface{}{"hello.sh"}},
			"python -c": {"hello": []interface{}{"print(\"hello from python!\")"}},
		},
	}

	config, _ := yaml.Marshal(&c)

	assert.Equal(t, `
version: 1
aliases: {}
outputdir: ""
templates: ""
exec:
  bash:
    hello-bash:
    - hello.sh
  go:
    gobin:
    - github.com/myitcv/gobin
    gohack:
    - github.com/rogppepe/gohack
  python -c:
    hello:
    - print("hello from python!")
`, "\n"+string(config))
}

func TestConfigReader(t *testing.T) {
	config := dedent.Dedent(`
    version: 1
    exec:
      python -c:
        hello: [print("hello")]
	`)

	c := Config{}
	err := c.Unmarshal([]byte(config))

	require.Nil(t, err)
	assert.Equal(t, "print(\"hello\")", c.Executables["python -c"]["hello"][0])
}

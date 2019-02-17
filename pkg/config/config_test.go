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
			"go": Executable{
				"gobin":  "github.com/myitcv/gobin",
				"gohack": "github.com/rogppepe/gohack",
			},
			"bash":      Executable{"hello-bash": "hello.sh"},
			"python -c": Executable{"hello": "print(\"hello from python!\")"},
		},
	}

	config, _ := yaml.Marshal(&c)

	assert.Equal(t, dedent.Dedent(`
    version: 1
    aliases: {}
    outputdir: ""
    exec:
      bash:
        hello-bash: hello.sh
      go:
        gobin: github.com/myitcv/gobin
        gohack: github.com/rogppepe/gohack
      python -c:
        hello: print("hello from python!")
    `), "\n"+string(config))
}

func TestConfigReader(t *testing.T) {
	config := dedent.Dedent(`
    version: 1
    exec:
      python -c:
        hello: print("hello")
	`)

	var c Config
	err := yaml.Unmarshal([]byte(config), &c)

	require.Nil(t, err)
	assert.Equal(t, "print(\"hello\")", c.Executables["python -c"]["hello"])
}

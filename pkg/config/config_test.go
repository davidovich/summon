package config

import (
	"testing"

	"github.com/lithammer/dedent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// func TestConfigWriter(t *testing.T) {
// 	c := Config{
// 		Version: 1,
// 		Executables: map[string]Executable{
// 			"go": {
// 				"gobin":  []interface{}{"github.com/myitcv/gobin"},
// 				"gohack": []interface{}{"github.com/rogppepe/gohack"},
// 			},
// 			"bash":      {"hello-bash": []interface{}{"hello.sh"}},
// 			"python -c": {"hello": []interface{}{"print(\"hello from python!\")"}},
// 		},
// 	}

// 	config, _ := yaml.Marshal(&c)

// 	assert.Equal(t, `
// version: 1
// aliases: {}
// outputdir: ""
// templates: ""
// exec:
//   bash:
//     hello-bash:
//     - hello.sh
//   go:
//     gobin:
//     - github.com/myitcv/gobin
//     gohack:
//     - github.com/rogppepe/gohack
//   python -c:
//     hello:
//     - print("hello from python!")
// `, "\n"+string(config))
// }

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
	args := c.Executables["python -c"]["hello"].Value.(ArgSliceSpec)
	assert.Equal(t, "print(\"hello\")", args[0])
}

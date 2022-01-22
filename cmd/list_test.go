package cmd

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/davidovich/summon/pkg/summon"
	"github.com/stretchr/testify/assert"
)

func TestListCmd(t *testing.T) {

	tests := []struct {
		name string
		args []string
		// roostCmd  *cobra.Command
		expected string
	}{
		{
			name:     "no-args",
			expected: "a.txt\nb.txt\njson-for-template.json\nsummon.config.yaml",
		},
		{
			name: "--tree",
			args: []string{"--tree"},
			expected: `testdata
├── a.txt
├── b.txt
├── json-for-template.json
└── summon.config.yaml`,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i)+"_"+tt.name, func(t *testing.T) {
			s, _ := summon.New(cmdTestFS)

			cmd := newListCmd(false, nil, s)
			cmd.cmd.SetArgs(tt.args)

			b := &bytes.Buffer{}
			cmd.cmd.SetOut(b)
			cmd.cmd.Execute()

			assert.Contains(t, b.String(), tt.expected)
		})
	}
}

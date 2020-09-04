package cmd

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/davidovich/summon/pkg/summon"
	"github.com/gobuffalo/packr/v2"
	"github.com/stretchr/testify/assert"
)

func TestListCmd(t *testing.T) {

	box := packr.New("test box", "testdata/plain")
	tests := []struct {
		name string
		args []string
		// roostCmd  *cobra.Command
		expected string
	}{
		{
			name:     "no-args",
			expected: "json-for-template.json\nsummon.config.yaml",
		},
		{
			name: "--tree",
			args: []string{"--tree"},
			expected: `plain
├── json-for-template.json
└── summon.config.yaml`,
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i)+"_"+tt.name, func(t *testing.T) {
			s, _ := summon.New(box)

			cmd := newListCmd(s)
			cmd.SetArgs(tt.args)

			b := &bytes.Buffer{}
			cmd.SetOut(b)
			cmd.Execute()

			assert.Contains(t, b.String(), tt.expected)
		})
	}
}

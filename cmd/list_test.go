package cmd

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/davidovich/summon/pkg/summon"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestListCmd(t *testing.T) {

	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "no-args",
			args:     []string{"ls"},
			expected: "a.txt\nb.txt\njson-for-template.json\nsummon.config.yaml",
		},
		{
			name: "--tree",
			args: []string{"ls", "--tree"},
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

			rootCmd := &cobra.Command{Use: "root", Run: func(cmd *cobra.Command, args []string) {}}
			newListCmd(false, rootCmd, s, &mainCmd{})
			if tt.args == nil {
				tt.args = make([]string, 0)
			}
			rootCmd.SetArgs(tt.args)

			b := &bytes.Buffer{}
			rootCmd.SetOut(b)
			rootCmd.Execute()

			assert.Contains(t, b.String(), tt.expected)
		})
	}
}

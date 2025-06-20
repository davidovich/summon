package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompletionCommand(t *testing.T) {
	tt := []struct {
		name      string
		args      []string
		want      string
		assertion assert.ComparisonAssertionFunc
	}{
		{
			name:      "bash",
			args:      []string{"completion", "bash"},
			want:      "# bash completion V2 for summon",
			assertion: assert.Contains,
		},
		{
			name:      "zsh",
			args:      []string{"completion", "zsh"},
			want:      "compdef _summon summon",
			assertion: assert.Contains,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			_, rootCmd := makeRootCmd(false, tc.args...)

			b := &bytes.Buffer{}

			rootCmd.SetOut(b)
			err := rootCmd.Execute()

			assert.NoError(t, err)
			tc.assertion(t, b.String(), tc.want)
		})
	}
}

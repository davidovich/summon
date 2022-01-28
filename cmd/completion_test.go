package cmd

import (
	"bytes"
	"testing"

	"github.com/davidovich/summon/pkg/summon"
	"github.com/stretchr/testify/assert"
)

func TestCompletionCommand(t *testing.T) {
	s, _ := summon.New(cmdTestFS)

	cmd := newCompletionCmd(s)

	b := &bytes.Buffer{}

	cmd.SetOutput(b)
	cmd.Execute()

	assert.Contains(t, b.String(), "# bash completion V2 for completion")
}
